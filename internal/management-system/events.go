package managementsystem

type UpdateConfirmationRequest struct {
	Operation          string `json:"operation"`
	CenterId           string `json:"centerId"`
	ConfirmationStatus bool   `json:"confirmationStatus"`
	AppointmentDate    string `json:"appointmentDate"`
	AppointmentId      string `json:"appointmentId"`
}

type ListAppointmentsRequest struct {
	Operation string `json:"operation"`
	CenterId  string `json:"centerId"`
}
