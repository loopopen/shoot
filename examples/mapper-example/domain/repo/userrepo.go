package repo

import "mapperexample/domain/model"

type UserRepo interface {
	RepoBase[uint, *model.User]
}
