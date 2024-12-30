package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/bytehi/resp"
)

func parse() (ip, port string, leftArgs []string) {
	ip = "127.0.0.1"
	port = "6379"
	leftArgs = make([]string, 0, len(os.Args)-1)
	for i := 1; i < len(os.Args); i++ {
		if os.Args[i] == "-h" {
			i++
			ip = os.Args[i]
		} else if os.Args[i] == "-p" {
			i++
			port = os.Args[i]
		} else {
			leftArgs = append(leftArgs, os.Args[i])
		}
	}
	return
}

func main() {
	ip, port, leftArgs := parse()
	fmt.Println(os.Args, ip, port, leftArgs)
	proxyPort, closer := startProxy(ip, port)
	defer closer()
	leftArgs = append([]string{"-p", proxyPort}, leftArgs...)
	execRedisCli(leftArgs)
}

//start proxy to connect redis, and return proxy local listen port
func startProxy(ip, port string) (string, func()) {
	//random listen local port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Error listening: %v", err)
	}
	_, proxyPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		log.Fatalf("Error parsing address:", err.Error())
	}
	log.Println("Server started, listening on:", listener.Addr().String(), proxyPort)

	go func() {
		for {
			clientConn, err := listener.Accept()
			if err != nil {
				log.Printf("Error accepting connection: %v", err)
				break
			}

			serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip, port))
			if err != nil {
				log.Printf("Error connecting to server: %v", err)
				clientConn.Close()
				continue
			}

			//go forward(clientConn, serverConn)
			//go forward(serverConn, clientConn)
			go forwardOneByOne(clientConn, serverConn, reqHookFunc, ackHookFunc)
		}
	}()
	return proxyPort, func() { listener.Close() }
}

func execRedisCli(args []string) {
	log.Println("call redis-cli, args:", args)
	cmd := exec.Command("redis-cli", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting command:", err.Error())
		return
	}
	cmd.Wait()
}

//unused
func forward(src io.ReadCloser, dst io.WriteCloser) {
	defer src.Close()
	defer dst.Close()

	reader := bufio.NewReader(src)
	for {
		msg, err := resp.Parse(reader)
		if err != nil {
			log.Printf("Error parse resp: %v\n", err)
			break
		}
		//fmt.Println(msg)
		bs, err := msg.Marshal()
		if err != nil {
			log.Printf("Error marshal resp: %v\n", err)
			break
		}
		_, err = dst.Write(bs)
		if err != nil {
			log.Printf("Error forwarding data: %v", err)
			break
		}
	}
}

func forwardOneByOne(src, dst net.Conn, reqHook ReqHookFunc, ackHook AckHookFunc) {
	defer src.Close()
	defer dst.Close()

	srcReader := bufio.NewReader(src)
	dstReader := bufio.NewReader(dst)
	for {
		msg, err := resp.Parse(srcReader)
		if err != nil {
			log.Printf("Error parse resp: %v\n", err)
			break
		}

		msg = reqHook(msg)
		bs, err := msg.Marshal()
		if err != nil {
			log.Printf("Error marshal resp: %v\n", err)
			break
		}
		_, err = dst.Write(bs)
		if err != nil {
			log.Printf("Error forwarding data: %v\n", err)
			break
		}

		response, err := resp.Parse(dstReader)
		if err != nil {
			log.Printf("Error parse response: %v", err)
			break
		}
		response = ackHook(msg, response)
		bs, err = response.Marshal()
		if err != nil {
			log.Printf("Error marshal response: %v\n", err)
			break
		}
		_, err = src.Write(bs)
		if err != nil {
			log.Printf("Error forwarding response: %v\n", err)
			break
		}
	}
}

type ReqHookFunc func(req *resp.RESP) *resp.RESP
type AckHookFunc func(req, rsp *resp.RESP) *resp.RESP

func reqHookFunc(req *resp.RESP) *resp.RESP {
	log.Println("req", req)
	return req
}
func ackHookFunc(req, ack *resp.RESP) *resp.RESP {
	log.Println("ack", ack)
	reqArray := reqToArray(req)
	if strings.ToUpper(reqArray[0]) == "PING" {
		ack = adjustPingAck(ack)
	}
	return ack
}

func reqToArray(req *resp.RESP) []string {
	members := req.Value.([]*resp.RESP)
	array := make([]string, 0, len(members))
	for _, member := range members {
		array = append(array, member.Value.(string))
	}
	return array
}

func adjustPingAck(ack *resp.RESP) *resp.RESP {
	ack.Value = "received ping ack:" + ack.Value.(string)
	return ack
}
