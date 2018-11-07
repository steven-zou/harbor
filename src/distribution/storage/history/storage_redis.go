package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/distribution/models"
	"github.com/goharbor/harbor/src/distribution/storage"
)

// RedisStorage implements storage based on redis backend.
type RedisStorage struct {
	redisBase *storage.RedisBase
}

// NewRedisStorage is constructor of RedisStorage
func NewRedisStorage(pool *redis.Pool, namespace string) *RedisStorage {
	if pool == nil || len(namespace) == 0 {
		return nil
	}

	return &RedisStorage{
		redisBase: storage.NewRedisBase(pool, namespace),
	}
}

// AppendHistory implements @Storage.AppendHistory
func (rs *RedisStorage) AppendHistory(record *models.HistoryRecord) error {
	if err := validHistoryRecord(record); err != nil {
		return err
	}

	// no id needed for history record
	// use time ticks as a placeholder key
	id := fmt.Sprintf("%d", time.Now().UnixNano())
	return rs.redisBase.Save(id, record)
}

// LoadHistories implements @Storage.LoadHistories
func (rs *RedisStorage) LoadHistories(params *storage.QueryParam) ([]*models.HistoryRecord, error) {
	rawData, err := rs.redisBase.List(params)
	if err != nil {
		return nil, err
	}

	results := []*models.HistoryRecord{}
	for _, jsonText := range rawData {
		r := &models.HistoryRecord{}
		if err := json.Unmarshal([]byte(jsonText), r); err != nil {
			return nil, err
		}

		results = append(results, r)
	}

	return results, nil
}

func validHistoryRecord(record *models.HistoryRecord) error {
	if record == nil {
		return errors.New("nil history record")
	}

	errs := []string{}
	val := reflect.ValueOf(record).Elem()
	for i := 0; i < val.NumField(); i++ {
		v := val.Field(i)
		t := val.Type().Field(i)
		switch t.Type.Kind() {
		case reflect.String:
			if len(v.Interface().(string)) == 0 {
				errs = append(errs, t.Name)
			}
		case reflect.Int64:
			if v.Int() == 0 {
				errs = append(errs, t.Name)
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("missing [%s]", strings.Join(errs, ","))
	}

	return nil
}
