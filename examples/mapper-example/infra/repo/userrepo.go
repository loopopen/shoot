package repo

import (
	"context"
	"mapperexample/domain/model"
	"mapperexample/domain/repo"
	"mapperexample/infra/po"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	r := &UserRepo{db: db}
	var _ repo.UserRepo = r
	return r
}

func (r *UserRepo) Get(ctx context.Context, id uint) (*model.User, error) {
	var userPO po.User
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&userPO).Error
	if err != nil {
		return nil, err
	}

	var addrPOs []*po.UserAddress
	err = r.db.WithContext(ctx).
		Where("user_id = ?", id).
		Scan(&addrPOs).Error
	if err != nil {
		return nil, err
	}

	userPO.AddressList = addrPOs
	user := userPO.ToDomain()
	return user, nil
}
