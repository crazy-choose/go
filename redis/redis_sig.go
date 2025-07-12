package redis

import (
	"context"
	"fmt"
	"github.com/crazy-choose/go/log"
	"github.com/crazy-choose/go/meta"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	cfgSig meta.Redis
	rcSig  *redis.Client   // rds 连接
	CTXSig context.Context // rds 上下文
)

func InitSig(opt meta.Redis) bool {
	cfgSig = opt
	rcSig = redis.NewClient(&redis.Options{
		Addr:         cfgSig.Host,
		Password:     cfgSig.AuthKey,  // no password set
		DB:           cfgSig.DbIndex,  // use default DB
		PoolSize:     cfgSig.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 2,               //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；
		//超时
		DialTimeout: cfgSig.DialTimeout * time.Second, //连接建立超时时间，默认5秒。
		PoolTimeout: cfgSig.PoolTimeout * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒
		//IdleTimeout:  cfg.IdleTimeout * time.Second,
		ReadTimeout:  cfgSig.ReadTimeout * time.Second,  //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: cfgSig.WriteTimeout * time.Second, //写超时，默认等于读超时
	})
	CTXSig = context.Background()
	pong, err := rcSig.Ping(CTXSig).Result()
	if err != nil {
		log.Error(fmt.Sprintf("rdb.Pong(%+v), err%+v", pong, err))
		return false
	}

	log.Info("Redis(idx:%d) InitSig Success", opt.DbIndex)
	return true
}

func ImplSig() *redis.Client {
	if rcSig == nil {
		InitSig(cfgSig)
	}
	return rcSig
}

func SigGet(key string) *redis.StringCmd {
	return rcSig.Get(CTXSig, key)
}

func SigDoGet(key string) (interface{}, error) {
	return rcSig.Do(CTXSig, "GET", key).Result()
}

func SigSet(key string, value interface{}, expire time.Duration) *redis.StatusCmd {
	return rcSig.Set(CTXSig, key, value, expire)
}

func SigMGet(keys ...string) *redis.SliceCmd {
	return rcSig.MGet(CTXSig, keys...)
}

func SigMSet(values ...interface{}) *redis.StatusCmd {
	return rcSig.MSet(CTXSig, values...)
}

func SigPublish(channel string, message interface{}) *redis.IntCmd {
	return rcSig.Publish(CTXSig, channel, message)
}

func SigSubscribe(channel ...string) *redis.PubSub {
	return rcSig.Subscribe(CTXSig, channel...)
}

func SigPSubscribe(handler func(ps *redis.PubSub), channel ...string) error {
	handler(rcSig.PSubscribe(CTXSig, channel...))
	return nil
}

func SigSMembers(key string) *redis.StringSliceCmd {
	return rcSig.SMembers(CTXSig, key)
}

func SigZRange(key string, start, stop int64) *redis.StringSliceCmd {
	return rcSig.ZRange(CTXSig, key, start, stop)
}

func SigHSet(opt, key, field string, value interface{}) *redis.IntCmd {
	return rcSig.HSet(CTX, key, field, value)
}

func SigHGet(opt, key, field string) *redis.StringCmd {
	return rcSig.HGet(CTX, key, field)
}

func SigHGetAll(key string) *redis.MapStringStringCmd {
	return rcSig.HGetAll(CTXSig, key)
}

func SigLIndex(key string, idx int64) *redis.StringCmd {
	return rcSig.LIndex(CTXSig, key, idx)
}

func SigLRange(key string, start, stop int64) *redis.StringSliceCmd {
	return rcSig.LRange(CTXSig, key, start, stop)
}

func SigLPush(key string, val interface{}) *redis.IntCmd {
	return rcSig.LPush(CTXSig, key, val)
}

func SigRPush(key string, val interface{}) *redis.IntCmd {
	return rcSig.RPush(CTXSig, key, val)
}

func SigDelKey(key ...string) *redis.IntCmd {
	return rcSig.Del(CTXSig, key...)
}

func SigSScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	return rcSig.SScan(CTXSig, key, cursor, match, count)
}

func SigZScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	return rcSig.ZScan(CTXSig, key, cursor, match, count)
}

func SigFlush() {
	rcSig.FlushDB(CTXSig) // 清除数据
}
