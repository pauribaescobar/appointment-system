package domain

import "context"

type AppointmentSystemFactory interface {
	Build(
		systemName string,
	) AppointmentSystem
}

type AppointmentSystem interface {
	ListAppointments(
		ctx context.Context,
		centerId string,
	) ([]Appointment, error)
	ConfirmAppointment(
		ctx context.Context,
		centerId string,
		appointmentId string,
		appointmentDate string,
	) error
	CancelAppointment(
		ctx context.Context,
		centerId string,
		appointmentId string,
		appointmentDate string,
	) error
}
