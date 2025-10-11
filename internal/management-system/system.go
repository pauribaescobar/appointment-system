package systems

import "appointment-system/internal/appointments"

type AppointmentSystem interface {
	GetAvailableAppointments(centerId string, daysAhead int) ([]models.Appointment, error)
	ConfirmAppoinment(appointmentId string) error
	CancelAppointment(appointmentId string) error
}
