package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func usage() {
	fmt.Println("Error while parsing arguments")
	fmt.Println("Usage:")
	fmt.Println(os.Args[0], "<IP:PORT>")
}

var BogusCoinRe = regexp.MustCompile("\b(7[0-9a-zA-Z]{25,34})\b")

func replaceBoguscoin(arg string) string {
	// find all the BogusCoin
	matches := BogusCoinRe.FindAllString(arg, -1)
	for _, match := range matches {
		log.Printf("Replacing %s with 7YWHMfk9JZe0LM0g1ZauHuiSxhI\n", match)
		arg = strings.ReplaceAll(arg, match, "7YWHMfk9JZe0LM0g1ZauHuiSxhI")
	}
	return arg
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

type Connection struct {
	ClientConn net.Conn
	ProtoConn  net.Conn
}

func HandleRequest(conn *Connection) {
	// initiating the connection to chat.protohackers.com
	conn.ProtoConn, _ = net.Dial("tcp", "chat.protohackers.com:16963")

	defer conn.ProtoConn.Close()
	defer conn.ClientConn.Close()
	log.Printf("New connection from %s", conn.ProtoConn.RemoteAddr())

	// anything received by the server is forwarded to the client
	go func() {
		scanner := bufio.NewScanner(conn.ProtoConn)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			message := fmt.Sprintf("%s\n", scanner.Text())
			log.Println("received from server:", message)
			b, err := conn.ClientConn.Write([]byte(message))
			if err != nil {
				log.Panic(err)
			} else {
				log.Printf("sent %d bytes to client", b)
			}
		}
		log.Println("Closing server -> client loop")
	}()

	// anything sent by the client is forwarded to the server
	// checks the string from the ClientConn
	scanner := bufio.NewScanner(conn.ClientConn)
	scanner.Split(bufio.ScanLines)
	var err error
	for scanner.Scan() {
		message := fmt.Sprintf("%s\n", scanner.Text())
		log.Println("received from client:", message)
		// replace with Bogus
		message = replaceBoguscoin(message)
		_, err = conn.ProtoConn.Write([]byte(message))
		if err != nil {
			log.Panic(err)
		}
	}
	log.Println("Client disconnected")
}

func ListenServer(ip string, port int) (err error) {
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	log.Println("Listening on", ip, "port", port)
	defer server.Close()
	for {
		var conn Connection
		conn.ClientConn, err = server.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go HandleRequest(&conn)
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
