package main

/*
// linux
#include <sys/types.h>
#include <sys/stat.h>
#include <sys/time.h>
#include <sys/ioctl.h>
#include <sys/param.h>
#include <sys/resource.h>

// net
#include <netinet/in.h>
#include <netinet/ip.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <ifaddrs.h>
#include <netinet/if_ether.h>
#include <linux/netfilter_ipv4.h>
#include <net/if.h>
#include <netinet/in.h>
#include <netinet/ip.h>
#include <netinet/tcp.h>
#include <netinet/udp.h>
#include <stdint.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <stdlib.h>
#include <errno.h>
#include <assert.h>
#include <signal.h>
#include <limits.h>
#include <time.h>
#include <pthread.h>
int ListenUdp(int port) {
	int OPT_ON = 1;
	int OPT_SIZE = sizeof(OPT_ON);
	int fd = socket(AF_INET, SOCK_DGRAM, 0);
	if (setsockopt(fd, SOL_SOCKET, SO_REUSEADDR, &OPT_ON, OPT_SIZE) < 0) {
		goto fail;
	}
	if (setsockopt(fd, SOL_SOCKET, SO_REUSEPORT, &OPT_ON, OPT_SIZE) < 0) {
		goto fail;
	}
	struct sockaddr_in local_addr;
	memset(&local_addr, 0, sizeof(local_addr));

	local_addr.sin_family = AF_INET;
	local_addr.sin_addr.s_addr = 0;
	local_addr.sin_port = htons(port);
	if (bind(fd, (struct sockaddr*)(&local_addr), sizeof(local_addr))) {
		goto fail;
	}
	if (0 != setsockopt( fd, IPPROTO_IP, IP_PKTINFO, &OPT_ON, OPT_SIZE)){
		goto fail;
	};
	return fd;
fail:
	close(fd);
	return -1;
}

struct Recvfrom_return {
	uint32_t ip;
	uint16_t port;
	int errnum;
};

struct Recvfrom_return Recvfrom(int fd, char *buff, int len) {
	struct msghdr msg;
	struct sockaddr_in src_addr, dst_addr;
	struct iovec iov[1];
	char cntrlbuf[CMSG_SPACE(sizeof(struct in_pktinfo))];
	struct Recvfrom_return ret;
	ret.errnum = 0;
	size_t length;
	msg.msg_name = &src_addr;
	msg.msg_namelen = sizeof(src_addr);
	msg.msg_control = cntrlbuf;
	msg.msg_controllen = sizeof(cntrlbuf);
	iov[0].iov_base = buff;
	iov[0].iov_len = len;
	msg.msg_iov = iov;
	msg.msg_iovlen = 1;
	length = recvmsg(fd, &msg, 0);
	if (length < 0) {
		ret.errnum = length;
		return ret;
	}

	//int result = -1;
	//struct cmsghdr* cmsg;
	//for (cmsg = CMSG_FIRSTHDR(msg); cmsg; cmsg = CMSG_NXTHDR(msg, cmsg)) {
	//	if (cmsg->cmsg_level == SOL_IP && cmsg->cmsg_type == IP_RECVORIGDSTADDR) {
	//		memcpy (original_dst, CMSG_DATA(cmsg), sizeof (struct sockaddr_in));
	//		result = 0;
	//		break;
	//	}
	//}
	//if (result <= 0) {
	//	return -1;
	//}
	ret.ip = src_addr.sin_addr.s_addr;
	ret.port = src_addr.sin_port;
	return ret;
}

char *MakeBuf(int size) {
	return malloc(size);
}

void FreeBuf(char *buf) {
	free(buf);
}

 */
import "C"
import (
"time"
"fmt"
_ "net/http/pprof"
"net/http"
"sync/atomic"
)


type ServerEndpoint_t struct{
	Endpoint
	fd C.int
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

//func NewServerEndpoint(fd C.int, addr net.Addr, callback recv_callback) (*ServerEndpoint_t) {
//	server_ep := new (ServerEndpoint_t)
//	server_ep.initEndpoint(0, func(buf []byte, size int) {
//		// output cb
//		//fmt.Println("output data ", size, conn)
//		// todo maybe we need go here, but buf not support now
//		_, err  := conn.WriteTo(buf[:size], addr)
//		if err != nil {
//			fmt.Println(err)
//		}
//		//fmt.Println("write to", addr.String(), "len", n)
//	})
//	server_ep.recv_cb = callback
//	server_ep.fd = fd
//
//	return server_ep
//}

func listenudp(index int) {

	fd := C.ListenUdp(8000)
	if fd < 0 {
		fmt.Println("listen udp 8000 fail")
		return
	}
	buff := C.MakeBuf(1500)
	for {
		ip, port, err := C.Recvfrom(fd, buff, C.int(1500))
		fmt.Println("recv msg",  C.GoString(buff), ip, port, err)
		//var ep *ServerEndpoint_t
		//var ok bool
		//if ep, ok = kcp_map[addr.String()]; !ok {
		//	ep = NewServerEndpoint(conn, addr, func(buf[]byte, size int) {
		//		fmt.Println("recvUserData", string(buf))
		//	})
		//	kcp_map[addr.String()] = ep
		//	ep.fd = fd
		//	conv, payload, size, err := ep.GetSYN(data, n)
		//	if err {
		//		fmt.Println(err)
		//		return
		//	}
		//	ep.Input(data, n)
		//	ep.SendSYN([]byte("hello"), len("hello"))
		//	fmt.Println("recvUserData syn conv", conv, "payload", string(payload), "len", size)
		//	go ep.Process()
		//	go senddata(ep)
		//	go statistic_snd()
		//} else {
		//	ep.Input(data, n)
		//}
		//fmt.Println("index", index, "recvUserData", n, "from", addr.String())
	}
}

func main() {
	listenudp(1)
	for {
		time.Sleep(time.Minute)
	}
}