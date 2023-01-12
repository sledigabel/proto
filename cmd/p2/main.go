package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
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

type Message struct {
	Type string
	Arg1 int32
	Arg2 int32
}

func readMessage(conn net.Conn) (*Message, error) {
	// read Message type
	messageType := make([]byte, 1)
	arg1 := make([]byte, 4)
	arg2 := make([]byte, 4)

	if _, err := conn.Read(messageType); err != nil {
		return nil, err
	}
	if _, err := conn.Read(arg1); err != nil {
		return nil, err
	}
	if _, err := conn.Read(arg2); err != nil {
		return nil, err
	}
	log.Println("Raw data for arg1 and arg2:")
	log.Println(arg1)
	log.Println(arg2)

	var arg1Value, arg2Value int32
	if err := binary.Read(bytes.NewBuffer(arg1), binary.BigEndian, &arg1Value); err != nil {
		return nil, err
	}
	if err := binary.Read(bytes.NewBuffer(arg2), binary.BigEndian, &arg2Value); err != nil {
		return nil, err
	}

	message := Message{
		Type: string(messageType),
		Arg1: arg1Value,
		Arg2: arg2Value,
	}
	log.Println("Read message:", message)
	return &message, nil
}

func HandleRequest(conn net.Conn) error {
	var previous []*Message

	defer conn.Close()
	for {
		message, err := readMessage(conn)
		if err != nil {
			return err
		}
		switch message.Type {
		case "I":
			previous = append(previous, message)
		case "Q":
			var res, sum, total int32
			for _, prev := range previous {
				// for each previous, check whether the timestamp in within
				if message.Arg1 <= prev.Arg1 && prev.Arg1 <= message.Arg2 {
					sum += prev.Arg2
					total += 1
				}
			}
			// sending results
			res = sum / total
			log.Println("Writing results:", res)
			binary.Write(conn, binary.BigEndian, res)
			return nil
		}

	}
}

func ListenServer(ip string, port int) (err error) {
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	log.Println("Listening on", ip, "port", port)
	defer server.Close()
	for {
		conn, err := server.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go func() {
			err := HandleRequest(conn)
			if err != nil {
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
