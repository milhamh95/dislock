package main

import (
	"context"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/go-redsync/redsync/v4"
	"github.com/labstack/echo/v4"
	goredislib "github.com/redis/go-redis/v9"
	"net/http"
	"sync"
	"time"

	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
)

type counterMutex struct {
	sync.Mutex
	val int
}

func (c *counterMutex) Add(x int) {
	c.Lock()
	c.val++
	c.Unlock()
}

func main() {
	client := goredislib.NewClient(&goredislib.Options{
		Network: "tcp",
		Addr:    "127.0.0.1:6379",
	})

	locker := redislock.New(client)

	//backoff := redislock.LimitRetry(
	//	redislock.ExponentialBackoff(200*time.Millisecond, 500*time.Millisecond),
	//	3,
	//)

	counterMeter := 0

	e := echo.New()

	// Create a pool with go-redis (or redigo) which is the pool redisync will
	// use while communicating with Redis. This can also be any pool that
	// implements the `redis.Pool` interface.
	clientredsync := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})
	pool := goredis.NewPool(clientredsync) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	// Obtain a new mutex by using the same name for all instances wanting the
	// same lock.
	mutexname := "my-global-mutex"
	mutex := rs.NewMutex(
		mutexname,
		redsync.WithExpiry(10*time.Second),
		redsync.WithRetryDelay(300*time.Millisecond),
	)

	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(http.StatusOK, counterMeter)
	})

	//e.GET("/counter-mutex", func(c echo.Context) error {
	//	var mtx sync.Mutex
	//
	//	mtx.Lock()
	//
	//	// simulate update counter in database
	//	time.Sleep(100 * time.Millisecond)
	//	meterMutex.Add(1)
	//	mtx.Unlock()
	//
	//	fmt.Println("meter mutex lock:", meterMutex.val)
	//	return c.JSON(http.StatusOK, meterMutex.val)
	//})

	e.GET("/counter2", func(c echo.Context) error {
		// Obtain a lock for our given mutex. After this is successful, no one else
		// can obtain the same lock (the same mutex name) until we unlock it.

		err := mutex.Lock()
		if err != nil {
			return c.JSON(http.StatusTooManyRequests, "failed get lock")
		}

		time.Sleep(100 * time.Millisecond)
		counterMeter++

		// Release the lock so other processes or threads can obtain a lock.
		defer mutex.Unlock()

		return c.JSON(http.StatusOK, counterMeter)
	})

	e.GET("/counter", func(c echo.Context) error {
		//ctx := c.Request().Context()
		//lock, err := locker.Obtain(ctx, "counter-lock", 1*time.Second, &redislock.Options{
		//	RetryStrategy: backoff,
		//})
		ctx := context.Background()

		// Try to obtain lock.
		lock, err := locker.Obtain(ctx, "my-key", 2*time.Second, nil)
		if err == redislock.ErrNotObtained {
			return c.JSON(http.StatusTooManyRequests, "too many request")
		} else if err != nil {
			return c.JSON(http.StatusInternalServerError, "internal error")
		}

		// Don't forget to defer Release.
		defer lock.Release(ctx)
		fmt.Println("I have a lock!")

		// Sleep and check the remaining TTL.
		time.Sleep(100 * time.Millisecond)
		counterMeter += 1

		return c.JSON(http.StatusOK, counterMeter)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
