package main

import (
"fmt"
"net"
	"time"
	"github.com/fatih/pool"
	"sync/atomic"
	"net/http"
)


type ClientEndpoint_t struct {
	Endpoint
	conn          net.Conn
}

func init() {
	go func() {
		http.ListenAndServe("0.0.0.0:10000", nil)
	}()

}

func NewClientEndpoint(conn net.Conn, recv_cb recv_callback) (*ClientEndpoint_t) {
	client_ep := new(ClientEndpoint_t)
	client_ep.conn = conn

	client_ep.initEndpoint(12345, func(buf []byte, size int) {
		// output cb
		// todo maybe we need go here, but buf not support now
		_, err  := conn.Write(buf[:size])
		if err != nil {
			fmt.Println(err)
		}
		//fmt.Println("write n", n)
	})
	client_ep.recv_cb = recv_cb
	return client_ep
}

func (client_ep *ClientEndpoint_t) readDataFromServer() {
	for {
		buf := make([]byte, 1500)
		n, err := client_ep.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		client_ep.Input(buf, n)
	}
}

func process(ep *ClientEndpoint_t){
	data := make([]byte, 1500)
	copy(data, []byte("hello world"))
	ep.kcp.SendSYN(data, len("hello world"))
	ep.Send(data, len("hello world"))
	go ep.readDataFromServer()
	ep.Process()
}

var rcv_byte uint32

func statistic_rcv() {
	for {
		rate := rcv_byte / 10
		fmt.Println("recv rate", rate/1024, "KB/s")
		atomic.SwapUint32(&rcv_byte, 0)
		time.Sleep(time.Second * 10)
	}
}

var cnt int
var before_ts time.Time
func main() {
	factory := func() (net.Conn, error) { return net.Dial("udp", "192.168.195.133:8000") }
	p, err := pool.NewChannelPool(5, 30, factory)
	if err != nil {
		fmt.Println(err)
		return
	}
	conn, err := p.Get()
	cnt = 0
	before_ts = time.Now()
	ep := NewClientEndpoint(conn, func(buf []byte, size int) {
		atomic.AddUint32(&rcv_byte, uint32(size))
	})
	go process(ep)
	go statistic_rcv()
	for{
		time.Sleep(time.Minute)
	}
}
