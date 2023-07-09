package response

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"net/http"
	"strings"
)

type Response struct {
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

const (
	StatusOK = 200
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string, status int) Response {
	if status == 0 {
		status = http.StatusInternalServerError
	}

	return Response{
		Status: status,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required", err.Field()))
		case "url":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s must be a valid url", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", err.Field()))
		}
	}

	return Response{
		Status: http.StatusBadRequest,
		Error:  strings.Join(errMsgs, ", "),
	}
}
