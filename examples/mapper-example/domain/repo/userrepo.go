package repo

import "examples/mapperexample/domain/model"

type UserRepo interface {
	RepoBase[uint, *model.User]
}
