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

func validateRequest(req map[string]interface{}) bool {
	if _, ok := req["method"]; !ok {
		log.Println("Method missing")
		return false
	}
	if _, ok := req["number"]; !ok {
		log.Println("Number missing")
		return false
	}
	switch req["method"].(type) {
	case string:
	default:
		log.Println("Method is not string")
		return false
	}

	if _, ok := req["method"].(string); !ok {
		log.Println("method is not string")
		return false
	}

	if _, ok := req["number"].(float64); !ok {
		log.Println("Number is not int64")
		return false
	}

	// if float64(int64(req["number"].(float64))) != req["number"].(float64) {
	// 	return false
	// }

	return true
}

func HandleRequest(conn net.Conn) error {
	defer conn.Close()

	log.Println("Got a new connection from", conn.RemoteAddr().String())

	// data, err := ioutil.ReadAll(conn)

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Bytes()
		if err := scanner.Err(); err != nil {
			return err
		}
		log.Println("Got some new data", string(data))

		var orig map[string]interface{}
		if err := json.Unmarshal(data, &orig); err != nil {
			_, _ = conn.Write([]byte("malformed"))
			return err
		}

		if !validateRequest(orig) {
			log.Println("Invalid input")
			_, _ = conn.Write([]byte("malformed"))
			return errors.New("invalid input")
		}

		var resp Response
		if float64(int64(orig["number"].(float64))) != orig["number"].(float64) {
			resp = Response{Method: orig["method"].(string), Prime: false}
		} else {
			req := Request{Method: orig["method"].(string), Number: int64(orig["number"].(float64))}
			resp = Response{Method: req.Method, Prime: big.NewInt(req.Number).ProbablyPrime(0)}
		}

		log.Println("Writing response", resp)
		if data, err := json.Marshal(resp); err != nil {
			return errors.New("could not marshall")
		} else {
			// adding CR in the end
			data = append(data, byte(10))
			if _, err = conn.Write(data); err != nil {
				log.Println("Could not send back data")
				return err
			}
			log.Println("Sending", string(data))
		}

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
