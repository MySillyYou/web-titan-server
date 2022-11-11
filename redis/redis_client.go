package redis

import (
	log "web-server/alog"
	"github.com/gomodule/redigo/redis"
	"strings"
	"sync"
	"time"
)

var gDefaultRedis *RedisClient
var once sync.Once

type RedisClient struct {
	pool *redis.Pool
}

func GetInstance() *RedisClient {
	once.Do(func() {
		gDefaultRedis = &RedisClient{}
	})
	return gDefaultRedis
}

func (r *RedisClient) Init(addr string, db int) {

	pwd := ""

	idx := strings.Index(addr, "@")
	if idx >= 0 {
		pwd = addr[idx+1:]
		addr = addr[:idx]
	}

	if r.pool == nil {
		r.pool = &redis.Pool{
			MaxIdle:         2,
			MaxActive:       10,
			IdleTimeout:     time.Minute * 10,
			MaxConnLifetime: time.Minute * 30,
			Dial: func() (conn redis.Conn, err error) {
				conn, err = redis.Dial("tcp", addr, redis.DialDatabase(db), redis.DialPassword(pwd))
				if err != nil {
					return nil, err
				}

				return conn, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				if err != nil {
					log.Error(err.Error())
				}
				return err
			},
		}
	}

	_, err := r.innerDo("PING")
	if err != nil {
		panic(err.Error())
	}
}

func (r *RedisClient) Pool() *redis.Pool {
	return r.pool
}

func (r *RedisClient) Close() error {
	return r.pool.Close()
}

//func (r *RedisClient) Existed() bool {
//	if r.pool == nil {
//		return false
//	}
//	return true
//}

func (r *RedisClient) innerDo(commandName string, args ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()
	return conn.Do(commandName, args...)
}

func (r *RedisClient) Command(commandName string, args ...interface{}) (interface{}, error) {
	return r.innerDo(commandName, args...)
}

func (r *RedisClient) GetString(key string) (string, error) {
	return redis.String(r.innerDo("GET", key))
}

func (r *RedisClient) SetString(key, value string) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) GetByte(key string) ([]byte, error) {
	return redis.Bytes(r.innerDo("GET", key))
}

func (r *RedisClient) SetByte(key string, value []byte) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

//func (r *RedisClient) GetStruct(key string, value interface{}) error {
//	_, err := r.innerDo("GET", key)
//	return err
//}

func (r *RedisClient) SelectDB(value interface{}) error {
	_, err := r.innerDo("SELECT", value)
	return err
}

func (r *RedisClient) SetStruct(key string, value interface{}) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) GetMap(key string) (map[string]string, error) {
	return redis.StringMap(r.innerDo("HGETALL", key))
}

func (r *RedisClient) SetMap(key string, valMap map[string]string) error {
	_, err := r.innerDo("HMSET", key)
	return err
}

func (r *RedisClient) GetInt(key string) (int, error) {
	return redis.Int(r.innerDo("GET", key))
}

func (r *RedisClient) SetInt(key string, value int) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) GetInt64(key string) (int64, error) {
	return redis.Int64(r.innerDo("GET", key))
}

func (r *RedisClient) SetInt64(key string, value int64) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) GetFloat64(key string) (float64, error) {
	return redis.Float64(r.innerDo("GET", key))
}

func (r *RedisClient) SetFloat64(key string, value float64) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) SetMapKeyValue(key, field, value string) error {
	_, err := r.innerDo("HSET", key, field, value)
	return err
}

func (r *RedisClient) GetMapKeyValue(key, field string) (string, error) {
	return redis.String(r.innerDo("HGET", key, field))
}

func (r *RedisClient) SetKeyValue(key string, value interface{}) error {
	_, err := r.innerDo("SET", key, value)
	return err
}

func (r *RedisClient) ExistsMapKey(key string, field interface{}) (bool, error) {
	return redis.Bool(r.innerDo("HEXISTS", key, field))
}

// 设置过期时间
func (r *RedisClient) Expire(key string, t int64) error {
	_, err := r.innerDo("EXPIRE", key, t)
	return err
}

// 设置值并带过期时间
func (r *RedisClient) SetWithExpire(key string, val interface{}, seconds int64) error {
	_, err := r.innerDo("SET", key, val, "EX", seconds)
	return err
}

