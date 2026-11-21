package mapper

import "database/sql"

type SQLMapper struct{}

func (*SQLMapper) PtrToNull[T any](value *T) sql.Null[T] {
	if value == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{V: *value, Valid: true}
}

func (*SQLMapper) NullToPtr[T any](value sql.Null[T]) *T {
	if !value.Valid {
		return nil
	}
	return &value.V
}
