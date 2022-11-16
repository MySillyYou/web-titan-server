package redis

import (
	"fmt"
	"github.com/hoisie/redis"
	"strconv"
	"strings"
	log "web-server/alog"
)

var Redis Util

type Util struct {
	client *redis.Client
}

func (r *Util) Init(addr string, db int) {

	pwd := ""

	idx := strings.Index(addr, "@")
	if idx >= 0 {
		pwd = addr[idx+1:]
		addr = addr[:idx]
	}

	if r.client == nil {
		r.client = &redis.Client{
			Addr:        addr,
			Db:          db,
			Password:    pwd,
			MaxPoolSize: 10,
		}
	}
}

func (r *Util) GetClient() *redis.Client {
	return r.client
}

func (r *Util) Existed() bool {
	if r.client == nil {
		return false
	}
	return true
}

func (r *Util) GetString(key string) (string, error) {
	val, err := r.client.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") { // 关闭Key 不存在的error log输出
			log.Error(err.Error())
		}
		return "", err
	}

	return string(val), nil
}

func (r *Util) SetString(key, value string) error {
	err := r.client.Set(key, []byte(value))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetByte(key string) ([]byte, error) {
	val, err := r.client.Get(key)
	if err != nil {
		log.Error(err.Error())
		return []byte(""), err
	}

	return val, nil
}

func (r *Util) SetByte(key string, value []byte) error {
	err := r.client.Set(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetStruct(key string, value interface{}) error {

	err := r.client.Hgetall(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) SetStruct(key string, value interface{}) error {
	err := r.client.Hmset(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetMap(key string, value interface{}) error {

	err := r.client.Hgetall(key, value)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return err
	}

	return nil
}

func (r *Util) SetMap(key string, value interface{}) error {
	err := r.client.Hmset(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetInt(key string) (int, error) {
	val, err := r.client.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return 0, err
	}

	value, err := strconv.Atoi(string(val))
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}

	return value, nil
}

func (r *Util) SetInt(key string, value int) error {
	s := strconv.Itoa(value)
	err := r.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetInt64(key string) (int64, error) {
	val, err := r.client.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return 0, err
	}

	value, err := strconv.ParseInt(string(val), 10, 64)
	if err != nil {
		log.Error(err.Error())
		return 0, err
	}

	return value, nil
}

func (r *Util) SetInt64(key string, value int64) error {
	s := fmt.Sprintf("%d", value)
	err := r.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetFloat(key string) (float64, error) {
	val, err := r.client.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return 0.0, err
	}

	value, err := strconv.ParseFloat(string(val), 64)
	if err != nil {
		log.Error(err.Error())
		return 0.0, err
	}

	return value, nil
}

func (r *Util) SetFloat(key string, value float64) error {
	s := fmt.Sprintf("%f", value)
	err := r.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) SetMapKeyValue(main_key, key, value string) error {
	_, err := r.client.Hset(main_key, key, []byte(value))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (r *Util) GetMapKeyValue(main_key, key string) (string, error) {
	valByte, err := r.client.Hget(main_key, key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return "", err
	}

	return string(valByte), nil
}

// SetTTL 设置过期时间
func (r *Util) SetTTL(key string, t int64) error {
	_, err := r.client.Expire(key, t)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// GetTTL 获取过期时间
func (r *Util) GetTTL(key string) (int64, error) {
	ttl, err := r.client.Ttl(key)
	if err != nil {
		log.Error(err)
		return -2, err
	}
	return ttl, nil
}

// DelKey 删除key
func (r *Util) DelKey(key string) error {
	_, err := r.client.Del(key)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// Incr 自增
func (r *Util) Incr(key string) (int64, error) {
	res, err := r.client.Incr(key)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

// IncrStep 带步数的自增
func (r *Util) IncrStep(key string, val int64) (int64, error) {
	res, err := r.client.Incrby(key, val)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

func (r *Util) Decr(key string) (int64, error) {
	res, err := r.client.Decr(key)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

func (r *Util) Decrby(key string, val int64) (int64, error) {
	res, err := r.client.Decrby(key, val)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}
