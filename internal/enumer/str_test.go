package enumer

import (
	"testing"

	"github.com/loopopen/shoot/internal/transfer"
)

func TestEnumStringCase(t *testing.T) {
	tests := map[transfer.Case]string{
		"":                  "HTTPServerError",
		transfer.CasePascal: "HTTPServerError",
		transfer.CaseCamel:  "httpServerError",
		transfer.CaseSnake:  "http_server_error",
		transfer.CaseLower:  "httpservererror",
		transfer.CaseUpper:  "HTTPSERVERERROR",
	}
	for enumCase, want := range tests {
		t.Run(enumCase.String(), func(t *testing.T) {
			got := enumString("HTTPStatus", "HTTPStatusHTTPServerError", enumCase)
			if got != want {
				t.Fatalf("enumString() = %q, want %q", got, want)
			}
		})
	}
}

func TestEnumStringWithoutTypePrefix(t *testing.T) {
	if got := enumString("HTTPStatus", "Unknown", ""); got != "Unknown" {
		t.Fatalf("enumString() = %q, want %q", got, "Unknown")
	}
}
