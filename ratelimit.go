package ratelimit

import (
	"fmt"
	"time"

	"github.com/caarlos0/env"
	"github.com/garyburd/redigo/redis"
)

var redisPool *redis.Pool

type Config struct {
	Host string `env:"RATELIMIT_REDIS_HOST"`
	Port int    `env:"RATELIMIT_REDIS_PORT"`
	Db   int    `env:"RATELIMIT_REDIS_DB"`
}

func init() {

	cfg := Config{}
	err := env.Parse(&cfg)

	if err != nil {
		panic(err)
	}

	if cfg.Host == "" || cfg.Port == 0 {
		panic("Need ENV `RATELIMIT_REDIS_HOST`, `RATELIMIT_REDIS_PORT`")
	}

	redisPool = newRedisPool(cfg)
}

type Bucket struct {
	Key            string
	Capacity       int
	CountPerPeriod int
	Period         int
}

func NewBucket(key string, capacity int, countPerPeriod int, period int) *Bucket {

	bucket := new(Bucket)
	bucket.Key = key
	bucket.Capacity = capacity
	bucket.CountPerPeriod = countPerPeriod
	bucket.Period = period

	return bucket
}

type Result struct {
	Allow        bool
	Capacity     int
	LeftQuota    int
	WaitSecond   int
	RefullSecond int
}

func (this *Bucket) Take(quota int) (*Result, error) {

	conn := redisPool.Get()

	reply, err := conn.Do("CL.THROTTLE", this.Key, this.Capacity, this.CountPerPeriod, this.Period, quota)

	if err != nil {
		return nil, err
	}

	keys, err := redis.Ints(reply, nil)

	if err != nil {
		return nil, err
	}

	result := new(Result)

	if keys[0] == 0 {
		result.Allow = true
	} else {
		result.Allow = false
	}
	result.Capacity = keys[1]
	result.LeftQuota = keys[2]
	result.WaitSecond = keys[3]
	result.RefullSecond = keys[4]

	return result, nil
}

func newRedisPool(cfg Config) *redis.Pool {

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
			if err != nil {
				return nil, err
			}
			if _, err := c.Do("SELECT", cfg.Db); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
