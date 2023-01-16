package models

import (
	pb "keeper/internal/rpc/proto"
)

type TextDetails struct {
	ID       *int   `json:"id" db:"id"`
	RecordID int    `json:"record_id" db:"column:detail_id"`
	Data     string `json:"data" db:"column:data"`
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

func (b TextDetails) SetRecordId(id int) {
	b.RecordID = id
}
