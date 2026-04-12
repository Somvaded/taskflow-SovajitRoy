package domain_error

import "fmt"

type ErrorOption func(*AppError)

func WithExtraData(data map[string]any) ErrorOption {
	return func(ae *AppError) {
		ae.ExtraData = data
	}
}

// Raise constructs an AppError. msg overrides the default message from msgMap.
// Pass an empty string to use the default message.
func Raise(code ErrorCode, msg string, cause error, opts ...ErrorOption) error {
	ae := &AppError{
		ErrorCode: code,
		Cause:     cause,
	}

	if msg != "" {
		ae.Message = msg + ": " + ae.GetMsg()
	} else {
		ae.Message = ae.GetMsg()
	}

	for _, opt := range opts {
		opt(ae)
	}

	if cause == nil {
		ae.Cause = fmt.Errorf("%s: %s", ae.GetCode(), ae.GetMsg()) //nolint:err113
	}

	return ae
}
