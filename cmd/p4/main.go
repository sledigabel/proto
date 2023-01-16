package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func usage() {
	fmt.Println("Error while parsing arguments")
	fmt.Println("Usage:")
	fmt.Println(os.Args[0], "<IP:PORT>")
}

func parseIPAndPort(arg string) (string, int, error) {
	if res := strings.Index(arg, ":"); res < 0 {
		return "", -1, errors.New("could not parse argument")
	} else {
		if port, err := strconv.Atoi(arg[res+1:]); err != nil {
			return "", -1, err
		} else {
			return arg[:res], port, nil
		}
	}
}

func HandleRequest(conn net.Conn) error {
	// reads the content
	//
	log.Println("Got a new connection from", conn.RemoteAddr().String())
	data, err := ioutil.ReadAll(conn)
	if err != nil {
		log.Println("error while reading in", err)
		return err
	}

	// looking at the data
	log.Println("Received:", string(data))
	return nil
}

func HandleResponse(buf []byte, addr net.Addr) error {
	log.Println("Received:", string(buf))
	return nil
}

func ListenServer(ip string, port int) (err error) {
	// server, err := net.Listen("udp", fmt.Sprintf("%s:%d", ip, port))
	server, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	log.Println("Listening on", ip, "port", port)
	defer server.Close()
	for {
		buf := make([]byte, 1024)
		_, addr, err := server.ReadFrom(buf)
		log.Println("Just read something")
		if err != nil {
			log.Println(err)
			continue
		}

		go func() {
			if err := HandleResponse(buf, addr); err != nil {
				log.Println(err)
			}
		}()
	}
}

func main() {
	if len(os.Args) != 2 {
		usage()
		os.Exit(1)
	}
	ip, port, err := parseIPAndPort(os.Args[1])
	if err != nil {
		fmt.Println("IP:Port cannot be parsed")
		usage()
		os.Exit(1)
	}
	// fmt.Println(ip, port)
	log.Println("Starting server")
	if err := ListenServer(ip, port); err != nil {
		log.Fatalln(err)
	}
}
