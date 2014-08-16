package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	"github.com/fzzy/radix/redis"
	log "github.com/grooveshark/golib/gslog"
	"strings"

	"github.com/mediocregopher/breadis/config"
)

var sentinelClientCh = make(chan *sentinel.Client)

func init() {
	switch strings.ToLower(config.Mode) {
	case "single":
		singleinit()
	case "multi":
		multiinit()
	default:
		log.Fatalf("Unknown mode: %s", config.Mode)
	}
}

func GetBucket(bucket string) (*redis.Client, error) {
	return (<-sentinelClientCh).GetMaster(bucket)
}

func PutBucket(bucket string, conn *redis.Client) {
	(<-sentinelClientCh).PutMaster(bucket, conn)
}
