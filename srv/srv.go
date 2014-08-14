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
		log.Fatal(err)
	}
	log.Println("Listening")
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Accept: %s", err)
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
	var rconn *redis.Client
	var rm *resp.Message

	// We don't bother testing for errors when writing to the connection because
	// if the connection truly fails then we'll handle the error when we go to
	// read again

	ms, err = m.Array()
	if err != nil || len(ms) < 2 || ms[1].Type != resp.BulkStr {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}

	key, err = ms[1].Str()
	if err != nil {
		resp.WriteArbitrary(conn, errBadCmd)
		return
	}

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

	err = resp.WriteMessage(rconn.Conn, m)
	if err != nil {
		log.Printf("WriteMessage(rconn, m): %s", err)
		resp.WriteArbitrary(conn, errBackend)
		return
	}

	rm, err = resp.ReadMessage(rconn.Conn)
	if err != nil {
		log.Printf("ReadMessage(rconn): %s", err)
		resp.WriteArbitrary(conn, errBackend)
		return
	}

	resp.WriteMessage(conn, rm)
	bak.PutBucket(bucket, rconn)
}
