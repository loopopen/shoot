package transfer

import (
	"strings"
)

func ID(x string) string { return x }

func FirstLowerLetter(str string) string {
	if str == "" {
		return str
	}
	first := str[:1]
	return strings.ToLower(first)
}

func ToCamelCaseGO(str string) string {
	if len(str) == 0 {
		return str
	}

	if str == strings.ToUpper(str) {
		return strings.ToLower(str)
	}

	str = ToPascalCase(str)

	bytes := []byte(str)
	i := 0
	for i < len(bytes) && IsUpper(bytes[i]) {
		i++
	}

	if i <= 1 {
		bytes[0] = ToLower(bytes[0])
		return string(bytes)
	}

	prefix := strings.ToLower(string(bytes[:i-1]))
	prefix += string(bytes[i-1])

	return prefix + string(bytes[i:])
}

func ToCamelCase(str string) string {
	if len(str) == 0 {
		return str
	}

	str = ToPascalCase(str)

	tokens := splitCamelTokensASCII(str)

	for i := range tokens {
		if i == 0 {
			tokens[i] = strings.ToLower(tokens[i])
		} else {
			if len(tokens[i]) > 1 {
				tokens[i] = strings.ToUpper(tokens[i][:1]) + strings.ToLower(tokens[i][1:])
			} else {
				tokens[i] = strings.ToUpper(tokens[i])
			}
		}
	}

	return strings.Join(tokens, "")
}

func splitCamelTokensASCII(str string) []string {
	var tokens []string
	start := 0
	n := len(str)

	for i := 1; i < n; i++ {
		if IsUpper(str[i]) && (IsLower(str[i-1]) || (i+1 < n && IsLower(str[i+1]))) {
			tokens = append(tokens, str[start:i])
			start = i
		}
	}

	tokens = append(tokens, str[start:])
	return tokens
}

func ToPascalCase(str string) string {
	if len(str) == 0 {
		return str
	}

	tokens := strings.Split(str, "_")
	for i, part := range tokens {
		if len(part) == 0 {
			continue
		}
		bytes := []byte(part)
		bytes[0] = ToUpper(bytes[0])
		tokens[i] = string(bytes)
	}
	return strings.Join(tokens, "")
}

func ToUpper(b byte) byte {
	if IsLower(b) {
		return b - 'a' + 'A'
	}
	return b
}

func ToLower(b byte) byte {
	if IsUpper(b) {
		return b - 'A' + 'a'
	}
	return b
}

func IsUpper(b byte) bool { return b >= 'A' && b <= 'Z' }

func IsLower(b byte) bool { return b >= 'a' && b <= 'z' }
