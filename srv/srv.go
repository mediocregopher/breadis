package srv

import (
	"errors"
	"github.com/fzzy/radix/redis/resp"
	"io"
	"log"
	"net"
	"time"
)

// errors
var (
	errUnkCmd = errors.New("ERR unknown command")
	errBadCmd = errors.New("ERR bad command")
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
	var ms []*resp.Message
	var err error
	for {
		err = conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Println("error setting read deadline")
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

		ms, err = m.Array()
		if err != nil || len(ms) < 2 || ms[1].Type != resp.BulkStr {
			resp.WriteArbitrary(conn, errBadCmd)
			return 
		}

		cmd, _ := ms[0].Str()
		key, _ := ms[1].Str()
		log.Println(cmd, key)

		resp.WriteMessage(conn, m)
	}
}
