package glue

import (
	"fmt"
	"net/http"
)

type RespError struct {
	ErrMessage string        `json:"message"`
	ErrStatus  int           `json:"status"`
	ErrError   string        `json:"error"`
	ErrCauses  []interface{} `json:"causes"`
}

func (e RespError) Error() string {
	return fmt.Sprintf("message: %s - status: %d - error: %s - causes: %v",
		e.ErrMessage, e.ErrStatus, e.ErrError, e.ErrCauses)
}

func (e RespError) Message() string {
	return e.ErrMessage
}

func (e RespError) Status() int {
	return e.ErrStatus
}

func (e RespError) Causes() []interface{} {
	return e.ErrCauses
}

// NewNotFoundError ...
func NewNotFoundError(msg string) RespError {
	return RespError{
		ErrMessage: msg,
		ErrStatus:  http.StatusNotFound,
		ErrError:   "not found",
	}
}

// NewBadRequestError ...
func NewBadRequestError(msg string) RespError {
	return RespError{
		ErrMessage: msg,
		ErrStatus:  http.StatusBadRequest,
		ErrError:   "bad request",
	}
}

// NewInternalServerError ...
func NewInternalServerError(msg string, err error) RespError {
	result := RespError{
		ErrMessage: msg,
		ErrStatus:  http.StatusInternalServerError,
		ErrError:   "internal_server_error",
	}
	if err != nil {
		result.ErrCauses = append(result.ErrCauses, err.Error())
	}
	return result
}
