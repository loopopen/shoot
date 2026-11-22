package mapper

import "database/sql"

type Mapper struct{}

func (*Mapper) PtrToNull[T any](value *T) sql.Null[T] {
	if value == nil {
		return sql.Null[T]{}
	}
	return sql.Null[T]{V: *value, Valid: true}
}

func (*Mapper) NullToPtr[T any](value sql.Null[T]) *T {
	if !value.Valid {
		return nil
	}
	return &value.V
}
