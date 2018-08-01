package ratelimit

import (
	"fmt"
	"testing"
	"time"
)

func Test_RedisConn(t *testing.T) {

	defer func() {
		if recover() == nil {
			t.Fatal("Can not run if redis can not connected.")
		}
	}()

	var cfg Config
	cfg.Host = "foo"

	newRedisClient(cfg)
}

func Test_Take(t *testing.T) {

	now := time.Now()
	key := now.Format(time.RFC3339)

	capacity := 3

	b := NewBucket(key, capacity, 5, 60)

	i := 0

	for {
		res, err := b.Take(1)

		if err != nil {
			t.Fatal(err)
		}

		if i < capacity+1 {
			if !res.Allow {
				t.Fatal("Want allow get disallow.")
			}
		}

		if i >= capacity+1 {
			if res.Allow {
				t.Fatal("Want disallow get allow.")
			}
		}

		i++
		if i >= capacity+2 {
			break
		}
	}

	b = NewBucket(fmt.Sprintf("%s-new", key), 0, 1, 5)

	res, err := b.Take(1)

	if err != nil {
		t.Fatal(err)
	}

	if !res.Allow {
		t.Fatal("Want allow get disallow.")
	}

	res, err = b.Take(1)

	if err != nil {
		t.Fatal(err)
	}

	if res.Allow {
		t.Fatal("Want disallow get allow.")
	}

	time.Sleep(time.Duration(res.WaitSecond+1) * time.Second)

	res, err = b.Take(1)

	if err != nil {
		t.Fatal(err)
	}

	if !res.Allow {
		t.Fatal("Want allow get disallow.")
	}
}
