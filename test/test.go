package main

import (
	"fmt"
	"time"
)

func putdata(ch chan[]byte) {
	time.Sleep(time.Minute)
	ch <- []byte("hello")
}
func main() {
	ch := make (chan []byte, 1)
	//go putdata(ch)
	//var before_ts uint32
	//var cnt int
	//for {
	//	select {
	//	case <- ch:
	//		//fmt.Println(string(data))
	//	case <-time.Tick(time.Millisecond * 10):
	//		now_ts := uint32(time.Now().UnixNano()/1e6)
	//		if before_ts > 0 && now_ts - before_ts > 20 {
	//			cnt++
	//			fmt.Println("tick outof ts", now_ts - before_ts, "cnt", cnt)
	//		}
	//		before_ts = now_ts
	//
	//	}
	//}
go putdata(ch)
	data := <- ch
	fmt.Println(data)
}
