package transfer_test

import (
	"testing"

	"github.com/lopolopen/shoot/internal/transfer"
)

func TestFirstLowerLetter(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "lowercase",
			str:  "lowercase",
			want: "l",
		},
		{
			name: "Uppercase",
			str:  "Uppercase",
			want: "u",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transfer.FirstLowerLetter(tt.str)
			if got != tt.want {
				t.Errorf("FirstLowerLetter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "camel case",
			str:  "camelCase",
			want: "camelCase",
		},
		{
			name: "pascal case",
			str:  "PascalCase",
			want: "pascalCase",
		},
		{
			name: "pascal case with acronyms",
			str:  "ACRONYMSPascalCaseACRONYMS",
			want: "acronymsPascalCaseAcronyms",
		},
		{
			name: "all caps",
			str:  "ALLCAPS",
			want: "allcaps",
		},
		{
			name: "snake case",
			str:  "snake_case",
			want: "snakeCase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transfer.ToCamelCase(tt.str)
			if got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToCamelCaseGO(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "camel case",
			str:  "camelCase",
			want: "camelCase",
		},
		{
			name: "pascal case",
			str:  "PascalCase",
			want: "pascalCase",
		},
		{
			name: "pascal case with acronyms",
			str:  "ACRONYMSPascalCaseACRONYMS",
			want: "acronymsPascalCaseACRONYMS",
		},
		{
			name: "all caps",
			str:  "ALLCAPS",
			want: "allcaps",
		},
		{
			name: "snake case",
			str:  "snake_case",
			want: "snakeCase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transfer.ToCamelCaseGO(tt.str)
			if got != tt.want {
				t.Errorf("ToCamelCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for target function.
		str  string
		want string
	}{
		{
			name: "empty",
			str:  "",
			want: "",
		},
		{
			name: "pascal case",
			str:  "PascalCase",
			want: "PascalCase",
		},
		{
			name: "camel case",
			str:  "camelCase",
			want: "CamelCase",
		},
		{
			name: "snake case",
			str:  "snake_case",
			want: "SnakeCase",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transfer.ToPascalCase(tt.str)
			if got != tt.want {
				t.Errorf("ToPascalCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
