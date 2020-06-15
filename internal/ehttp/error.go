package ehttp

import (
	"fmt"
	"net/http"
)

type HTTPError struct {
	Msg        string `json:"error,omitempty"`
	StatusCode int    `json:"-"`
	Detail     string `json:"-"`
}

func (h HTTPError) Error() string {
	return fmt.Sprintf("message: %v; status code: %v; detail: %v", h.Msg, h.StatusCode, h.Detail)
}

func New(msg string, status int, detail string) error {
	return HTTPError{
		Msg:        msg,
		StatusCode: status,
		Detail:     detail,
	}
}

func IncorrectID(id int64) error {
	msg := fmt.Sprintf("incorrect id: %v", id)
	return HTTPError{
		Msg:        msg,
		StatusCode: http.StatusBadRequest,
		Detail:     msg,
	}
}

func JSONUnmarshalErr(err error) error {
	return HTTPError{
		Msg:        "",
		StatusCode: http.StatusBadRequest,
		Detail:     fmt.Sprintf("can't unmarshal input json: %v", err),
	}
}

func InternalServerErr(detail string) error {
	return HTTPError{
		Msg:        "",
		StatusCode: http.StatusInternalServerError,
		Detail:     detail,
	}
}

func NotFoundErr(msg string, detail string) error {
	return HTTPError{
		Msg:        msg,
		StatusCode: http.StatusNotFound,
		Detail:     detail,
	}
}