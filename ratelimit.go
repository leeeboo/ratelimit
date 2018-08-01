package ratelimit

import (
	"fmt"

	"github.com/caarlos0/env"
	"github.com/garyburd/redigo/redis"
)

var redisClient redis.Conn

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

	redisClient = newRedisClient(cfg)
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

	reply, err := redisClient.Do("CL.THROTTLE", this.Key, this.Capacity, this.CountPerPeriod, this.Period, quota)

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

func newRedisClient(cfg Config) redis.Conn {

	client, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))

	if err != nil {
		panic(err)
	}

	if _, err := client.Do("SELECT", cfg.Db); err != nil {
		client.Close()
		panic(err)
	}

	return client
}
