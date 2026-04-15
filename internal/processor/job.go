package processor

import (
	"context"

	"backend-path/internal/domain"
)

type Job struct {
	ID          string
	Transaction *domain.Transaction
	Ctx         context.Context
	RetryCount  int
	MaxRetries  int
}
