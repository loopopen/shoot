package shoot

import (
	"fmt"

	"github.com/loopopen/shoot/constraints"
)

// ParseEnum attempts to convert a string into its corresponding enum value.
// It uses the ValueMap() provided by the enum type to look up the value.
// Returns an error if the string does not match any known enum value.
func ParseEnum[T constraints.EnumShooter[T]](str string) (T, error) {
	var t T
	m := t.ValueMap()
	t, ok := m[str]
	if !ok {
		return t, fmt.Errorf("requested value '%s' was not found", str)
	}
	return t, nil
}

// TryParseEnum is a safe variant of ParseEnum.
// It attempts to parse the string into an enum value and stores the result in v.
// Returns true if successful, false otherwise. Does not return an error.
func TryParseEnum[T constraints.EnumShooter[T]](str string, v *T) bool {
	t, err := ParseEnum[T](str)
	if err != nil {
		return false
	}
	*v = t
	return true
}

// IsEnum checks whether a given integer value is a valid enum value of type T.
// Useful for validating raw input before casting or using it as an enum.
func IsEnum[T constraints.EnumShooter[T], TV constraints.Integer](value TV) bool {
	var t T
	for _, v := range t.Values() {
		if v == T(value) {
			return true
		}
	}
	return false
}
