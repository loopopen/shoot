package po

import (
	"database/sql"
	"examples/mapperexamplego127/infra/mapper"
	"time"
)

//go:generate go tool shoot map -path=../../domain/model -type=Order

type Order struct {
	*mapper.SQLMapper
	OrderedAt sql.Null[time.Time]
}
