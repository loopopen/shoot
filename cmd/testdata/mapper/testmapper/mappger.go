package testmapper

import (
	"time"

	"github.com/lopolopen/shoot/cmd/testdata/mapper/dest"
)

type Mapper struct {
	//shoot: map
}

func (Mapper) StringToDecimal(s string) dest.Decimal {
	return dest.Decimal{}
}

func (Mapper) DecimalToString(d dest.Decimal) string {
	return ""
}

func (Mapper) StringToTime(s string) time.Time {
	return time.Time{}
}

func (Mapper) TimeToString(t time.Time) string {
	return ""
}
