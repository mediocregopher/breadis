package config

import (
	"github.com/mediocregopher/flagconfig"
	log "github.com/grooveshark/golib/gslog"
	"strings"
)

var (
	ListenAddr   string
	SentinelAddr string
	PoolSize     int
	LogLevel     string
	Mode         string

	SingleMode   bool // populated by config, not flagconfig
	SingleBucket string

	LocatorName   string
	LocatorSet    string
	LocatorPrefix string
	Buckets       []string
	CacheSize     int
)

func init() {
	fc := flagconfig.New("breadis")
	fc.StrParam(
		"listen-addr",
		"Address breadis will listen for client connections on",
		":36379",
	)
	fc.StrParam(
		"sentinel-addr",
		"Address redis sentinel is listening on",
		"127.0.0.1:26379",
	)
	fc.IntParam(
		"conn-pool-size",
		"Number of connections per bucket/locator to use as an initial pool size",
		10,
	)
	fc.StrParam(
		"log-level",
		"Minimum level of severity to log to stderr (debug, info, warn, error, fatal)",
		"info",
	)
	fc.StrParam(
		"mode",
		"Either 'single' (only proxy and handle failover for a single bucket, specified by the single-bucket flag) or 'multi' (proxy and handle failover for an entire cluster, with key sharding handled by a locator redis node)",
		"single",
	)
	fc.StrParam(
		"single-bucket",
		"(Single mode) Name of the master to proxy to. Required if in single mode",
		"",
	)
	fc.StrParam(
		"locator-master-name",
		"(Multi mode) Name of the master to use as a locator, to be found in the sentinel",
		"locator",
	)
	fc.StrParam(
		"locator-set-name",
		"(Multi mode) Name of the redis SET to use on the locator",
		"members",
	)
	fc.StrParam(
		"locator-prefix",
		"(Multi mode) Prefix to give all location keys on the locator node",
		"loc:",
	)
	fc.StrParams(
		"bucket-name",
		"(Multi mode) Names of the buckets in sentinel to seed the pool with on breadis startup. Leave unspecified to always do it manually, specify multiple times for multiple buckets",
	)
	fc.IntParam(
		"cache-size",
		"(Multi mode) Number of keys to keep cached in memory, to reduce round trips to the locator instance. Set to 0 for no cache",
		4096,
	)
	if err := fc.Parse(); err != nil {
		log.Fatalf("FlagConfig.parse(): %s", err)
	}
	ListenAddr = fc.GetStr("listen-addr")
	SentinelAddr = fc.GetStr("sentinel-addr")
	PoolSize = fc.GetInt("conn-pool-size")
	LogLevel = fc.GetStr("log-level")
	Mode = fc.GetStr("mode")
	SingleBucket = fc.GetStr("single-bucket")
	LocatorName = fc.GetStr("locator-master-name")
	LocatorSet = fc.GetStr("locator-set-name")
	LocatorPrefix = fc.GetStr("locator-prefix")
	Buckets = fc.GetStrs("bucket-name")
	CacheSize = fc.GetInt("cache-size")

	// We do this here so that it happens before anything else can have a chance
	// to log anything.
	if err := log.SetMinimumLevel(LogLevel); err != nil {
		log.Fatalf("log.SetMinimumLevel(%s): %s", LogLevel, err)
	}
	log.Info("Log level set to: %s", LogLevel)

	SingleMode = strings.ToLower(Mode) == "single"
}
