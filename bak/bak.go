package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	"github.com/fzzy/radix/redis"
	"log"

	"github.com/mediocregopher/breadis/config"
)

var (
	sentinelConn   *redis.Client
	sentinelClient *sentinel.Client
)

func init() {
	var err error
	var locConn *redis.Client

	sentinelConn, err = redis.Dial("tcp", config.SentinelAddr)
	if err != nil {
		log.Fatal(err)
	}

	initialBuckets := []string{config.LocatorName}
	initialBuckets = append(initialBuckets, config.Buckets...)

	sentinelClient, err = sentinel.NewClient(
		"tcp",
		config.SentinelAddr,
		10,
		initialBuckets...,
	)
	if err != nil {
		log.Fatal("sentinel.NewClient", err)
	}

	bis := make([]interface{}, 0, len(initialBuckets)+1)
	bis = append(bis, config.LocatorSet)
	for i := range bis {
		bis = append(bis, initialBuckets[i])
	}

	if locConn, err = sentinelClient.GetMaster(config.LocatorName); err != nil {
		log.Fatal("sentinelClient.GetMaster", err)
	}

	_, err = locConn.Cmd("SADD", bis...).Int()
	if err != nil {
		log.Fatal(err)
	}
}

func GetBucket(bucket string) (*redis.Client, error) {
	return sentinelClient.GetMaster(bucket)
}

func PutBucket(bucket string, conn *redis.Client) {
	sentinelClient.PutMaster(bucket, conn)
}
