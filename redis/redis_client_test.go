package redis

import (
	"testing"
	"time"
)

func TestRedisClient_Some(t *testing.T) {
	GetInstance().Init("192.168.1.222:6379", 0)

	index, err := GetInstance().Incr("test_client")
	if err != nil {
		t.Log(err.Error())
		return
	}

	t.Log("incr index", index)

	idx, err := GetInstance().GetInt64("test_client")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("get index", idx)

	err = GetInstance().SetString("test_string", "haha")
	if err != nil {
		t.Log(err.Error())
		return
	}

}

func TestRedisClient_SetWithExpire(t *testing.T) {
	GetInstance().Init("192.168.1.222:6379", 0)

	//err := gDefaultRedis.SetWithExpire("test_expire", "haha", 60)
	//if err != nil {
	//	t.Log(err.Error())
	//	return
	//}

	//ttl, err := GetInstance().GetTTL("test_expire")
	//if err != nil {
	//	t.Log(err.Error())
	//	return
	//}
	//t.Log("test_expire ttl", ttl)

	d, err := GetInstance().GetInt("sdfsdfsdf")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("sdfsdfsdf", d)
}

func TestRedisClient_DoScript(t *testing.T) {
	GetInstance().Init("192.168.1.222:6379", 0)

	start := time.Now()

	//script := `
	//	if redis.call('exists', KEYS[1]) == 1 and redis.call('get', KEYS[1]) >= ARGV[1] then
	//		return -1
	//	end
	//	return redis.call('incr', KEYS[1])`
	//
	//d, err := GetInstance().DoScript(script, 1, "script_test", 10)
	//if err != nil {
	//	t.Log(err.Error())
	//	return
	//}

	d, err := GetInstance().IncrAndJudge("script_test", 6, 30)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("script_test", d)
	t.Log("use time", time.Since(start))
}

func TestRedisClient_Hash(t *testing.T) {
	GetInstance().Init("192.168.1.222:6379", 0)

	err := GetInstance().SetMapKeyValue("test_hash", "1", "1")
	if err != nil {
		t.Log(err.Error())
		return
	}

	err = GetInstance().SetMapKeyValue("test_hash", "2", "2")
	if err != nil {
		t.Log(err.Error())
		return
	}

	d, err := GetInstance().GetMap("test_hash")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(d)

	s, err := GetInstance().GetMapKeyValue("test_hash", "3")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(s)
}

func TestRedisClient_Set(t *testing.T) {

	GetInstance().Init("192.168.1.222:6379", 0)

	n, err := GetInstance().SAddString("test_set", "1", "2", "3")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(n)
}

func TestRedisClient_ZSet(t *testing.T) {

	GetInstance().Init("192.168.1.222:6379", 0)

	redKey := "test_zset"

	n, err := GetInstance().ZAdd(redKey, "3", "c", 4, "d")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(n)

	list, err := GetInstance().ZRange(redKey, 0, -1, "WITHSCORES")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(list)

	n, err = GetInstance().ZScore(redKey, "a")
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(n)

	n, err = GetInstance().ZCard(redKey)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log(n)

	err = GetInstance().ZRem(redKey, "1")
	if err != nil {
		t.Log(err.Error())
		return
	}
}
