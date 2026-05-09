package appointments

import (
	"appointment-system/internal/domain"
	"context"
)

type AppointmentRepository interface {
	PutIfNotExists(ctx context.Context, appointment *domain.Appointment) (*domain.Appointment, error)
	GetByID(ctx context.Context, id string) (*domain.Appointment, error)
	GetByMessageID(ctx context.Context, messageID string) (*domain.Appointment, error)
	MarkSent(ctx context.Context, id, messageId, messageChannel string) error
	Delete(ctx context.Context, id string) error
	QueryPendingAppointments(
		ctx context.Context,
	) (*[]domain.Appointment, error)
}
