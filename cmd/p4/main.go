package main

import (
	"bytes"
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

func HandleResponse(server net.PacketConn, buf []byte, addr net.Addr, dict map[string]string) error {
	log.Println("Received:", string(buf))
	str := string(bytes.Trim(buf, "\x00"))
	// str := string(buf)
	if strings.ContainsRune(str, '=') {
		// this is an insert
		log.Println("This is an insert")
		spl := strings.SplitN(str, "=", 2)
		// log.Println("Split into")
		log.Printf("Split into '%s' and '%s'\n", spl[0], spl[1])
		if spl[0] == "version" {
			dict[spl[0]] = spl[1]
		}
		// this is a retrieve
		log.Printf("This is a query for '%s'", str)
		if val, ok := dict[str]; ok {
			// it's in, respond the value
			log.Println("Value:", val)
			server.WriteTo([]byte(fmt.Sprintf("%s=%s", str, val)), addr)
		} else {
			// not found
			log.Println("Not found")
		}
	}
	return nil
}

func ListenServer(ip string, port int) (err error) {
	dict := make(map[string]string)
	dict["version"] = "1.0"

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
			if err := HandleResponse(server, buf, addr, dict); err != nil {
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
