package repo

import "context"

type RepoBase[TID any, TEntity any] interface {
	Get(ctx context.Context, id TID) (TEntity, error)
}
