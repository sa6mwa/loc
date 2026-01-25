package report

import (
	"encoding/json"
	"io"

	"pkt.systems/prettyx"
)

// ErrorResponse describes an error payload.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail provides error information.
type ErrorDetail struct {
	Message string `json:"message"`
	Detail  string `json:"detail,omitempty"`
}

// HelpResponse describes the help payload.
type HelpResponse struct {
	Usage       string   `json:"usage"`
	Description string   `json:"description"`
	Examples    []string `json:"examples"`
	Extensions  []string `json:"extensions"`
}

// WriteJSON pretty-prints JSON with prettyx.
func WriteJSON(w io.Writer, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return prettyx.PrettyTo(w, data, nil)
}
