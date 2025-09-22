package resource

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/cccteam/ccc/accesstypes"
	"github.com/cccteam/httpio"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/errors/v5"
)

// TestOperationsValid tests the Operations function with a valid JSON patch array.
func TestOperationsValid(t *testing.T) {
	jsonStr := ` + "`" + `[
  {
    "op": "add",
    "path": "/X",
    "value": { "description": "Office X" }
  },
  {
    "op": "patch",
    "path": "/O",
    "value": { "description": "Office O 2" }
  },
  {
    "op": "remove",
    "path": "/W"
  }
]` + "`" + `
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonStr))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Valid pattern starting with '/'
	pattern := "/{id}/{id2}"

	opCount := 0
	expected := []struct {
		opType    string
		httpMethod string
	}{
		{"add", http.MethodPost},
		{"patch", http.MethodPatch},
		{"remove", http.MethodDelete},
	}

	ops := Operations(req, pattern, RequireCreatePath())
	ops(func(op *Operation, opErr error) bool {
		if opErr != nil {
			t.Fatalf("unexpected error: %v", opErr)
		}

		if opCount >= len(expected) {
			t.Fatalf("received more operations than expected")
		}

		exp := expected[opCount]
		if string(op.Type) != exp.opType {
			t.Errorf("operation %d: expected type %s, got %s", opCount, exp.opType, op.Type)
		}

		if op.Req.Method != exp.httpMethod {
			t.Errorf("operation %d: expected HTTP method %s, got %s", opCount, exp.httpMethod, op.Req.Method)
		}

		opCount++
		return true
	})

	if opCount != len(expected) {
		t.Fatalf("expected %d operations, but got %d", len(expected), opCount)
	}
}

// TestOperationsInvalidPattern tests that Operations returns an error when the pattern does not start with a slash.
func TestOperationsInvalidPattern(t *testing.T) {
	jsonStr := ` + "`" + `[
  { "op": "add", "path": "/X", "value": { "description": "Office X" } }
]` + "`" + `
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonStr))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	// Invalid pattern (doesn't start with "/")
	pattern := "invalid-pattern"

	errReceived := ""

	ops := Operations(req, pattern, RequireCreatePath())
	ops(func(op *Operation, opErr error) bool {
		if opErr != nil {
			errReceived = opErr.Error()
			return false
		}
		return true
	})

	if errReceived == "" {
		t.Fatal("expected error due to invalid pattern, but got none")
	}
}

// TestOperationsInvalidJSON tests that Operations returns an error when provided invalid JSON.
func TestOperationsInvalidJSON(t *testing.T) {
	invalidJSON := "not a json array"
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(invalidJSON))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	pattern := "/{id}/{id2}"
	errReceived := ""

	ops := Operations(req, pattern, RequireCreatePath())
	ops(func(op *Operation, opErr error) bool {
		if opErr != nil {
			errReceived = opErr.Error()
			return false
		}
		return true
	})

	if errReceived == "" {
		t.Fatal("expected error due to invalid JSON, but got none")
	}
}

// TestOperationsCreatePathEmpty tests that Operations returns an error when the create path is empty and RequireCreatePath is enabled.
func TestOperationsCreatePathEmpty(t *testing.T) {
	jsonStr := ` + "`" + `[
  { "op": "add", "path": "/", "value": { "description": "Office X" } }
]` + "`" + `
	req, err := http.NewRequest("POST", "http://example.com", bytes.NewBufferString(jsonStr))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	pattern := "/{id}/{id2}"
	errReceived := ""

	ops := Operations(req, pattern, RequireCreatePath())
	ops(func(op *Operation, opErr error) bool {
		if opErr != nil {
			errReceived = opErr.Error()
			return false
		}
		return true
	})

	if errReceived == "" {
		t.Fatal("expected error due to empty create path, but got none")
	}
}
