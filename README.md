# appointment-system

A system that automates appointment handling via WhatsApp for several appointment management systems.

## Architecture

The system consists of two AWS Lambda functions:

- **appointment-confirmation-sender** -- Triggered by EventBridge on a schedule. Fetches upcoming appointments from the management system (Flowww), persists them in DynamoDB, and sends confirmation messages via WhatsApp with interactive buttons (Confirm / Cancel).
- **whatsapp-webhook** -- Receives customer replies via an API Gateway HTTP API. When a customer taps a button, the webhook processes the response: confirms or cancels the appointment on Flowww, sends an acknowledgment message, and deletes the appointment from DynamoDB.

## Prerequisites

- Go 1.24+
- AWS CLI configured with a `personal` profile
- An AWS account with Lambda, DynamoDB, API Gateway, and EventBridge access
- A Meta Business account with WhatsApp Business API and approved message templates

## Environment Variables

Create a `.env.lambda` file in the project root with the following variables:

```
DYNAMODB_APPOINTMENTS_TABLE=appointments
DYNAMODB_CENTERS_TABLE=centers
ENV=prod
PHONE_HASH_SALT=your-salt
LOG_LEVEL=info
WHATSAPP_AUTH_TOKEN=your-whatsapp-auth-token
WEBHOOK_VERIFY_TOKEN=your-webhook-verify-token
```

> `AWS_REGION` is set automatically by Lambda -- do not include it here.

## Makefile Targets

### Confirmation Sender

| Target | Description |
|--------|-------------|
| `make build-confirmation-sender` | Builds the binary (linux/arm64) and creates a zip |
| `make deploy-confirmation-sender` | Builds and uploads the code to the Lambda |
| `make configure-confirmation-sender` | Sets environment variables on the Lambda |
| `make deploy-confirmation-sender-full` | Configures env vars + deploys code (full deploy) |

### Webhook

| Target | Description |
|--------|-------------|
| `make build-webhook` | Builds the binary (linux/arm64) and creates a zip |
| `make deploy-webhook` | Builds and uploads the code to the Lambda |
| `make configure-webhook` | Sets environment variables on the Lambda |
| `make deploy-webhook-full` | Configures env vars + deploys code (full deploy) |
| `make create-webhook` | Creates the Lambda function in AWS (one-time, idempotent) |

### One-time Infrastructure Setup

| Target | Description |
|--------|-------------|
| `make setup-api-gateway` | Creates the HTTP API Gateway with GET/POST routes, Lambda integration, and auto-deploy stage |
| `make setup-message-id-index` | Creates the `message_id-index` GSI on the `appointments` DynamoDB table |
| `make setup-iam-permissions` | Applies IAM policy to the Lambda role (DynamoDB access + Lambda invoke) |
| `make setup-lambda-timeouts` | Sets the timeout (default 600s) on all Lambda functions |

## First-time Setup

```bash
# 1. Configure your environment variables in .env.lambda

# 2. Create the webhook Lambda function
make create-webhook

# 3. Configure env vars on both Lambdas
make configure-confirmation-sender
make configure-webhook

# 4. Deploy code to both Lambdas
make deploy-confirmation-sender
make deploy-webhook

# 5. Create the API Gateway (prints the Webhook URL at the end)
make setup-api-gateway

# 6. Create the DynamoDB GSI for message_id lookups
make setup-message-id-index

# 7. Apply IAM permissions to the Lambda role
make setup-iam-permissions

# 8. Set Lambda timeouts
make setup-lambda-timeouts

# 9. Configure the webhook in Meta (see below)
```

## Subsequent Deploys

```bash
# Deploy confirmation sender (env vars + code)
make deploy-confirmation-sender-full

# Deploy webhook (env vars + code)
make deploy-webhook-full
```

## Meta Webhook Configuration

After deploying the webhook Lambda and running `make setup-api-gateway`, you need to register the webhook URL with Meta so it sends incoming WhatsApp messages to your Lambda.

