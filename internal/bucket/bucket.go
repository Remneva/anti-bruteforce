package bucket

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

var ctx = context.Background()

type Client struct {
	redis.Client
	limitLogin    int64
	limitPassword int64
	limitIp       int64
}

func NewClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	pong, err := client.Ping(ctx).Result()
	fmt.Println(pong, err)
	return client
}

//new bucket creating
func (c *Client) checkLoginLimit(login string) (error, bool) {
	currentTime := time.Now()
	timeInterval := currentTime.Truncate(1 * time.Minute)
	fmt.Println("timeInterval: ", timeInterval)

	key := fmt.Sprintf("key_%s_%s", login, timeInterval)
	fmt.Println("key: ", key)
	incr := c.Incr(ctx, key)
	count, _ := incr.Result()
	fmt.Println("current count: ", count)
	err := c.Expire(ctx, key, 60*time.Second)
	if err != nil {
		return fmt.Errorf("set TTL error: %s", err), false
	}
	if count > c.limitLogin {
		return nil, false
	}
	return nil, true
}

func (c *Client) checkPasswordLimit(password string) (error, bool) {
	currentTime := time.Now()
	timeInterval := currentTime.Truncate(1 * time.Minute)
	fmt.Println("timeInterval: ", timeInterval)

	key := fmt.Sprintf("key_%s_%s", password, timeInterval)
	fmt.Println("key: ", key)
	incr := c.Incr(ctx, key)
	count, _ := incr.Result()
	fmt.Println("count: ", count)
	err := c.Expire(ctx, key, 60*time.Second)
	if err != nil {
		return fmt.Errorf("set TTL error: %s", err), false
	}
	if count > c.limitPassword {
		return nil, false
	}
	return nil, true
}

func (c *Client) checkIpLimit(ip string) (error, bool) {
	currentTime := time.Now()
	timeInterval := currentTime.Truncate(1 * time.Minute)
	fmt.Println("timeInterval: ", timeInterval)

	key := fmt.Sprintf("key_%s_%s", ip, timeInterval)
	fmt.Println("key: ", key)
	incr := c.Incr(ctx, key)
	count, _ := incr.Result()
	fmt.Println("count: ", count)
	err := c.Expire(ctx, key, 60*time.Second)
	if err != nil {
		return fmt.Errorf("set TTL error: %s", err), false
	}
	if count > c.limitIp {
		return nil, false
	}
	return nil, true
}
