package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	log "github.com/grooveshark/golib/gslog"

	"github.com/mediocregopher/breadis/config"
)

func singleinit() {
	if config.SingleBucket == "" {
		log.Fatal("Must specify single-bucket when in single mode")
	}

	go func() {
		sentinelClient, err := sentinel.NewClient(
			"tcp",
			config.SentinelAddr,
			config.PoolSize,
			config.SingleBucket,
		)
		if err != nil {
			log.Fatalf("sentinel.NewClient: %s", err)
		}

		for {
			sentinelClientCh <- sentinelClient
		}
	}()
}
