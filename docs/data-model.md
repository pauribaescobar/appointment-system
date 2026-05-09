# Data Model

## Appointment

Structure used by the system.

Fields:

- id
- center_id
- customer_id
- customer_name
- customer_phone_number
- date
- start_time
- end_time
- confirmation_expiration_timestamp
- status
- message_sent_at
- message_responded_at

## Appointment Status

Possible values:

- pending
- sent
- confirmed
- cancelled

## Center

Stores metadata about each business location.

Fields:

- id
- name
- address
- phone_number