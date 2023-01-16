package models

import (
	"github.com/go-playground/validator/v10"

	pb "keeper/internal/rpc/proto"
)

type User struct {
	ID       int    `json:"id" db:"id"`
	Username string `json:"login" db:"login" validate:"min=1"`
	Password string `json:"password" db:"password" validate:"min=1"`
}

func NewUser(request *pb.SignInRequest) User {
	return User{
		Username: request.Login,
		Password: request.Password,
	}
}

func (user *User) Validate() error {
	return validator.New().Struct(user)
}

func (user *User) SignInRequest() pb.SignInRequest {
	return pb.SignInRequest{
		Login:    user.Username,
		Password: user.Password,
	}
}
