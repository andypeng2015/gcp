package main

import (
	"time"
	"fmt"
	"net"
	"context"
	_ "net/http/pprof"
	"net/http"
	"sync/atomic"
)


type ServerEndpoint_t struct{
	Endpoint
	conn net.PacketConn
}

var kcp_map  map[string]*ServerEndpoint_t

func init() {
	kcp_map = make(map[string]*ServerEndpoint_t)

	go func() {
		http.ListenAndServe("0.0.0.0:10000", nil)
	}()

}

var send_byte uint32

func senddata(ep *ServerEndpoint_t) {
	data := make([]byte, 1000)
	copy(data, "hello world")
	for {
		ep.Send(data, len(data))
		atomic.AddUint32(&send_byte, uint32(len(data)))
		time.Sleep(time.Millisecond * 10)
	}
}

func statistic_snd() {
	for {
		rate := send_byte / 10
		fmt.Println("send rate", rate)
		atomic.SwapUint32(&send_byte, 0)
		time.Sleep(time.Second * 10)
	}
}

func NewServerEndpoint(conn net.PacketConn, addr net.Addr, callback recv_callback) (*ServerEndpoint_t) {
	server_ep := new (ServerEndpoint_t)
	server_ep.initEndpoint(0, func(buf []byte, size int) {
		// output cb
		//fmt.Println("output data ", size, conn)
		// todo maybe we need go here, but buf not support now
		_, err  := conn.WriteTo(buf[:size], addr)
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println("write to", addr.String(), "len", n)
	})
	server_ep.recv_cb = callback
	server_ep.conn = conn
	server_ep.conn = conn

	return server_ep
}

func listenudp(index int) {
	d := net.ListenConfig{
		Control:   Control,
	}
	conn, err := d.ListenPacket(context.Background(), "udp", ":8000")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		data := make([]byte, 1500)
		n, addr, err := conn.ReadFrom(data)
		if err != nil {
			fmt.Println(err)
			return
		}
		var ep *ServerEndpoint_t
		var ok bool
		if ep, ok = kcp_map[addr.String()]; !ok {
			ep = NewServerEndpoint(conn, addr, func(buf[]byte, size int) {
				fmt.Println("recvUserData", string(buf))
			})
			kcp_map[addr.String()] = ep
			conv, payload, size, err := ep.GetSYN(data, n)
			if err {
				fmt.Println(err)
				return
			}
			ep.Input(data, n)
			ep.SendSYN([]byte("hello"), len("hello"))
			fmt.Println("recvUserData syn conv", conv, "payload", string(payload), "len", size)
			go ep.Process()
			go senddata(ep)
			go statistic_snd()
		} else {
			ep.Input(data, n)
		}
		//fmt.Println("index", index, "recvUserData", n, "from", addr.String())
	}
}

func main() {
	listenudp(1)
	for {
		time.Sleep(time.Minute)
	}
}