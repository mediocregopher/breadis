package loc

import (
	"errors"
	"github.com/fzzy/radix/redis"
	"io"
	"strings"

	"github.com/mediocregopher/breadis/bak"
	"github.com/mediocregopher/breadis/config"
)

func BucketForKey(key string) (string, error) {
	if config.SingleMode {
		return config.SingleBucket, nil
	}
	if key[len(key)-1] == '}' {
		if i := strings.Index(key, "{"); i > -1 {
			key = key[i+1 : len(key)-1]
		}
	}

	loc := getFromCache(key)
	if loc != "" {
		return loc, nil
	}

	prefixedKey := config.LocatorPrefix + key
	loc, err := bucketForKeyRaw(prefixedKey)
	if err != nil {
		return "", err
	}
	setInCache(key, loc)
	return loc, nil
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
	} else if r.Type == redis.BulkReply {
		return r.Str()
	}

	// we only get here if the key doesn't have a bucket assigned to it yet
	r = conn.Cmd("SRANDMEMBER", config.LocatorSet)
	if r.Type == redis.NilReply {
		return "", errors.New("No buckets in pool")
	}
	bucket, err = r.Str()
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
