package bak

import (
	"errors"
	"log"

	"github.com/mediocregopher/breadis/config"
	"github.com/mediocregopher/radix.v2/cluster"
	"github.com/mediocregopher/radix.v2/redis"
)

// errors
var (
	errBadCmd = errors.New("ERR bad command")
)

var c *cluster.Cluster

func init() {
	var err error
	for _, addr := range config.RedisAddrs {
		if c, err = cluster.New(addr); err != nil {
			log.Printf("%s: %s", addr, err)
			continue
		}

		return
	}
	log.Fatal("no available redis cluster nodes")
}

func Cmd(m *redis.Resp) *redis.Resp {
	ms, err := m.Array()
	if err != nil || len(ms) < 1 {
		return redis.NewResp(errBadCmd)
	}

	cmd, err := ms[0].Str()
	if err != nil {
		return redis.NewResp(errBadCmd)
	}

	args := make([]interface{}, 0, len(ms[1:]))
	for _, argm := range ms[1:] {
		arg, err := argm.Str()
		if err != nil {
			return redis.NewResp(errBadCmd)
		}
		args = append(args, arg)
	}

	return c.Cmd(cmd, args...)
}
