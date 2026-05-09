package flowww

import (
	"appointment-system/internal/domain"
	"context"
)

type FlowwwSystem struct {
	client *FlowwwClient
	mapper *FlowwwMapper
}

func NewFlowwwSystem(client *FlowwwClient, mapper *FlowwwMapper) *FlowwwSystem {
	return &FlowwwSystem{
		client: client,
		mapper: mapper,
	}
}

func (s *FlowwwSystem) ListAppointments(
	ctx context.Context,
	centerId string,
) ([]domain.Appointment, error) {
	resp, err := s.client.ListAppointments(
		ctx,
		centerId,
	)

	if err != nil {
		return nil, err
	}

	apmnts, err := s.mapper.MapFlowwwAppointmentDTOToAppointment(*resp, centerId)
	if err != nil {
		return nil, err
	}

	return apmnts, nil
}

func (s *FlowwwSystem) ConfirmAppointment(
	ctx context.Context,
	centerId string,
	appointmentId string,
	appointmentDate string,
) error {
	rawId := s.mapper.ExtractFlowwwAppointmentId(appointmentId)
	return s.client.UpdateConfirmation(ctx, true, centerId, rawId, appointmentDate)
}

func (s *FlowwwSystem) CancelAppointment(
	ctx context.Context,
	centerId string,
	appointmentId string,
	appointmentDate string,
) error {
	rawId := s.mapper.ExtractFlowwwAppointmentId(appointmentId)
	return s.client.UpdateConfirmation(ctx, false, centerId, rawId, appointmentDate)
}
