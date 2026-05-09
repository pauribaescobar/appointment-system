# System Architecture

## Components

### appointment-confirmation-sender (Lambda)

Responsible for:

- retrieving appointments from Flowww
- storing appointments in DynamoDB
- sending confirmation messages

### WhatsApp Webhook (Lambda)

Responsible for:

- receiving customer replies
- mapping replies to appointments
- updating appointment status in Flowww

### Flowww Operations (Lambda)

Responsible for interacting with Flowww. Uses Puppeteer only for authentication (login requires browser-based form submission and dialog handling). All data operations are performed via direct HTTP calls to FLOWww's internal `dllrequest.asp` endpoint using session cookies extracted after login.

Supported operations:

- switch active center (`SEC_CONFIG_CLISEL_FORM.V4` + `SEC_REFRESH_FORM`)
- retrieve appointments (`SEC_DIARY_DAY_LOAD_FORM.V2`)
- confirm appointment (`SEC_APP_TAG_SAVE_FORM`)
- cancel appointment (`SEC_APP_NOT_ATTENDED_FORM`)

Browser setup:
- **Lambda**: `puppeteer-core` + `@sparticuz/chromium` (Linux Chromium binary)
- **Local**: `puppeteer-core` + `puppeteer`'s bundled Chrome with `@sparticuz/chromium` args (same driver and config, macOS-compatible binary)

## Database

### DynamoDB Table: appointments

Primary key:

- `id`

Fields stored:

- `center_id`
- `customer_id`
- `customer_name`
- `customer_phone_number`
- `date`
- `start_time`
- `end_time`
- `confirmation_expiration_timestamp`
- `status`
- `message_sent_at`
- `message_responded_at`

## Messaging Providers

The system interacts with messaging providers through a messaging client.

Current implementation:

- WhatsApp Cloud API

The messaging client must remain provider-agnostic.