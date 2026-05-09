# Business Flows

## Sending Confirmation Messages

1. Retrieve appointments from Flowww.
2. Store the appointment in DynamoDB using a conditional write.
3. If the appointment already exists, skip insertion.
4. Send confirmation message to the customer.
5. Update appointment status to `sent`.
6. Store `message_sent_at`.

## Customer Confirmation

1. Customer replies to the message.
2. Webhook receives the response.
3. The system identifies the appointment.
4. `message_responded_at` is recorded.
5. The appointment status is updated.
6. Flowww is updated accordingly.

## Appointment Expiration

1. If the customer does not respond before `confirmation_expiration_timestamp`, the appointment is processed automatically.
2. The appointment status is updated in Flowww.

## Cancellation Flow

1. Customer declines the appointment.
2. Webhook processes the decline.
3. Flowww appointment is cancelled.
4. Appointment status becomes `cancelled`.