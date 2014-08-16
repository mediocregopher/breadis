package loc

import (
	"github.com/fzzy/radix/redis"
	"io"
	"strings"

	"github.com/mediocregopher/breadis/bak"
	"github.com/mediocregopher/breadis/config"
)

func BucketForKey(key string) (string, error) {
	if key[len(key)-1] == '}' {
		if i := strings.Index(key, "{"); i > -1 {
			key = key[i+1:len(key)-1]
		}
	}
	key = config.LocatorPrefix + key

	return bucketForKeyRaw(key)
}

func tryReturnConn(bucket string, conn *redis.Client, err *error) {
	if *err == io.EOF {
		return
	}
	bak.PutBucket(bucket, conn)
}

func bucketForKeyRaw(key string) (string, error) {
	var conn *redis.Client
	var err error
	var bucket string
	conn, err = bak.GetBucket(config.LocatorName)
	if err != nil {
		return "", err
	}
	defer tryReturnConn(config.LocatorName, conn, &err)

	r := conn.Cmd("GET", key)
	if r.Type == redis.ErrorReply {
		return "", r.Err
	} else if r.Type == redis.BulkReply  {
		return r.Str()
	}

	// we only get here if the key doesn't have a bucket assigned to it yet
	bucket, err = conn.Cmd("SRANDMEMBER", config.LocatorName).Str()
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
