package history

//HistroryRecord represnets one record of the image preheating process.
type HistroryRecord struct {
	Image     string
	Timestamp int64
	Status    string
	Provider  string
	Instance  string
}
