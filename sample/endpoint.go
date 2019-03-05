	package main

	import (
		"net"
		"github.com/tenchlee/gcp"
		"math/rand"
		"fmt"
		"time"
	)

	type recv_callback func(buf []byte, size int)

	type Endpoint struct {
		input_queue   chan []byte
		output_queue  chan []byte
		recv_queue    chan *gcp.ByteBuffer
		conv          uint32
		kcp           *gcp.KCP
		src_addr      net.Addr
		recv_cb       recv_callback
		prior_tick_ts uint32
	}

	func (ep *Endpoint) initEndpoint(conv uint32, output_cb func(buf []byte, size int)) {
		ep.conv = rand.Uint32()
		ep.input_queue = make (chan []byte, 10240)
		ep.output_queue = make (chan []byte, 10240)
		ep.recv_queue = make (chan *gcp.ByteBuffer, 10240)

		ep.kcp = gcp.NewKCP(conv, output_cb)
	}

	func (ep *Endpoint)Send(data []byte, len int) {
		select {
			case ep.output_queue <- data[:len]:
			default:
				fmt.Println("output queue is full")
		}
	}

	func (ep *Endpoint)Input(data []byte, len int) {
		select {
			case ep.input_queue <- data[:len]:
			default:
				fmt.Println("input queue is full")
		}
	}

	func (ep *Endpoint) Process() {
		go ep.sendData2User()
		tick := time.Tick(10 * time.Millisecond)
		var data []byte
		for {
			select {
			case data = <- ep.input_queue:
				//fmt.Println("recvUserData", len(data))
				start := uint32(time.Now().UnixNano()/1e6)
				i := 0
				for {
					ret := ep.kcp.Input(data, true, true)
					if ret != 0 {
						fmt.Println("input fail ret", ret)
					}
					i++
					if len (ep.input_queue) > 0 && i < 10 {
						data = <- ep.input_queue
					} else {
						break
					}
				}
				ep.kcp.ForceFlush()
				end1 := uint32(time.Now().UnixNano()/1e6)
				ep.recvUserData()
				end2 := uint32(time.Now().UnixNano()/1e6)
				if end1 - start > 10 || end2 - end1 > 10 {
					fmt.Println("input end1", end1 - start, "end2", end2 - end1)
				}
			case data = <- ep.output_queue:
				//fmt.Println("send", len(data))
				start := uint32(time.Now().UnixNano()/1e6)
				i := 0
				for {
					ep.kcp.Send(data)
					i++
					if len (ep.output_queue) > 0 && i < 10 {
						data = <- ep.output_queue
					} else {
						break
					}
				}
				ep.kcp.ForceFlush()
				end := uint32(time.Now().UnixNano()/1e6)
				if end - start > 10 {
					fmt.Println("kcp send", end-start)
				}
			case <- tick:
				start := uint32(time.Now().UnixNano()/1e6)
				ep.kcp.Update()
				end1 := uint32(time.Now().UnixNano()/1e6)
				ep.recvUserData()
				end2 := uint32(time.Now().UnixNano()/1e6)
				if end1 - start > 10 || end2 - end1 > 10 {
					fmt.Println("tock", end1 - start, "end2", end2 - end1)
				}
			}
		}
	}

	func (ep *Endpoint) GetSYN(data []byte, size int) (conv uint32, payload []byte, length int, error bool) {
		return ep.kcp.GetSYN(data, size)
	}

	func (ep *Endpoint) SendSYN(data []byte, size int) {
		ep.kcp.SendSYN(data, size)
	}

	func (ep *Endpoint) recvUserData() {
		for {
			data, n := ep.kcp.Recv()
			if n <= 0 {
				break
			}
			select {
				case ep.recv_queue <- data:
				default:
					fmt.Println("recvUserData queue is full")
					return
			}
		}
	}

	func (ep *Endpoint) sendData2User() {
		for {
			data := <- ep.recv_queue
			ep.recv_cb(data.Bytes(), data.Len())
			ep.kcp.ReleaseUserData(data)
		}
	}