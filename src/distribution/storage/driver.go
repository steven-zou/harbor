package storage

//Driver defines related operations for storing distribution data.
//The data may include distribution history records and
//statistical metrics.
type Driver interface {
	//Handle history records
	HistoryRecordOperator

	//Handle statistical metrics
	MetricsOperator
}

//HistroryRecord represnets one record of the image preheating process.
type HistroryRecord struct{}

//QueryParam is a collection of parameters for querying preheating history records.
type QueryParam struct{}

//HistoryRecordOperator defines the related storing operations for history records.
type HistoryRecordOperator interface {
	//Append new preheating history record
	//If succeed, a nil error should be returned.
	AppendHistory(record HistroryRecord) error

	//Load history records on top of the query parameters
	//If succeed, a record list will be returned.
	//Otherwise, a non nil error will be set.
	LoadHistories(params QueryParam) ([]*HistroryRecord, error)
}

//MetricsOperator defines the related storing operations of statistical metrics.
type MetricsOperator interface{}
