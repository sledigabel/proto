package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
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

type Request struct {
	Method string `json:"method"`
	Number int64  `json:"number"`
}

type Response struct {
	Method string `json:"method"`
	Prime  bool   `json:"prime"`
}

func HandleRequest(conn net.Conn) error {
	defer conn.Close()

	log.Println("Got a new connection from", conn.RemoteAddr().String())

	// data, err := ioutil.ReadAll(conn)

	scanner := bufio.NewScanner(conn)
	scanner.Scan()
	data := scanner.Bytes()
	err := scanner.Err()
	log.Println("Got some new data", string(data))
	if err != nil {
		log.Println("error while reading in", err)
		return err
	}

	// now we need to unmarshall it
	var req Request
	if err = json.Unmarshal(data, &req); err != nil {
		_, _ = conn.Write([]byte("malformed"))
		return err
	}

	if req.Method != "isPrime" {
		_, _ = conn.Write([]byte("malformed"))
		return errors.New("method is not IsPrime")
	}

	log.Println(req)

	resp := Response{Method: req.Method, Prime: big.NewInt(req.Number).ProbablyPrime(0)}

	// now need to write the data back
	log.Println("Writing response", resp)
	if data, err := json.Marshal(resp); err != nil {
		return errors.New("could not marshall")
	} else {
		if _, err = conn.Write(data); err != nil {
			log.Println("Could not send back data")
			return err
		}
		log.Println("Sending", string(data))
	}
	return nil
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