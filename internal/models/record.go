package models

import (
	"github.com/go-playground/validator/v10"

	pb "keeper/internal/rpc/proto"
)

const (
	TextType  = "TEXT"
	BlobType  = "BLOB"
	CardType  = "CARD"
	LoginType = "DETAIL"
)

type Details interface {
	SetPrivateData(record *pb.Record)
	SetRecordId(id int)
}

type Record struct {
	ID          int    `json:"id" db:"id"`
	UserID      int    `json:"user_id" db:"user_id"`
	Metadata    string `json:"metadata" db:"metadata"`
	DetailType  string `json:"detail" db:"detail" validate:"oneof=TEXT BLOB CARD DETAIL"`
	PrivateData Details
}

func NewRecord(request *pb.Record) *Record {
	if request == nil {
		return nil
	}

	result := &Record{
		ID:         int(request.GetId()),
		Metadata:   request.GetMetadata(),
		DetailType: request.GetDetailType(),
	}

	if result.DetailType == TextType {
		result.PrivateData = NewTextDetails(request.GetText())
	}

	if result.DetailType == BlobType {
		result.PrivateData = NewBlobDetails(request.GetBlob())
	}

	if result.DetailType == LoginType {
		result.PrivateData = NewLoginDetails(request.GetLogin())
	}

	if result.DetailType == CardType {
		result.PrivateData = NewCardDetails(request.GetCard())
	}

	return result
}

func (r *Record) Validate() error {
	return validator.New().Struct(r)
}

func (r Record) UpdatePrivateData(record *pb.Record) {
	r.PrivateData.SetPrivateData(record)
}
