package instance

import (
	"encoding/json"
	"errors"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/distribution/storage"
)

// RedisStorage implements the storage interface based on redis backend
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

// Save implements @Storage.Save
func (rs *RedisStorage) Save(inst *Metadata) (string, error) {
	if inst == nil {
		return "", errors.New("nil instance metadata")
	}

	inst.ID = storage.UUID()

	if err := rs.redisBase.Save(inst.ID, inst); err != nil {
		return "", err
	}

	return inst.ID, nil
}

// Delete implements @Storage.Delete
func (rs *RedisStorage) Delete(id string) error {
	if len(id) == 0 {
		return errors.New("empty id")
	}

	if !rs.redisBase.Exists(id) {
		return storage.ErrObjectNotFound
	}

	return rs.redisBase.Delete(id)
}

// Update implements @Storage.Update
func (rs *RedisStorage) Update(inst *Metadata) error {
	if inst == nil {
		return errors.New("nil instance metadata")
	}

	if len(inst.ID) == 0 {
		return errors.New("missing id of instance metadata")
	}

	if !rs.redisBase.Exists(inst.ID) {
		return storage.ErrObjectNotFound
	}

	return rs.redisBase.Save(inst.ID, inst)
}

// Get implements @Storage.Get
func (rs *RedisStorage) Get(id string) (*Metadata, error) {
	if !rs.redisBase.Exists(id) {
		return nil, storage.ErrObjectNotFound
	}

	raw, err := rs.redisBase.Get(id)
	if err != nil {
		return nil, err
	}

	inst := &Metadata{}
	if err := json.Unmarshal([]byte(raw), inst); err != nil {
		return nil, err
	}

	return inst, nil
}

// List implements @Storage.List
func (rs *RedisStorage) List(param *storage.QueryParam) ([]*Metadata, error) {
	raws, err := rs.redisBase.List(param)
	if err != nil {
		return nil, err
	}

	results := []*Metadata{}
	for _, raw := range raws {
		m := &Metadata{}
		if err := json.Unmarshal([]byte(raw), m); err != nil {
			return nil, err
		}

		results = append(results, m)
	}

	return results, nil
}
