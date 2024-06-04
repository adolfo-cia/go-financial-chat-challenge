package user

import (
	"context"
	"time"
)

type User struct {
	Username       string    `json:"username"`
	HashedPassword string    `json:"-"`
	CreatedAt      time.Time `json:"createdAt"`
}

type userRepo interface {
	Save(ctx context.Context, user User) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
}

type hashPassword func(password string) (string, error)

type checkPassword func(password string, hashedPassword string) error

type tokenMaker interface {
	CreateToken(username string, duration time.Duration) (string, error)
}

type service struct {
	repo          userRepo
	hashPassword  hashPassword
	checkPassword checkPassword
	tokenMaker    tokenMaker
}

func NewService(
	repo userRepo,
	hashPassword hashPassword,
	checkPassword checkPassword,
	tokenMaker tokenMaker) *service {
	return &service{
		repo:          repo,
		hashPassword:  hashPassword,
		checkPassword: checkPassword,
		tokenMaker:    tokenMaker}
}

type createUserReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *service) Create(ctx context.Context, req createUserReq) (User, error) {
	hashedPassword, err := s.hashPassword(req.Password)
	if err != nil {
		return User{}, err
	}

	newUser, err := s.repo.Save(ctx, User{
		Username:       req.Username,
		HashedPassword: hashedPassword})
	if err != nil {
		return User{}, err
	}

	return newUser, nil
}

type loginReq struct {
	createUserReq
}

type loginRes struct {
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

func (s *service) Login(ctx context.Context, req loginReq) (loginRes, error) {
	user, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return loginRes{}, err
	}

	err = s.checkPassword(req.Password, user.HashedPassword)
	if err != nil {
		return loginRes{}, err
	}

	token, err := s.tokenMaker.CreateToken(user.Username, 15*time.Minute)
	if err != nil {
		return loginRes{}, err
	}

	return loginRes{AccessToken: token, User: user}, nil
}
