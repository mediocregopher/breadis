package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	"github.com/fzzy/radix/redis"
	//"github.com/fzzy/radix/redis/resp"
	"io"
	"log"
	"strings"
)

var (
	sentinelConn   *redis.Client
	sentinelClient *sentinel.Client
)

// <CONFIG STUFF>
var initialBuckets = []string{
	"locator",
	"bucket0",
}
const (
	SENTINEL_ADDR = "localhost:26379"
	LOCATOR_NAME = "locator"
	LOCATOR_SET = "members"
	LOCATOR_PREFIX = "loc:"
)
// </CONFIG STUFF>

func init() {
	var err error
	var locConn *redis.Client

	sentinelConn, err = redis.Dial("tcp", SENTINEL_ADDR)
	if err != nil {
		log.Fatal(err)
	}

	sentinelClient, err = sentinel.NewClient(
		"tcp",
		SENTINEL_ADDR,
		10,
		initialBuckets...
	)
	if err != nil {
		log.Fatal("sentinel.NewClient", err)
	}

	bis := make([]interface{}, 0, len(initialBuckets)+1)
	bis = append(bis, LOCATOR_SET)
	for i := range bis {
		bis = append(bis, initialBuckets[i])
	}

	if locConn, err = sentinelClient.GetMaster(LOCATOR_NAME); err != nil {
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

// TODO locator stuff should be broken out into its own package

func BucketForKey(key string) (string, error) {
	if key[len(key)-1] == '}' {
		if i := strings.Index(key, "{"); i > -1 {
			key = key[i+1:len(key)-1]
		}
	}
	key = LOCATOR_PREFIX + key

	return bucketForKeyRaw(key)
}

func tryReturnConn(bucket string, conn *redis.Client, err *error) {
	if *err == io.EOF {
		return
	}
	PutBucket(bucket, conn)
}

func bucketForKeyRaw(key string) (string, error) {
	var conn *redis.Client
	var err error
	var bucket string
	conn, err = GetBucket(LOCATOR_NAME)
	if err != nil {
		return "", err
	}
	defer tryReturnConn(LOCATOR_NAME, conn, &err)

	r := conn.Cmd("GET", key)
	if r.Type == redis.ErrorReply {
		return "", r.Err
	} else if r.Type == redis.BulkReply  {
		return r.Str()
	}

	// we only get here if the key doesn't have a bucket assigned to it yet
	bucket, err = conn.Cmd("SRANDMEMBER", LOCATOR_SET).Str()
	if err != nil {
		return "", err
	}

	var wasSet int
	wasSet, err = conn.Cmd("SETNX", key, bucket).Int()
	if err != nil {
		return "", err
	} else if wasSet == 0 {
		// Another process set the key before we could. We go back to square one
		// in order to fetch it for real this time
		return bucketForKeyRaw(key)
	}

	return bucket, nil
}


