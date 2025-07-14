package redis

import (
	"context"
	"github.com/crazy-choose/go/meta"
	"github.com/crazy-choose/helper/log"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	cfg        map[string]meta.Redis
	rcm        map[string]*redis.Client // rds 连接map
	rc         *redis.Client            // 默认的链接
	CTX        context.Context          // rds 上下文
	ExpireTime = 86400 * time.Second
)

func init() {
	cfg = make(map[string]meta.Redis)
	rcm = make(map[string]*redis.Client)
}

func Init(key string, opt meta.Redis) bool {
	if cfg == nil {
		cfg = make(map[string]meta.Redis)
	}
	if rcm == nil {
		rcm = make(map[string]*redis.Client)
	} else {
		r, ok := rcm[key]
		if ok {
			r.Close()
		}
	}

	_rc_ := redis.NewClient(&redis.Options{
		Addr:         opt.Host,
		Password:     opt.AuthKey,  // no password set
		DB:           opt.DbIndex,  // use default DB
		PoolSize:     opt.PoolSize, // 连接池最大socket连接数，默认为4倍CPU数， 4 * runtime.NumCPU
		MinIdleConns: 2,            //在启动阶段创建指定数量的Idle连接，并长期维持idle状态的连接数不少于指定数量；
		//超时
		DialTimeout: opt.DialTimeout * time.Second, //连接建立超时时间，默认5秒。
		PoolTimeout: opt.PoolTimeout * time.Second, //当所有连接都处在繁忙状态时，客户端等待可用连接的最大等待时长，默认为读超时+1秒
		//IdleTimeout:  opt.IdleTimeout * time.Second,
		ReadTimeout:  opt.ReadTimeout * time.Second,  //读超时，默认3秒， -1表示取消读超时
		WriteTimeout: opt.WriteTimeout * time.Second, //写超时，默认等于读超时
	})
	CTX = context.Background()
	pong, err := _rc_.Ping(CTX).Result()
	if err != nil {
		log.Error("rdb.Pong(%+v), err%+v, Redis(%s) Init->Success", pong, err, key)
		return false
	}

	rcm[key] = _rc_
	cfg[key] = opt
	if opt.Default {
		rc = _rc_
	}
	log.Info("Redis(%s) Init->Success", key)
	return true
}

func Impl(opt string) *redis.Client {
	if opt == "" {
		opt = "HFT"
	}
	r, ok := rcm[opt]
	if !ok {
		if c, _ok := cfg[opt]; !_ok {
			return nil
		} else {
			if Init(opt, c) {
				r = rcm[opt]
			}
		}
	}
	return r
}

func Get(opt, key string) *redis.StringCmd {
	_rc_ := rcm[opt]
	return _rc_.Get(CTX, key)
}

func DoGet(opt, key string) (interface{}, error) {
	_rc_ := rcm[opt]
	return _rc_.Do(CTX, "GET", key).Result()
}

func Set(opt, key string, value interface{}, expire time.Duration) *redis.StatusCmd {
	_rc_ := rcm[opt]
	return _rc_.Set(CTX, key, value, expire)
}

func MGet(opt string, keys ...string) *redis.SliceCmd {
	_rc_ := rcm[opt]
	return _rc_.MGet(CTX, keys...)
}

func MSet(opt string, values ...interface{}) *redis.StatusCmd {
	_rc_ := rcm[opt]
	return _rc_.MSet(CTX, values...)
}

func Publish(opt, channel string, message interface{}) *redis.IntCmd {
	_rc_ := rcm[opt]
	return _rc_.Publish(CTX, channel, message)
}

func Subscribe(opt string, channel ...string) *redis.PubSub {
	_rc_ := rcm[opt]
	return _rc_.Subscribe(CTX, channel...)
}

func PSubscribe(opt string, handler func(ps *redis.PubSub), channel ...string) error {
	_rc_ := rcm[opt]
	handler(_rc_.PSubscribe(CTX, channel...))
	return nil
}

func SMembers(opt, key string) *redis.StringSliceCmd {
	_rc_ := rcm[opt]
	return _rc_.SMembers(CTX, key)
}

func ZRange(opt, key string, start, stop int64) *redis.StringSliceCmd {
	_rc_ := rcm[opt]
	return _rc_.ZRange(CTX, key, start, stop)
}

func HSet(opt, key, field string, value interface{}) *redis.IntCmd {
	_rc_ := rcm[opt]
	return _rc_.HSet(CTX, key, field, value)
}

func HGet(opt, key, field string) *redis.StringCmd {
	_rc_ := rcm[opt]
	return _rc_.HGet(CTX, key, field)
}

func HGetAll(opt, key string) *redis.MapStringStringCmd {
	_rc_ := rcm[opt]
	return _rc_.HGetAll(CTX, key)
}

func LIndex(opt, key string, idx int64) *redis.StringCmd {
	_rc_ := rcm[opt]
	return _rc_.LIndex(CTX, key, idx)
}

func LRange(opt, key string, start, stop int64) *redis.StringSliceCmd {
	_rc_ := rcm[opt]
	return _rc_.LRange(CTX, key, start, stop)
}

func LPush(opt, key string, val interface{}) *redis.IntCmd {
	_rc_ := rcm[opt]
	return _rc_.LPush(CTX, key, val)
}

func RPush(opt, key string, val interface{}) *redis.IntCmd {
	_rc_ := rcm[opt]
	return _rc_.RPush(CTX, key, val)
}

func DelKey(opt string, key ...string) *redis.IntCmd {
	_rc_ := rcm[opt]
	return _rc_.Del(CTX, key...)
}

func SScan(opt, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	_rc_ := rcm[opt]
	return _rc_.SScan(CTX, key, cursor, match, count)
}

func ZScan(opt, key string, cursor uint64, match string, count int64) *redis.ScanCmd {
	_rc_ := rcm[opt]
	return _rc_.ZScan(CTX, key, cursor, match, count)
}

func Flush() {
	for _, r := range rcm {
		r.FlushDB(CTX) // 清除数据
	}
}
