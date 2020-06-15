package render

import (
	"encoding/json"
	"net/http"
)

const (
	ContentTypeHeader = "Content-Type"
	JSONContentType   = "application/json"
)

type ErrorMsg struct {
	Detail string `json:"error"`
	Status int    `json:"-"`
}

func (e ErrorMsg) Error() string {
	return e.Detail
}

func HTTPError(msg string, status int, w http.ResponseWriter) {
	Error(ErrorMsg{msg, status}, w)
}

func Error(msg ErrorMsg, w http.ResponseWriter) {
	w.WriteHeader(msg.Status)

	if msg.Detail != "" {
		w.Header().Add(ContentTypeHeader, JSONContentType)

		out, err := json.Marshal(msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// no need to handle error here
		_, _ = w.Write(out)
	}
}
