package srv

import (
	"log"
	"net"
	"time"

	"github.com/mediocregopher/breadis/bak"
	"github.com/mediocregopher/breadis/config"
	"github.com/mediocregopher/radix.v2/redis"
)

func Listen() {
	ln, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalf("Listening on %s: %s", config.ListenAddr, err)
	}
	log.Printf("Listening on %s", config.ListenAddr)
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

	rr := redis.NewRespReader(conn)
	for {
		err := conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		if err != nil {
			log.Printf("error setting read deadline: %s", err)
			return
		}

		r := rr.Read()
		if redis.IsTimeout(r) {
			continue
		} else if r.IsType(redis.IOErr) {
			return
		}

		bak.Cmd(r).WriteTo(conn)
	}

}
