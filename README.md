The golang ratelimit package based on [redis-cell](https://github.com/brandur/redis-cell)

##Example

```go
import "fmt"
import "github.com/leeeboo/ratelimit"

func test() {

    key := "user-send-message-rate"
    capacity := 10
    countPerPeriod := 5
    period := 60

    b := ratelimit.NewBucket(key, capacity, countPerPeriod, period)

	res, err := b.Take(1)

    if !res.Allow {

        fmt.Fprintln("Request limited. Waiting for %d seconds", res.WaitSecond)
    }
}
```
