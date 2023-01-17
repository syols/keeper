package models

import (
	pb "github.com/syols/keeper/internal/rpc/proto"
)

type LoginDetails struct {
	ID       *int   `json:"id" db:"id"`
	RecordID int    `json:"record_id" db:"record_id"`
	Login    string `json:"login" db:"column:login"`
	Password string `json:"password" db:"column:password"`
}

func NewLoginDetails(request *pb.LoginDetails) *LoginDetails {
	if request == nil {
		return nil
	}

	return &LoginDetails{
		Login:    request.GetLogin(),
		Password: request.GetPassword(),
	}
}

func (b LoginDetails) SetPrivateData(record *pb.Record) {
	record.PrivateData = &pb.Record_Login{Login: &pb.LoginDetails{Login: b.Login, Password: &b.Password}}
}

func (b LoginDetails) SetRecordId(id int) Details {
	b.RecordID = id
	return b
}
