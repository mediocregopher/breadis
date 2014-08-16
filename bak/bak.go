package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	"github.com/fzzy/radix/redis"
	log "github.com/grooveshark/golib/gslog"
	"reflect"
	"time"

	"github.com/mediocregopher/breadis/config"
)

var sentinelClientCh = make(chan *sentinel.Client)

func init() {
	go forit()

	if len(config.Buckets) == 0 {
		return
	}

	log.Infof("Adding buckets to pool: %v", config.Buckets)

	bis := make([]interface{}, 0, len(config.Buckets)+1)
	bis = append(bis, config.LocatorSet)
	for i := range config.Buckets {
		bis = append(bis, config.Buckets[i])
	}

	locConn, err := (<-sentinelClientCh).GetMaster(config.LocatorName)
	if err != nil {
		log.Fatal("sentinelClient.GetMaster", err)
	}

	_, err = locConn.Cmd("SADD", bis...).Int()
	if err != nil {
		log.Fatalf("SADD buckets: %s", err)
	}
}

func forit() {
	sentinelConn, err := redis.Dial("tcp", config.SentinelAddr)
	if err != nil {
		log.Fatalf("Dial sentinelConn at %s: %s", config.SentinelAddr, err)
	}

	allBuckets, err := getBucketList(sentinelConn)
	if err != nil {
		log.Fatalf("getBucketList: %s", err)
	}

	bucketHash := map[string]bool{}
	for i := range allBuckets {
		bucketHash[allBuckets[i]] = true
	}

	sentinelClient, err := sentinel.NewClient(
		"tcp",
		config.SentinelAddr,
		10,
		allBuckets...,
	)
	if err != nil {
		log.Fatal("sentinel.NewClient", err)
	}
	log.Infof("Connected to sentinel buckets: %v", allBuckets)

	tick := time.Tick(10 * time.Second)
	for {
		select {
		case sentinelClientCh <- sentinelClient:
		case <-tick:
			allBuckets, err = getBucketList(sentinelConn)
			if err != nil {
				log.Fatalf("getBucketList: %s", err)
			}
			newBucketHash := map[string]bool{}
			for i := range allBuckets {
				newBucketHash[allBuckets[i]] = true
			}
			if !reflect.DeepEqual(bucketHash, newBucketHash) {
				go forit()
				sentinelConn.Close()
				sentinelClient.Close()
				return
			}
		}
	}
}

func getBucketList(sentinelConn *redis.Client) ([]string, error) {
	r := sentinelConn.Cmd("SENTINEL", "MASTERS")
	if r.Err != nil {
		return nil, r.Err
	}
	allBuckets := make([]string, len(r.Elems))
	for i := range r.Elems {
		masterInfo, err := r.Elems[i].Hash()
		if err != nil {
			return nil, err
		}
		allBuckets[i] = masterInfo["name"]
	}
	return allBuckets, nil
}

func GetBucket(bucket string) (*redis.Client, error) {
	return (<-sentinelClientCh).GetMaster(bucket)
}

func PutBucket(bucket string, conn *redis.Client) {
	(<-sentinelClientCh).PutMaster(bucket, conn)
}
