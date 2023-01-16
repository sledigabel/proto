package main

import (
	"bufio"
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

const validUsernameChars string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func validateUsername(username string) bool {
	if len(username) == 0 || len(username) >= 16 {
		return false
	}
	for _, c := range username {
		if !strings.ContainsRune(validUsernameChars, c) {
			return false
		}
	}
	return true
}

func validateMessage(message string) bool {
	if len(message) == 0 || len(message) >= 1000 {
		return false
	}
	return true
}

type User struct {
	Conn net.Conn
	Name string
}

func (u *User) WriteMessage(s string) error {
	_, err := u.Conn.Write([]byte(s))
	return err
}

type Room struct {
	Users []*User
}

func (r *Room) SendMessageFrom(user *User, message string) error {
	for _, u := range r.Users {
		if u != user {
			u.WriteMessage(fmt.Sprintf("[%s] %s\n", user.Name, message))
		}
	}
	return nil
}

func (r *Room) AddUser(user *User) {
	log.Println("New user joining", user.Name)
	var userList []string

	log.Println("Broadcasting")
	for _, u := range r.Users {
		userList = append(userList, u.Name)
		u.WriteMessage(fmt.Sprintf("* %s has entered the room\n", user.Name))
	}
	user.WriteMessage(fmt.Sprintf("* The room contains: %s\n", strings.Join(userList, ", ")))
	r.Users = append(r.Users, user)
}

func (r *Room) NotifyUserLeave(user *User) {
	var newUserList []*User
	for _, u := range r.Users {
		if u != user {
			newUserList = append(newUserList, u)
			u.WriteMessage(fmt.Sprintf("* %s has left the room\n", user.Name))
		}
	}
	r.Users = newUserList
}

func HandleRequest(user *User, room *Room) error {
	log.Println("New connection, prompting for username")

	scanner := bufio.NewScanner(user.Conn)
	scanner.Split(bufio.ScanLines)
	// ask for name
	user.WriteMessage("Welcome to budgetchat! What shall I call you?\n")
	if scanner.Scan() {
		username := scanner.Text()
		log.Println("Scanned user:", username)
		if !validateUsername(username) {
			log.Println("Non valid username")
			user.WriteMessage("Invalid username")
			user.Conn.Close()
		}
		user.Name = username
	} else {
		return errors.New("could not get username")
	}

	// user accepted, make the announcement
	room.AddUser(user)
	defer room.NotifyUserLeave(user)

	// then loop on all input
	for scanner.Scan() {
		message := scanner.Text()
		if validateMessage(message) {
			room.SendMessageFrom(user, message)
		}
	}
	return nil
}

func ListenServer(ip string, port int) (err error) {
	var room Room
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip, port))
	if err != nil {
		return err
	}
	log.Println("Listening on", ip, "port", port)
	defer server.Close()
	for {
		conn, err := server.Accept()
		user := User{Conn: conn}
		if err != nil {
			log.Fatalln(err)
		}
		go HandleRequest(&user, &room)
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