1. Copy the **Webhook URL** printed at the end of `make setup-api-gateway` (e.g., `https://<api-id>.execute-api.eu-west-3.amazonaws.com/webhook`)
2. Go to [Meta Developer Dashboard](https://developers.facebook.com) > Your App > **WhatsApp** > **Configuration**
3. Under **Webhook**, click **Edit**
4. Set **Callback URL** to the webhook URL and **Verify token** to your `WEBHOOK_VERIFY_TOKEN` value from `.env.lambda`
5. Click **Verify and save** -- Meta sends a GET request to verify your endpoint
6. After verification, click **Manage** and subscribe to the **messages** field

> The verify token in Meta must match the `WEBHOOK_VERIFY_TOKEN` environment variable configured on the Lambda.

## WhatsApp Message Templates

Three templates need to be created in Meta Business Manager (WhatsApp > Message Templates):

### `appointment_confirmation_v1` (sent by confirmation sender)

- **Category:** Utility
- **Body:** Appointment details with center name, date, start/end time, and address
- **Buttons:** Quick Reply -- "CONFIRMO" and "NO PUEDO ASISTIR"

### `appointment_confirmed_v1` (sent by webhook on confirm)

- **Category:** Utility
- **Header:** CITA CONFIRMADA
- **Body:** Confirmation acknowledgment with customer name, center, date, times, and address

### `appointment_cancelled_v1` (sent by webhook on cancel)

- **Category:** Utility
- **Header:** CITA CANCELADA
- **Body:** Cancellation acknowledgment with customer name, center, and date

## Technical Challenges

### Reverse-Engineering FLOWww's Session-Based API

FLOWww is a legacy ASP-based business management system with no public API. All interactions are performed through a browser-based SPA built with Stencil.js web components. This system automates FLOWww by reverse-engineering its internal HTTP requests.

#### The Problem: Center Switching in AWS Lambda

The system manages appointments across multiple clinic centers. To retrieve appointments for a specific center, the server-side session must first be loaded with that center — there is no way to pass a center ID directly to the appointments endpoint.

The original implementation used Puppeteer to click through a shadow DOM dropdown (`flw-clinic-item` Stencil components) to switch centers. This worked locally but **failed inconsistently in AWS Lambda** due to differences between the local Chromium binary and `@sparticuz/chromium`'s headless shell mode (Chromium ~131 vs ~146 locally).

**Symptoms observed in Lambda:**
- The dropdown's `collapsed` attribute never toggled after clicking
- All non-selected clinic items remained invisible (`offsetParent === null`, bounding rect all zeros)
- 4 different click strategies failed: synthetic `.click()`, `page.mouse.click()` with real coordinates, `dispatchEvent` with `composed: true`, and direct `host.click()`
- When the dropdown did render (~30% of invocations), items appeared as duplicates in a separate container

#### The Discovery Process

By adding a `page.on('request')` interceptor during a successful local center change, we captured the exact network requests FLOWww makes internally:

```
POST dllrequest.asp  →  frm=SEC_CONFIG_CLISEL_FORM.V4&...&SEC_CONFIG_CLISEL_FORM.V4_1={centerId}
POST dllrequest.asp  →  frm=SEC_REFRESH_FORM&SEC_REFRESH_FORM_Fields=0
```

#### The Solution

Replaced the entire UI-based center switching (~80 lines of Puppeteer interaction) with two direct API calls:

1. **`SEC_CONFIG_CLISEL_FORM.V4`** — tells the server to switch the session to the target center
2. **`SEC_REFRESH_FORM`** — commits the center change on the server-side session

Both calls alone return HTTP 200, but the center change only takes effect when **both** are called in sequence. This is 100% reliable across environments and significantly faster (~2 HTTP calls vs ~30 seconds of UI interaction).

#### Extending the Pattern to All Operations

After the center switching breakthrough, the same reverse-engineering approach was applied to all remaining FLOWww operations:

| Operation | Old approach (Puppeteer UI) | New approach (direct API) |
|---|---|---|
| **Switch center** | Click shadow DOM dropdown, select clinic item | `SEC_CONFIG_CLISEL_FORM.V4` + `SEC_REFRESH_FORM` |
| **Get appointments** | Already used direct API | `SEC_DIARY_DAY_LOAD_FORM.V2` |
| **Confirm appointment** | Navigate to agenda, find date, double-click appointment, click tags button, find green checkbox, click it | `SEC_APP_TAG_SAVE_FORM` with `{DiaryGID, TagID, TagChecked}` |
| **Cancel appointment** | Navigate to agenda, find date, double-click appointment, click delete button, accept dialog with reason | `SEC_APP_NOT_ATTENDED_FORM` with appointment ID and cancellation reason |

This eliminated all UI navigation code (agenda module navigation, date pagination, appointment loading via double-click, dialog handling) and reduced the entire `update_confirmation_status.js` from ~180 lines of Puppeteer automation to ~58 lines of direct API calls.

A key discovery was that confirmation and cancellation operations don't require navigating to the appointment's date or loading it first — the API calls work directly with the appointment ID, regardless of which page the browser session is on.

#### Key Insights

- **FLOWww is session-stateful**: operations like listing appointments depend on server-side session state (which center is loaded). You must "set up" the session with prerequisite requests before making data queries.
- **Not all operations require session state**: confirmation and cancellation work with just the appointment ID — no center loading or page navigation needed. This was discovered empirically.
- **The browser is still required for authentication**: the login flow involves form submission, navigation, and dialog handling. Once logged in, cookies are extracted and used for all subsequent direct API calls.
- **`@sparticuz/chromium` ships a Linux-only binary**: it cannot run on macOS (causes `ENOEXEC` / system error -8). For local development, we use `puppeteer`'s bundled Chrome with `@sparticuz/chromium`'s args/viewport/headless settings to match Lambda's configuration as closely as possible.

## Local Development

Use the VS Code launch configurations in `.vscode/launch.json`:

- **Debug Confirmation Sender** -- Runs the confirmation sender locally with env vars from `.env`
- **Debug Webhook** -- Runs the webhook locally with env vars from `.env`

Both configurations require a `.env` file with local development values (including `AWS_PROFILE=personal` and `ENV=local`).
