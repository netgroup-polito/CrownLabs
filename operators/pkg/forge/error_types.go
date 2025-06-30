package forge

import "net/http"

type HTTPError struct {
	Status  int
	Message string
}

func (e *HTTPError) Error() string {
	return e.Message
}

func NewInternalError(err error) *HTTPError {
	return &HTTPError{
		Status:  http.StatusInternalServerError,
		Message: err.Error(),
	}
}
