package broker

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/googollee/go-logger"
	"github.com/googollee/go-multiplexer"
	"github.com/googollee/godis"
	"model"
	"net"
	"time"
)

type RedisInstance_ struct {
	Conn  net.Conn
	Redis redis.Conn
	log   *logger.SubLogger
}

func (i *RedisInstance_) Ping() error {
	reply, err := redis.String(i.Redis.Do("PING"))
	if reply != "PONG" {
		err = fmt.Errorf("redis not pong.")
	}
	return err
}

func (i *RedisInstance_) Close() error {
	return i.Redis.Close()
}

func (i *RedisInstance_) Error(err error) {
	i.log.Err("%s", err)
}

type RedisPool struct {
	homo   *multiplexer.Homo
	config *model.Config
}

func NewRedisPool(config *model.Config) (*RedisPool, error) {
	if config.Redis.MaxConnections == 0 {
		return nil, fmt.Errorf("config Redis.MaxConnections should not 0!")
	}
	return &RedisPool{
		homo: multiplexer.NewHomo(func() (multiplexer.Instance, error) {
			conn, err := net.DialTimeout("tcp", config.Redis.Netaddr, NetworkTimeout)
			if err != nil {
				return nil, err
			}
			return &RedisInstance_{
				Conn:  conn,
				Redis: redis.NewConn(conn, 0, 0),
				log:   config.Log.SubPrefix("redis"),
			}, nil
		}, config.Redis.MaxConnections, -1, time.Duration(config.Redis.HeartBeatInSecond)*time.Second),
		config: config,
	}, nil
}

func (r *RedisPool) Do(f func(multiplexer.Instance)) error {
	return r.homo.Do(f)
}

func (r *RedisPool) Close() error {
	return r.homo.Close()
}

// old

type RedisInstance struct {
	redis *godis.Client
	log   *logger.SubLogger
}

func (i *RedisInstance) Ping() error {
	_, err := i.redis.Ping()
	return err
}

func (i *RedisInstance) Close() error {
	return i.redis.Quit()
}

func (i *RedisInstance) Error(err error) {
	i.log.Err("%s", err)
}

type RedisMultiplexer struct {
	homo   *multiplexer.Homo
	config *model.Config
}

func NewRedisMultiplexer(config *model.Config) *RedisMultiplexer {
	if config.Redis.MaxConnections == 0 {
		config.Log.Crit("config Redis.MaxConnections should not 0!")
		panic("config Redis.MaxConnections should not 0!")
	}
	return &RedisMultiplexer{
		homo: multiplexer.NewHomo(func() (multiplexer.Instance, error) {
			return &RedisInstance{
				redis: godis.New(fmt.Sprintf("tcp:", config.Redis.Netaddr), config.Redis.Db, config.Redis.Password),
				log:   config.Log.SubPrefix("redis"),
			}, nil
		}, config.Redis.MaxConnections, -1, time.Duration(config.Redis.HeartBeatInSecond)*time.Second),
		config: config,
	}
}

func (m *RedisMultiplexer) Close() error {
	return m.homo.Close()
}

func (m *RedisMultiplexer) Quit() error {
	return nil // no quit
}

func (m *RedisMultiplexer) Get(key string) (elem godis.Elem, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		elem, err = redis.Get(key)
	})
	return
}

func (m *RedisMultiplexer) Set(key string, value interface{}) (err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		err = redis.Set(key, value)
	})
	return
}

func (m *RedisMultiplexer) Incrby(key string, increment int64) (ret int64, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Incrby(key, increment)
	})
	return
}

func (m *RedisMultiplexer) Del(keys ...string) (ret int64, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Del(keys...)
	})
	return
}

func (m *RedisMultiplexer) Rpush(key string, value interface{}) (ret int64, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Rpush(key, value)
	})
	return
}

func (m *RedisMultiplexer) Lrange(key string, start, stop int) (ret *godis.Reply, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Lrange(key, start, stop)
	})
	return
}

func (m *RedisMultiplexer) Zadd(key string, score interface{}, member interface{}) (ret bool, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zadd(key, score, member)
	})
	return
}

func (m *RedisMultiplexer) Zrem(key string, member interface{}) (ret bool, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zrem(key, member)
	})
	return
}

func (m *RedisMultiplexer) Zcount(key string, min float64, max float64) (ret int64, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zcount(key, min, max)
	})
	return
}

func (m *RedisMultiplexer) Zscore(key string, member interface{}) (ret float64, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zscore(key, member)
	})
	return
}

func (m *RedisMultiplexer) Zrange(key string, start int, stop int) (ret *godis.Reply, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zrange(key, start, stop)
	})
	return
}

func (m *RedisMultiplexer) Zrangebyscore(key string, min string, max string, args ...string) (ret *godis.Reply, err error) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret, err = redis.Zrangebyscore(key, min, max, args...)
	})
	return
}

func (m *RedisMultiplexer) NewPipeClient() (ret RedisPipe) {
	m.homo.Do(func(i multiplexer.Instance) {
		redis := i.(*RedisInstance).redis
		ret = godis.NewPipeClientFromClient(redis)
	})
	return
}
