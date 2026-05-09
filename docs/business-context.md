# Appointment Confirmation System - Business Context

## Purpose

This system automatically sends appointment confirmation requests to customers and processes their responses.

The goal is to ensure that scheduled appointments are confirmed before the appointment date.

If the customer does not confirm attendance, the system updates the appointment status in Flowww.

## External Systems

### Flowww

Flowww is the business management system used to manage appointments.

The system interacts with Flowww through a dedicated Lambda (`flowww-operations`). Puppeteer is used only for authentication (login requires browser-based form submission). All data operations (center switching, appointment retrieval, confirmation, cancellation) are performed via direct HTTP calls to FLOWww's internal API, using session cookies extracted after login.

FLOWww has no public API — its internal request format was discovered by intercepting browser network traffic during manual operations.

Supported operations:

- Switch active center
- Retrieve upcoming appointments
- Confirm appointments
- Cancel appointments

### Messaging Providers

The system sends confirmation messages to customers.

Messaging is designed to be **channel-agnostic**.

Currently supported:

- WhatsApp Cloud API

Future channels may include:

- SMS
- Email

For this reason, message-related fields must remain **generic** and not tied to a specific provider.

## Core Business Flow

1. The system retrieves upcoming appointments from Flowww.
2. Each appointment is stored in DynamoDB if it does not already exist.
3. A confirmation message is sent to the customer.
4. The messaging provider returns a `message_id`.
5. The appointment is updated with message metadata.
6. The customer replies to the message.
7. A webhook processes the response.
8. The system updates the appointment status in Flowww.

## Idempotency

The system must be safe against retries.

Appointments are inserted using conditional writes.

The `MarkSent` operation ensures that only appointments in `PENDING` status can transition to `SENT`.

## Appointment Lifecycle

Possible statuses:

- `pending` → appointment stored but message not sent
- `sent` → confirmation message sent
- `confirmed` → customer confirmed attendance
- `cancelled` → appointment cancelled or declined

## Expiration

Appointments contain a `confirmation_expiration_timestamp`.

If the customer does not respond before this timestamp, the appointment will be automatically processed.

## Data Retention

Appointments may remain in DynamoDB for a period of time to allow debugging and metrics.

Expiration logic is separate from database cleanup.

## Key Design Principles

- Messaging must remain channel-agnostic
- Appointment operations must be idempotent
- DynamoDB writes must use conditional expressions
- External systems must be accessed through dedicated clients