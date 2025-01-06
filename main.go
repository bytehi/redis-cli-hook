package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"

	"github.com/bytehi/resp"
)

var logger = log.New(os.Stdout, "", log.LstdFlags)

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

// start proxy to connect redis, and return proxy local listen port
func startProxy(ip, port string) (string, func()) {
	//random listen local port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		logger.Fatalf("Error listening: %v", err)
	}
	_, proxyPort, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		logger.Fatalf("Error parsing address:%s", err.Error())
	}
	logger.Printf("Server started, listening on: %s, proxyPort: %s", listener.Addr().String(), proxyPort)

	go func() {
		for {
			clientConn, err := listener.Accept()
			if err != nil {
				if !errors.Is(err, net.ErrClosed) {
					logger.Printf("Error accepting connection: %v", err)
				}
				break
			}

			serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", ip, port))
			if err != nil {
				log.Printf("Error connecting to server: %v", err)
				clientConn.Close()
				continue
			}

			go forwardOneByOne(clientConn, serverConn)
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

func forwardOneByOne(src, dst net.Conn) {
	defer src.Close()
	defer dst.Close()

	srcReader := bufio.NewReader(src)
	dstReader := bufio.NewReader(dst)
	for {
		req, err := resp.Parse(srcReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatalf("Error parse resp: %v\n", err)
		}
		cmd, args := toRequestCmd(req)
		for _, plugin := range Plugins {
			cmd, args, err = plugin.BeforeCommand(cmd, args)
			if err != nil {
				logger.Fatalf("Error before command: %v\n", err)
			}
		}

		members := make([]*resp.RESP, 0, len(args)+1)
		members = append(members, resp.NewBulkString(cmd))
		for _, arg := range args {
			members = append(members, resp.NewBulkString(arg))
		}
		req = resp.NewArray(members)
		bs, err := req.Marshal()
		if err != nil {
			logger.Fatalf("Error marshal resp: %v\n", err)
		}

		_, err = dst.Write(bs)
		if err != nil {
			logger.Fatalf("Error forwarding data: %v\n", err)
		}

		ack, err := resp.Parse(dstReader)
		if err != nil {
			if err == io.EOF {
				break
			}
			logger.Fatalf("Error parse response: %v", err)
		}

		for _, plugin := range Plugins {
			ack, err = plugin.AfterCommand(cmd, args, ack)
			if err != nil {
				logger.Fatalf("Error after command: %v\n", err)
			}
		}
		bs, err = ack.Marshal()
		if err != nil {
			logger.Fatalf("Error marshal response: %v\n", err)
		}

		_, err = src.Write(bs)
		if err != nil {
			logger.Fatalf("Error forwarding response: %v\n", err)
		}
	}
}

func toRequestCmd(req *resp.RESP) (string, []string) {
	if req.Type != resp.Array {
		logger.Fatalf("request: %s not array\n", req)
	}
	members := req.Value.([]*resp.RESP)
	if len(members) == 0 {
		logger.Fatalf("request: %s not array\n", req)
	}
	cmd := members[0].Value.(string)
	args := make([]string, 0, len(members)-1)
	for _, member := range members[1:] {
		args = append(args, member.Value.(string))
	}
	return cmd, args
}
