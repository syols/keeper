package models

import (
	pb "github.com/syols/keeper/internal/rpc/proto"
)

type TextDetails struct {
	ID       *int   `json:"id" db:"id"`
	RecordID int    `json:"record_id" db:"record_id"`
	Data     string `json:"data" db:"data"`
}

func NewTextDetails(request *pb.TextDetails) *TextDetails {
	if request == nil {
		return nil
	}

	return &TextDetails{
		Data: request.GetText(),
	}
}

func (b TextDetails) SetPrivateData(record *pb.Record) {
	record.PrivateData = &pb.Record_Text{Text: &pb.TextDetails{Text: b.Data}}
}

func (b TextDetails) SetRecordId(id int) Details {
	b.RecordID = id
	return b
}
