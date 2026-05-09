package domain

type AppointmentStatus string

const (
	AppointmentStatusPending   AppointmentStatus = "pending"
	AppoinmentStatusSent       AppointmentStatus = "sent"
	AppointmentStatusConfirmed AppointmentStatus = "confirmed"
	AppointmentStatusCancelled AppointmentStatus = "cancelled"
)

type Appointment struct {
	ID                              string            `json:"id" dynamodbav:"id"`
	CenterID                        string            `json:"center_id" dynamodbav:"center_id"`
	CustomerId                      string            `json:"customer_id" dynamodbav:"customer_id"`
	CustomerName                    string            `json:"customer_name" dynamodbav:"customer_name"`
	CustomerPhoneNumber             string            `json:"customer_phone_number" dynamodbav:"customer_phone_number"`
	Date                            string            `json:"date" dynamodbav:"date"`
	StartTime                       string            `json:"start_time" dynamodbav:"start_time"`
	EndTime                         string            `json:"end_time" dynamodbav:"end_time"`
	ConfirmationExpirationTimestamp int64             `json:"confirmation_expiration_timestamp" dynamodbav:"confirmation_expiration_timestamp"`
	Status                          AppointmentStatus `json:"status" dynamodbav:"status"`
	MessageSentAt                   int64             `json:"message_sent_at" dynamodbav:"message_sent_at"`
	MessageRespondedAt              string            `json:"message_responded_at" dynamodbav:"message_responded_at"`
}
