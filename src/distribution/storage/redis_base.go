package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RawJSON represents the raw JSON text.
type RawJSON string

// RedisBase provides basic object storage capabilities based on redis backend.
// Object is stored as a whole raw json text.
// Objects are stored in a ZSET structure.
type RedisBase struct {
	pool      *redis.Pool
	namespace string
}

// Get the object with the specified key
func (rb *RedisBase) Get(key string) (RawJSON, error) {
	if len(key) == 0 {
		return "", errors.New("nil key")
	}

	conn := rb.pool.Get()
	defer conn.Close()

	score, err := redis.Int64(conn.Do("HGET", storageIndexKey(rb.namespace), key))
	if err != nil {
		return "", err
	}

	args := []interface{}{
		storageListKey(rb.namespace),
		score,
		score,
	}

	rawJSONs, err := redis.Strings(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil {
		if err == redis.ErrNil {
			return "", nil // not existing
		}

		return "", err
	}

	if len(rawJSONs) == 0 {
		return "", nil
	}

	return RawJSON(rawJSONs[0]), nil
}

// Save or update the object
func (rb *RedisBase) Save(key string, object interface{}) error {
	if object == nil {
		return errors.New("object is nil")
	}

	rawJSON, err := toJSON(object)
	if err != nil {
		return err
	}
	score := time.Now().UnixNano()

	conn := rb.pool.Get()
	defer conn.Close()

	if err := conn.Send("MULTI"); err != nil {
		return err
	}
	if err := conn.Send("ZADD", storageListKey(rb.namespace), score, rawJSON); err != nil {
		return err
	}
	if err := conn.Send("HSET", storageIndexKey(rb.namespace), key, score); err != nil {
		return err
	}
	if _, err := conn.Do("EXEC"); err != nil {
		return err
	}

	return nil
}

// Delete the specified object with the key
func (rb *RedisBase) Delete(key string) error {
	if len(key) == 0 {
		return errors.New("nil key")
	}

	conn := rb.pool.Get()
	defer conn.Close()

	score, err := redis.Int64(conn.Do("HGET", storageIndexKey(rb.namespace), key))
	if err != nil {
		return err
	}
	args := []interface{}{
		storageListKey(rb.namespace),
		score,
		score,
	}

	_, err = conn.Do("ZREMRANGEBYSCORE", args...)

	return err
}

// List the objects
func (rb *RedisBase) List(queryParam *QueryParam) ([]RawJSON, error) {
	var page, pageSize uint = 1, 25

	if queryParam != nil {
		if queryParam.Page > 0 {
			page = queryParam.Page
		}

		if queryParam.PageSize > 0 {
			pageSize = queryParam.PageSize
		}
	}

	// Total
	size, err := rb.Size()
	if err != nil {
		return nil, err
	}

	if size == 0 {
		return []RawJSON{}, nil
	}

	pages := (uint)(size / 25)
	if size%25 != 0 {
		pages++
	}

	if page > pages {
		return []RawJSON{}, nil
	}

	// Desc ordered
	start := (pages - page) * pageSize
	end := (int)(start + 24)
	if end >= (int)(size) {
		end = -1
	}

	conn := rb.pool.Get()
	defer conn.Close()

	args := []interface{}{
		storageListKey(rb.namespace),
		start,
		end,
	}

	rawJSONs, err := redis.Strings(conn.Do("ZREVRANGE", args...))
	if err != nil {
		return nil, err
	}

	results := []RawJSON{}
	for _, jsonText := range rawJSONs {
		shouldApppend := true

		if len(queryParam.Keyword) > 0 {
			if !strings.Contains(jsonText, queryParam.Keyword) {
				shouldApppend = false
			}
		}

		if shouldApppend {
			results = append(results, RawJSON(jsonText))
		}
	}

	return results, nil
}

// Size return the len of the list
func (rb *RedisBase) Size() (int64, error) {
	conn := rb.pool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do("ZCARD", storageListKey(rb.namespace)))
}

func toJSON(object interface{}) (RawJSON, error) {
	jsonData, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	return RawJSON(jsonData), nil
}

func storageListKey(ns string) string {
	return fmt.Sprintf("%s:store", ns)
}

func storageIndexKey(ns string) string {
	return fmt.Sprintf("%s:index", ns)
}
