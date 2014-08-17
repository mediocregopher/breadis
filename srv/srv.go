package srv

import (
	"errors"
	"github.com/fzzy/radix/redis"
	"github.com/fzzy/radix/redis/resp"
	"io"
	log "github.com/grooveshark/golib/gslog"
	"net"
	"strings"
	"time"
	
	"github.com/mediocregopher/breadis/bak"
	"github.com/mediocregopher/breadis/bak/loc"
	"github.com/mediocregopher/breadis/config"
)

// errors
var (
	errUnkCmd  = errors.New("ERR unknown command")
	errBadCmd  = errors.New("ERR bad command")
	errBackend = errors.New("ERR backend error")
)


func Listen() {
	ln, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalf("Listening on %s: %s", config.ListenAddr,  err)
	}
	log.Infof("Listening on %s", config.ListenAddr)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Warnf("Accept: %s", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Debug("got connection")

	var m *resp.Message
	var err error
	for {
		err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Warn("error setting read deadline")
			conn.Close()
			return
		}

		m, err = resp.ReadMessage(conn)
		if err == io.EOF {
			log.Debug("connection closed")
			return
		} else if t, ok := err.(*net.OpError); ok && t.Timeout() {
			continue;
		}  else if err != nil {
			// If this fails the connection read will fail on the next loop so
			// no need to check here
			resp.WriteArbitrary(conn, errUnkCmd)
			continue
		}

		handleCommand(conn, m)
	}

}

func handleCommand(conn net.Conn, m *resp.Message) {
	var err error
	var ms []*resp.Message
	var cmd, key string
	var rm *resp.Message

	// We don't bother testing for errors when writing to the connection because
	// if the connection truly fails then we'll handle the error when we go to
	// read again

	log.Debug("handleCommand")

	ms, err = m.Array()
	if err != nil || len(ms) < 2 || ms[1].Type != resp.BulkStr {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}
cmd, err = ms[0].Str()
	if err != nil {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}

	key, err = ms[1].Str()
	if err != nil {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}

	if (cmd[0] == 's' || cmd[0] == 'S') && strings.ToLower(cmd) == "sentinel" {
		rm = bak.SentinelDirectCmd(m)
	} else {
		rm, err = bucketCommand(conn, key, m)
	}

	if err != nil {
		resp.WriteArbitrary(conn, err)
	} else {
		resp.WriteMessage(conn, rm)
	}
}

func bucketCommand(
	conn net.Conn, key string, m *resp.Message,
) (
	*resp.Message, error,
) {
	var bucket string
	var err error
	var rconn *redis.Client
	var rm *resp.Message

	bucket, err = loc.BucketForKey(key)
	if err != nil {
		log.Errorf("BucketForKey(%s): %s", key, err)
		return nil, errBackend
	}

	rconn, err = bak.GetBucket(bucket)
	if err != nil {
		log.Errorf("GetBucket(%s): %s", bucket, err)
		return nil, errBackend
	}

	err = resp.WriteMessage(rconn.Conn, m)
	if err != nil {
		log.Errorf("WriteMessage(rconn, m): %s", err)
		return nil, errBackend
	}

	rm, err = resp.ReadMessage(rconn.Conn)
	if err != nil {
		log.Errorf("ReadMessage(rconn): %s", err)
		return nil, errBackend
	}

	bak.PutBucket(bucket, rconn)
	return rm, nil
}