// 获取过期时间
func (r *RedisClient) GetTTL(key string) (int64, error) {
	return redis.Int64(r.innerDo("TTL", key))
}

// 删除key
func (r *RedisClient) DelKey(key string) error {
	_, err := r.innerDo("DEL", key)
	return err
}

func (r *RedisClient) Exists(key string) (bool, error) {
	return redis.Bool(r.innerDo("EXISTS", key))
}

// 自增
func (r *RedisClient) Incr(key string) (int64, error) {
	return redis.Int64(r.innerDo("INCR", key))
}

// 带步数的自增
func (r *RedisClient) IncrBy(key string, val int64) (int64, error) {
	return redis.Int64(r.innerDo("INCRBY", key, val))
}

// 自减
func (r *RedisClient) Decr(key string) (int64, error) {
	return redis.Int64(r.innerDo("DECR", key))
}

// 带步数的自减
func (r *RedisClient) DecrBy(key string, val int64) (int64, error) {
	return redis.Int64(r.innerDo("DECRBY", key, val))
}

func (r *RedisClient) LPush(key string, val ...interface{}) error {
	args := make([]interface{}, 0, len(val)+1)
	args = append(args, key)
	args = append(args, val...)

	_, err := r.innerDo("LPUSH", args...)
	return err
}

func (r *RedisClient) LPop(key string) (interface{}, error) {
	return r.innerDo("LPOP", key)
}

func (r *RedisClient) LPopInt64(key string) (int64, error) {
	return redis.Int64(r.LPop(key))
}

func (r *RedisClient) LPopInt(key string) (int, error) {
	return redis.Int(r.LPop(key))
}

func (r *RedisClient) LPopString(key string) (string, error) {
	return redis.String(r.LPop(key))
}

func (r *RedisClient) LLen(key string) int64 {
	length, _ := redis.Int64(r.innerDo("LLEN", key))
	return length
}

func (r *RedisClient) RPush(key string, val ...interface{}) error {
	args := make([]interface{}, 0, len(val)+1)
	args = append(args, key)
	args = append(args, val...)

	_, err := r.innerDo("RPUSH", args...)
	return err
}

func (r *RedisClient) RPop(key string) (interface{}, error) {
	return r.innerDo("RPOP", key)
}

func (r *RedisClient) RPopInt64(key string) (int64, error) {
	return redis.Int64(r.RPop(key))
}

func (r *RedisClient) RPopInt(key string) (int, error) {
	return redis.Int(r.RPop(key))
}

func (r *RedisClient) RPopString(key string) (string, error) {
	return redis.String(r.RPop(key))
}

func (r *RedisClient) BLPop(timeout int64, keys ...interface{}) ([]interface{}, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, keys...)
	args = append(args, timeout)

	return redis.Values(r.innerDo("BLPOP", args...))
}

func (r *RedisClient) BRPop(timeout int64, keys ...interface{}) ([]interface{}, error) {
	args := make([]interface{}, 0, len(keys)+1)
	args = append(args, keys...)
	args = append(args, timeout)

	return redis.Values(r.innerDo("BRPOP", args...))
}

func (r *RedisClient) RPopLPush(src, dst string) (interface{}, error) {
	return r.innerDo("RPOPLPUSH", src, dst)
}

func (r *RedisClient) BRPopLPush(src, dst string, timeout int64) (interface{}, error) {
	return r.innerDo("BRPOPLPUSH", src, dst, timeout)
}

/*
	redis操作lua脚本，使用方式请参考redis的EVAL命令
*/
func (r *RedisClient) DoScript(script string, keyCount int, keysAndArgs ...interface{}) (interface{}, error) {
	conn := r.pool.Get()
	defer conn.Close()

	getScript := redis.NewScript(keyCount, script)
	return getScript.Do(conn, keysAndArgs...)
}

