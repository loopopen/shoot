package mapper

import (
	"database/sql"
	"time"
)

type SQLMapper struct {
	//shoot: mapper
}

//todo: receiver must be pointer type

func (*SQLMapper) TimePtrToNullTime(t *time.Time) sql.Null[time.Time] {
	if t == nil {
		return sql.Null[time.Time]{}
	}
	return sql.Null[time.Time]{
		V:     *t,
		Valid: true,
	}
}
