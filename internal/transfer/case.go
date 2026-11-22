package transfer

import (
	"fmt"
	"strings"
)

type Case string

const (
	CasePascal Case = "pascal"
	CaseCamel  Case = "camel"
	CaseSnake  Case = "snake"
	CaseLower  Case = "lower"
	CaseUpper  Case = "upper"
)

func (c *Case) Set(value string) error {
	switch Case(value) {
	case CasePascal, CaseCamel, CaseSnake, CaseLower, CaseUpper:
		*c = Case(value)
		return nil
	default:
		return fmt.Errorf("invalid case: %s", value)
	}
}

func (c Case) String() string {
	return string(c)
}

func (c Case) Transform() func(string) string {
	switch c {
	case CasePascal:
		return ToPascalCase
	case CaseCamel:
		return ToCamelCase
	case CaseSnake:
		return ToSnakeCase
	case CaseLower:
		return strings.ToLower
	case CaseUpper:
		return strings.ToUpper
	default:
		return ID
	}
}

func (c Case) Apply(value string) string {
	return c.Transform()(value)
}