/*
	先判断再自增，原子操作。
	输入参数：key为redis的键，maxVal是需要进行判断的最大值，expire过期删除的时间，单位为秒。
			如果需要在自增之后设置该KEY的过期时间，将expire参数设为大于0，不设置过期时间请将expire设为0
	返回参数：int为返回自增后的值，error为错误信息。
		如果达到上限，则返回值int为-1，未达到上限返回值int为自增后的值
*/
func (r *RedisClient) IncrAndJudge(key string, maxVal, expire int) (int, error) {
	script := `
		if redis.call('exists', KEYS[1]) == 1 then
			local val = redis.call('get', KEYS[1])
			if tonumber(val) >= tonumber(ARGV[1]) then
				return -1
			end
			if tonumber(val) == 1 and tonumber(ARGV[2]) > 0 then
				redis.call('expire', KEYS[1], ARGV[2])
			end
		end
		return redis.call('incr', KEYS[1])`

	return redis.Int(r.DoScript(script, 1, key, maxVal, expire))
}

func (r *RedisClient) SRandMemberString(key string) (string, error) {
	return redis.String(r.innerDo("SRANDMEMBER", key))
}

func (r *RedisClient) SAddString(key string, val ...string) (int64, error) {
	args := make([]interface{}, 0, len(val)+1)
	args = append(args, key)
	for _, v := range val {
		args = append(args, v)
	}
	return redis.Int64(r.innerDo("SADD", args...))
}

func (r *RedisClient) SPopString(key string) (string, error) {
	return redis.String(r.innerDo("SPOP", key))
}

func (r *RedisClient) SCard(key string) (int64, error) {
	return redis.Int64(r.innerDo("SCARD", key))
}

func (r *RedisClient) ZAdd(key string, scoresAndKeys ...interface{}) (interface{}, error) {
	args := make([]interface{}, 0, len(scoresAndKeys)+1)
	args = append(args, key)
	args = append(args, scoresAndKeys...)

	return r.innerDo("ZADD", args...)
}

//返回有序集 key 中，成员 member 的 score 值
func (r *RedisClient) ZScore(key, member string) (int64, error) {
	return redis.Int64(r.innerDo("ZSCORE", key, member))
}

//返回有序集 key 的数量
func (r *RedisClient) ZCard(key string) (int64, error) {
	return redis.Int64(r.innerDo("ZCARD", key))
}

//返回有序集 key 中， score 值在 min 和 max 之间(默认包括 score 值等于 min 或 max )的成员的数量
func (r *RedisClient) ZCount(key string, min, max int64) (int64, error) {
	return redis.Int64(r.innerDo("ZCOUNT", key, min, max))
}

//为有序集 key 的成员 member 的 score 值加上增量 increment
//可以通过传递一个负数值 increment ，让 score 减去相应的值，比如 ZINCRBY key -5 member ，就是让 member 的 score 值减去 5
func (r *RedisClient) ZIncrBy(key string, increment int64, member string) (int64, error) {
	return redis.Int64(r.innerDo("ZINCRBY", key, increment, member))
}

func (r *RedisClient) ZRange(key string, start, stop int64, withScores string) ([]string, error) {
	if strings.ToUpper(withScores) != "WITHSCORES" {
		withScores = ""
	}
	return redis.Strings(r.innerDo("ZRANGE", key, start, stop, withScores))
}

func (r *RedisClient) ZRangeWithScores(key string, start, stop int64) ([]string, error) {
	return r.ZRange(key, start, stop, "WITHSCORES")
}

func (r *RedisClient) ZRangeWithNoScores(key string, start, stop int64) ([]string, error) {
	return r.ZRange(key, start, stop, "")
}

func (r *RedisClient) ZRevRange(key string, start, stop int64, withScores string) ([]string, error) {
	if withScores != "WITHSCORES" {
		withScores = ""
	}
	return redis.Strings(r.innerDo("ZREVRANGE", key, start, stop, withScores))
}

func (r *RedisClient) ZRevRangeWithScores(key string, start, stop int64) ([]string, error) {
	return r.ZRevRange(key, start, stop, "WITHSCORES")
}

func (r *RedisClient) ZRevRangeWithNoScores(key string, start, stop int64) ([]string, error) {
	return r.ZRevRange(key, start, stop, "")
}

func (r *RedisClient) ZRem(key string, members ...interface{}) error {
	args := make([]interface{}, 0, len(members)+1)
	args = append(args, key)
	args = append(args, members...)

	_, err := r.innerDo("ZREM", args...)
	return err
}

//移除有序集 key 中，指定排名(rank)区间内的所有成员
func (r *RedisClient) ZRemRangeByRank(key string, start, stop int64) error {
	_, err := r.innerDo("ZREMRANGEBYRANK", key, start, stop)
	return err
}
