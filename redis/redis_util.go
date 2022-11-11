package redis

import (
	log "web-server/alog"
	"fmt"
	"strconv"
	"strings"
	"github.com/hoisie/redis"
)

var Redis RedisUtil

type RedisUtil struct {
	client *redis.Client
}

func (this *RedisUtil) Init(addr string, db int) {

	pwd := ""

	idx := strings.Index(addr, "@")
	if idx >= 0 {
		pwd = addr[idx+1:]
		addr = addr[:idx]
	}

	if this.client == nil {
		this.client = &redis.Client{
			Addr:        addr,
			Db:          db,
			Password:    pwd,
			MaxPoolSize: 10,
		}
	}
}

func (this *RedisUtil) GetClient() *redis.Client {
	return this.client
}

func (this *RedisUtil) Existed() bool {
	if this.client == nil {
		return false
	}
	return true
}

func (this *RedisUtil) GetString(key string) (string, error) {
	val, err := this.client.Get(key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") { // 关闭Key 不存在的error log输出
			log.Error(err.Error())
		}
		return "", err
	}

	return string(val), nil
}

func (this *RedisUtil) SetString(key, value string) error {
	err := this.client.Set(key, []byte(value))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetByte(key string) ([]byte, error) {
	val, err := this.client.Get(key)
	if err != nil {
		log.Error(err.Error())
		return []byte(""), err
	}

	return val, nil
}

func (this *RedisUtil) SetByte(key string, value []byte) error {
	err := this.client.Set(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetStruct(key string, value interface{}) error {

	err := this.client.Hgetall(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) SetStruct(key string, value interface{}) error {
	err := this.client.Hmset(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetMap(key string, value interface{}) error {

	err := this.client.Hgetall(key, value)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return err
	}

	return nil
}

func (this *RedisUtil) SetMap(key string, value interface{}) error {
	err := this.client.Hmset(key, value)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetInt(key string) (int, error) {
	val, err := this.client.Get(key)
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

func (this *RedisUtil) SetInt(key string, value int) error {
	s := strconv.Itoa(value)
	err := this.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetInt64(key string) (int64, error) {
	val, err := this.client.Get(key)
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

func (this *RedisUtil) SetInt64(key string, value int64) error {
	s := fmt.Sprintf("%d", value)
	err := this.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetFloat(key string) (float64, error) {
	val, err := this.client.Get(key)
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

func (this *RedisUtil) SetFloat(key string, value float64) error {
	s := fmt.Sprintf("%f", value)
	err := this.client.Set(key, []byte(s))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) SetMapKeyValue(main_key, key, value string) error {
	_, err := this.client.Hset(main_key, key, []byte(value))
	if err != nil {
		log.Error(err.Error())
		return err
	}

	return nil
}

func (this *RedisUtil) GetMapKeyValue(main_key, key string) (string, error) {
	valByte, err := this.client.Hget(main_key, key)
	if err != nil {
		if !strings.Contains(err.Error(), "not exist") {
			log.Error(err.Error())
		}
		return "", err
	}

	return string(valByte), nil
}

// 设置过期时间
func (this *RedisUtil) SetTTL(key string, t int64) error {
	_, err := this.client.Expire(key, t)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// 获取过期时间
func (this *RedisUtil) GetTTL(key string) (int64, error) {
	ttl, err := this.client.Ttl(key)
	if err != nil {
		log.Error(err)
		return -2, err
	}
	return ttl, nil
}

// 删除key
func (this *RedisUtil) DelKey(key string) error {
	_, err := this.client.Del(key)
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

// 自增
func (this *RedisUtil) Incr(key string) (int64, error) {
	res, err := this.client.Incr(key)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

// 带步数的自增
func (this *RedisUtil) Incrby(key string, val int64) (int64, error) {
	res, err := this.client.Incrby(key, val)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

func (this *RedisUtil) Decr(key string) (int64, error) {
	res, err := this.client.Decr(key)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}

func (this *RedisUtil) Decrby(key string, val int64) (int64, error) {
	res, err := this.client.Decrby(key, val)
	if err != nil {
		log.Error(err)
		return -1, err
	}
	return res, nil
}
