package appointments

type AppointmentConfirmationSenderEvent struct {
	ManagementSystem string `json:"managementSystem"`
	MessagingClient  string `json:"messagingClient"`
	CenterId         string `json:"centerId"`
}
