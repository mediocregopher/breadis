package config

import (
	"github.com/mediocregopher/flagconfig"
	"log"
)

var (
	ListenAddr string

	LocatorName   string
	LocatorSet    string
	LocatorPrefix string

	SentinelAddr string
	Buckets      []string
	PoolSize     int
)

func init() {
	fc := flagconfig.New("breadis")
	fc.StrParam(
		"listen-addr",
		"Address breadis will listen for client connections on",
		":36379",
	)
	fc.StrParam(
		"locator-master-name",
		"Name of the master to use as a locator, to be found in the sentinel",
		"locator",
	)
	fc.StrParam(
		"locator-set-name",
		"Name of the redis SET to use on the locator",
		"members",
	)
	fc.StrParam(
		"locator-prefix",
		"Prefix to give all location keys on the locator node",
		"loc:",
	)
	fc.StrParam(
		"sentinel-addr",
		"Address redis sentinel is listening on",
		"127.0.0.1:26379",
	)
	fc.StrParams(
		"bucket-name",
		"Names of the buckets in sentinel to seed the pool with on breadis startup. Leave unspecified to always do it manually, specify multiple times for multiple buckets",
	)
	fc.IntParam(
		"conn-pool-size",
		"Number of connections per bucket/locator to use as an initial pool size",
		10,
	)
	if err := fc.Parse(); err != nil {
		log.Fatal(err)
	}
	ListenAddr = fc.GetStr("listen-addr")
	LocatorName = fc.GetStr("locator-master-name")
	LocatorSet = fc.GetStr("locator-set-name")
	LocatorPrefix = fc.GetStr("locator-prefix")
	SentinelAddr = fc.GetStr("sentinel-addr")
	Buckets = fc.GetStrs("bucket-name")
	PoolSize = fc.GetInt("conn-pool-size")
}
