package history

import "github.com/goharbor/harbor/src/distribution/storage"

// RedisStorage implements storage based on redis backend.
type RedisStorage struct{}

// AppendHistory implements @Storage.AppendHistory
func (rs *RedisStorage) AppendHistory(record HistroryRecord) error {
	return nil
}

// LoadHistories implements @Storage.LoadHistories
func (rs *RedisStorage) LoadHistories(params storage.QueryParam) ([]*HistroryRecord, error) {
	return nil, nil
}
