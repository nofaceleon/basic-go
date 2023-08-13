package repository

import (
	"context"
	"gitee.com/geekbang/basic-go/webook/internal/domain"
	"gitee.com/geekbang/basic-go/webook/internal/repository/dao"
	"time"
)

var (
	ErrUserDuplicateEmail = dao.ErrUserDuplicateEmail
	ErrUserNotFound       = dao.ErrUserNotFound
)

type UserRepository struct {
	dao *dao.UserDAO //组合dao
}

func NewUserRepository(dao *dao.UserDAO) *UserRepository {
	return &UserRepository{
		dao: dao,
	}
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	// SELECT * FROM `users` WHERE `email`=?
	u, err := r.dao.FindByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}
	return domain.User{
		Id:       u.Id,
		Email:    u.Email,
		Password: u.Password,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) error {
	return r.dao.Insert(ctx, dao.User{
		Email:    u.Email,
		Password: u.Password,
	})
}

func (r *UserRepository) FindById(ctx context.Context, id int64) (domain.User, error) {
	// 先从 cache 里面找
	// 再从 dao 里面找
	user, err := r.dao.FindById(ctx, id) //这边查询出来的是模型
	if err != nil {
		return domain.User{}, nil
	}
	//返回的是领域对象
	return domain.User{
		Id:       user.Id,
		Email:    user.Email,
		Password: user.Password,
		NickName: user.NickName,
		Describe: user.Describe,
		BirthDay: user.BirthDay,
		Ctime:    time.Time{},
	}, nil
	// 找到了回写 cache
}

func (r *UserRepository) Edit(ctx context.Context, u domain.User, userId int64) error {
	return r.dao.Update(ctx, dao.User{
		NickName: u.NickName,
		Describe: u.Describe,
		BirthDay: u.BirthDay,
		Ctime:    time.Now().UnixMicro(),
	}, userId)
}
