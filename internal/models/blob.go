package models

import (
	pb "keeper/internal/rpc/proto"
)

type BlobDetails struct {
	ID       *int   `json:"id" db:"id"`
	RecordID int    `json:"record_id" db:"column:detail_id"`
	Data     []byte `json:"data" db:"column:data"`
}

func NewBlobDetails(request *pb.BlobDetails) *BlobDetails {
	if request == nil {
		return nil
	}

	return &BlobDetails{
		Data: request.GetBlob(),
	}
}

func (b BlobDetails) SetPrivateData(record *pb.Record) {
	record.PrivateData = &pb.Record_Blob{Blob: &pb.BlobDetails{Blob: b.Data}}
}

func (b BlobDetails) SetRecordId(id int) {
	b.RecordID = id
}
