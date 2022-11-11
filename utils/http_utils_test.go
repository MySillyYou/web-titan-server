package utils

import (
	log "common/alog"
	redisLib "common/redis"
	"testing"
)

func TestGet(t *testing.T) {
	redis := &redisLib.RedisUtil{}
	redis.Init("127.0.0.1:6379", 1)
	log.Info(WanshuThreeKeyElements("夏笑声", "17620163567", "360681199401136130", redis))
	return
}
