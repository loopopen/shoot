//go:build !go1.27

package mapper

import (
	"go/types"

	"github.com/loopopen/shoot/internal/shoot"
)

func mapperFuncMatches(fn shoot.Func, param, result types.Type) bool {
	return shoot.TypeEquals(fn.Param, param) && shoot.TypeEquals(fn.Result, result)
}
