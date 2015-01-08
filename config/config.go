package config

import (
	"log"

	"github.com/mediocregopher/flagconfig"
)

var (
	ListenAddr string
	RedisAddrs []string
)

func init() {
	fc := flagconfig.New("breadis")
	fc.StrParam(
		"listen-addr",
		"Address breadis will listen for client connections on",
		":36379",
	)
	fc.StrParams(
		"redis-addr",
		"Address of a member of the redis cluster. Can be specified multiple times. Will go through each individually until a connection is successfully made",
		"127.0.0.1:6379",
	)

	if err := fc.Parse(); err != nil {
		log.Fatal(err)
	}

	ListenAddr = fc.GetStr("listen-addr")
	RedisAddrs = fc.GetStrs("redis-addr")
}
