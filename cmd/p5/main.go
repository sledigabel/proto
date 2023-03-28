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
	InputConn  net.Conn
	OutputConn net.Conn
}

func HandleRequest(conn *Connection) {
	// initiating the connection to chat.protohackers.com
	conn.OutputConn, _ = net.Dial("tcp", "chat.protohackers.com:16963")
	defer conn.OutputConn.Close()
	defer conn.InputConn.Close()

	// anything received by the server is forwarded to the client
	go func() {
		var err error
		for err == nil {
			newReader := bufio.NewReader(conn.OutputConn)
			newWriter := bufio.NewWriter(conn.InputConn)
			_, err = newWriter.ReadFrom(newReader)
		}
	}()

	// anything sent by the client is forwarded to the server
	go func() {
		// checks the string from the InputConn
		scanner := bufio.NewScanner(conn.InputConn)
		scanner.Split(bufio.ScanLines)
		var err error
		for scanner.Scan() {
			message := scanner.Text()
			// replace with Bogus
			message = replaceBoguscoin(message)
			_, err = conn.OutputConn.Write([]byte(message))
			if err != nil {
				log.Panic(err)
			}
		}
	}()
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
		conn.InputConn, err = server.Accept()
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
