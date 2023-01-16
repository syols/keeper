package models

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "keeper/internal/rpc/proto"
)

type CardDetails struct {
	ID         *int                   `json:"id" db:"id"`
	RecordID   int                    `json:"record_id" db:"column:detail_id"`
	Number     string                 `json:"number" db:"column:number"`
	Cardholder string                 `json:"cardholder" db:"column:cardholder"`
	Cvc        uint32                 `json:"cvc" db:"column:cvc"`
	Expiration *timestamppb.Timestamp `json:"expiration" db:"column:expiration"`
}

func NewCardDetails(request *pb.CardDetails) *CardDetails {
	if request == nil {
		return nil
	}

	return &CardDetails{
		Number:     request.GetNumber(),
		Cardholder: request.GetCardholder(),
		Cvc:        request.GetCvc(),
		Expiration: request.Expiration,
	}
}

func (b CardDetails) SetPrivateData(record *pb.Record) {
	record.PrivateData = &pb.Record_Card{Card: &pb.CardDetails{
		Number:     b.Number,
		Cardholder: &b.Cardholder,
		Cvc:        &b.Cvc,
		Expiration: b.Expiration,
	}}
}

func (b CardDetails) SetRecordId(id int) {
	b.RecordID = id
}
