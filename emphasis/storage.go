package emphasis

import (
	"context"
)

type Storage interface {
	FindEmphasis(ctx context.Context, id string) (*FitResult, error)
	SaveEmphasisInput(ctx context.Context, input Input) error
}
