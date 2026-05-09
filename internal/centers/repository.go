package centers

import (
	"appointment-system/internal/domain"
	"context"
)

type CentersRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Center, error)
}
