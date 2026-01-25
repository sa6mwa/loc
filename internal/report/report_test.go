package report

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestWriteJSONHelpResponse(t *testing.T) {
	payload := HelpResponse{
		Usage:       "loc [flags] [extensions...]",
		Description: "test help",
		Examples:    []string{"loc"},
		Extensions:  []string{".go", ".ts"},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, payload); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var out HelpResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Usage != payload.Usage || out.Description != payload.Description {
		t.Fatalf("unexpected help output: %+v", out)
	}
}

func TestWriteJSONErrorResponse(t *testing.T) {
	payload := ErrorResponse{
		Error: ErrorDetail{
			Message: "loc failed",
			Detail:  "boom",
		},
	}

	var buf bytes.Buffer
	if err := WriteJSON(&buf, payload); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var out ErrorResponse
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Error.Message != payload.Error.Message || out.Error.Detail != payload.Error.Detail {
		t.Fatalf("unexpected error output: %+v", out)
	}
}
