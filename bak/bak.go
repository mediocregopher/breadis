package bak

import (
	"github.com/fzzy/radix/extra/sentinel"
	"github.com/fzzy/radix/redis"
	"github.com/fzzy/radix/redis/resp"
	log "github.com/grooveshark/golib/gslog"
	"strings"

	"github.com/mediocregopher/breadis/config"
)

type sentinelReq struct {
	m  *resp.Message
	ch chan *resp.Message
}

var (
	sentinelReqCh    = make(chan *sentinelReq)
	sentinelClientCh = make(chan *sentinel.Client)
)

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

func sentinelDirect(conn *redis.Client, r *sentinelReq) {
	if err := resp.WriteMessage(conn.Conn, r.m); err != nil {
		log.Fatalf("sentinelConn write: %s", err)
	}
	rm, err := resp.ReadMessage(conn.Conn)
	if err != nil {
		log.Fatalf("sentinelConn read: %s", err)
	}
	r.ch <- rm
}

func SentinelDirectCmd(m *resp.Message) *resp.Message {
	req := sentinelReq{m, make(chan *resp.Message)}
	sentinelReqCh <- &req
	return <-req.ch
}
