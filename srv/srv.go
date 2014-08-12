package srv

import (
	"errors"
	"github.com/fzzy/radix/redis"
	"github.com/fzzy/radix/redis/resp"
	"io"
	"log"
	"net"
	"time"
	
	"github.com/mediocregopher/breadis/bak"
)

// errors
var (
	errUnkCmd  = errors.New("ERR unknown command")
	errBadCmd  = errors.New("ERR bad command")
	errBackend = errors.New("ERR backend error")
)


func Listen() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		// handle error
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			// handle error
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Println("got connection")

	var m *resp.Message
	var err error
	for {
		err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Println("error setting read deadline")
			conn.Close()
			return
		}

		m, err = resp.ReadMessage(conn)
		if err == io.EOF {
			log.Println("connection closed")
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
	var key, bucket string
	var cmd []interface{}
	var rconn *redis.Client
	ms, err = m.Array()
	if err != nil || len(ms) < 2 || ms[1].Type != resp.BulkStr {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}

	for i := range ms {
		cmd[i], err = ms[i].Str()
		if err != nil {
			resp.WriteArbitrary(conn, errBadCmd)
			return
		}
	}

	key = cmd[1].(string)

	bucket, err = bak.BucketForKey(key)
	if err != nil {
		log.Printf("BucketForKey(%s): %s", key, err)
		resp.WriteArbitrary(conn, errBackend)
		return
	}

	rconn, err = bak.GetBucket(bucket)
	if err != nil {
		log.Printf("GetBucket(%s): %s", bucket, err)
		resp.WriteArbitrary(conn, errBackend)
		return
	}

	
}

func forwardCommand(
	// TODO Pass in the message itself instead of cmd. We would need to allow
	// for getting the raw connection out of the redis client. Shouldn't be too
	// difficult. DO THIS NOW YOU MUST DO IT THERE IS NO CHOICE
	conn net.Conn, rconn *redis.Client, cmd []interface{},
) (
	*resp.Message, bool,
) {
	var err error
	var r *redis.Reply

	r = rconn.Cmd(cmd[0].(string), cmd[1:]...)
	if r.Type == redis.ErrorReply {
		if r.Err.Error()[:3] == "ERR" {
			return r, true
		} else {
			r.Err = errBackend
		}
	}
}
