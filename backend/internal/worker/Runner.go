package worker

import "context"

type Runner interface{
	RunOnce(ctx context.Context) error
	Name() string
}