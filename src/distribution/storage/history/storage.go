package history

import "github.com/goharbor/harbor/src/distribution/storage"

//Storage defines the related storing operations for history records.
type Storage interface {
	//Append new preheating history record
	//If succeed, a nil error should be returned.
	AppendHistory(record *HistroryRecord) error

	//Load history records on top of the query parameters
	//If succeed, a record list will be returned.
	//Otherwise, a non nil error will be set.
	LoadHistories(params *storage.QueryParam) ([]*HistroryRecord, error)
}
