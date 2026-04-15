package processor

import "errors"

var (
	ErrProcessorAlreadyStarted = errors.New("processor already started")
	ErrProcessorStopped        = errors.New("processor stopped")
)
