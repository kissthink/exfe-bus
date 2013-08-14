package qutimer

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/googollee/go-assert"
	"testing"
)

func TestTaskCompleteNumber(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number, []interface{}{1, 2, 3})
	err := task.Complete()
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
}

func TestTaskCompleteNumberAdd1(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number+1, []interface{}{1, 2, 3})
	err := task.Complete()
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)

	conn.Do("DEL", queue)
	conn.Do("DEL", queueOverwrite)
	conn.Do("DEL", queueLocker)
	conn.Do("DEL", queueData)
	conn.Do("DEL", timer)
}

func TestTaskCompleteNumberSub1(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number-1, []interface{}{1, 2, 3})
	err := task.Complete()
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("LLEN", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("HLEN", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 2)
	ontime, err := redis.Int(conn.Do("ZSCORE", timer, key))
	assert.Equal(t, err, nil)
	assert.Equal(t, ontime, 12345)
}

func TestTaskReleaseNumber(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number, []interface{}{1, 2, 3})
	err := task.Release(23456)
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("LLEN", queue))
	assert.Equal(t, err, nil)
	assert.MustEqual(t, l, 10)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("HLEN", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 2)
	ontime, err := redis.Int(conn.Do("ZSCORE", timer, key))
	assert.Equal(t, err, nil)
	assert.Equal(t, ontime, 23456)
}

func TestTaskReleaseNumberSub1(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number-1, []interface{}{1, 2, 3})
	err := task.Release(23456)
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("LLEN", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 10)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("HLEN", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 2)
	ontime, err := redis.Int(conn.Do("ZSCORE", timer, key))
	assert.Equal(t, err, nil)
	assert.Equal(t, ontime, 23456)
}

func TestTaskReleaseNumberAdd1(t *testing.T) {
	conn := redisPool.Get()
	defer conn.Close()
	defer clearQueue()

	for i := 0; i < number; i++ {
		conn.Do("RPUSH", queue, fmt.Sprintf("%d", i))
	}
	conn.Do("SET", queueOverwrite, "1")
	conn.Do("SET", queueLocker, "1234.5678")
	conn.Do("HSET", queueData, "data_key", "data_value")
	conn.Do("ZADD", timer, 12345, key)
	conn.Do("ZADD", timer, 12346, "other_key")

	task := newTask(redisPool, prefix, key, number+1, []interface{}{1, 2, 3})
	err := task.Release(23456)
	assert.Equal(t, err, nil)
	exist, err := redis.Bool(conn.Do("EXISTS", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err := redis.Int(conn.Do("LLEN", queue))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 10)
	exist, err = redis.Bool(conn.Do("EXISTS", queueOverwrite))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	exist, err = redis.Bool(conn.Do("EXISTS", queueLocker))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, false)
	exist, err = redis.Bool(conn.Do("EXISTS", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("HLEN", queueData))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 1)
	exist, err = redis.Bool(conn.Do("EXISTS", timer))
	assert.Equal(t, err, nil)
	assert.Equal(t, exist, true)
	l, err = redis.Int(conn.Do("ZCOUNT", timer, "-INF", "+INF"))
	assert.Equal(t, err, nil)
	assert.Equal(t, l, 2)
	ontime, err := redis.Int(conn.Do("ZSCORE", timer, key))
	assert.Equal(t, err, nil)
	assert.Equal(t, ontime, 23456)

}
