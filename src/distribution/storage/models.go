package storage

//QueryParam is a collection of parameters for querying preheating history records.
type QueryParam struct {
	Page     uint
	PageSize uint
	Keyword  string
}
