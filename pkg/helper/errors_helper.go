package helper

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ErrorResponse struct {
	RequestID string              `json:"request_id,omitempty"`
	Errors    map[string][]string `json:"errors"`
}

// SingleError creates an error response with a single error message
// Used for client-side errors
func SingleError(key string, message string) *ErrorResponse {
	return &ErrorResponse{
		Errors: map[string][]string{
			key: {message},
		},
	}
}

// ServerError creates an error response with request ID for tracing
// Used for server-side errors
func ServerError(log *logrus.Logger, message string) *ErrorResponse {
	e := &ErrorResponse{
		RequestID: uuid.NewString(),
		Errors: map[string][]string{
			"server_error": {message},
		},
	}

	log.Error(e.Error())
	return e
}

func (e *ErrorResponse) Error() string {
	if len(e.Errors) == 0 {
		return "Unknown error"
	}

	if e.RequestID != "" {
		for _, msgs := range e.Errors {
			if len(msgs) > 0 {
				return "Error " + e.RequestID + ": " + msgs[0]
			}
		}
	}

	for _, msgs := range e.Errors {
		if len(msgs) > 0 {
			return msgs[0]
		}
	}

	return "Unknown error"
}

func ClientError(message string) *ErrorResponse {
	return &ErrorResponse{
		Errors: map[string][]string{
			"client_error": {message},
		},
	}
}
