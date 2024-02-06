package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	login = os.Args[1]
	pass  = os.Args[2]
)

var (
	validChan = make(chan string, 1)
	errChan   = make(chan string, 1)
	msgChan   = make(chan *MsgQuery, 10)
	calcCahn  = make(chan []byte, 1)
	doneChan  = make(chan []byte, 1)
)

const (
	address = ":8080"

	binaryFuncPath = `./func`          // the math functions
	dataReadyPath  = `./ReadyData.txt` // the file where income args for the func execution will be stored

	binaryDataCalculPath = `./calcul`   // the programm which ask params for your func and stores them in a single json
	dataClPath           = `dataCl.txt` //file where calcul will store json-argument (for field "data in C requuest")

	binaryResultPath = `./Results` // Programm which outputs results (json-answer is passed as first paramnetr)
)

type MsgQuery struct {
	Reciever string `json:"recc"`
	Message  string `json:"msg"`
}

type calcQuery struct {
	Function string `json:"func"`
	Data     []byte `json:"data"`
}

func Login(name, pass string, conn net.Conn) {
	conn.Write([]byte("I{\"login\":\"" + name + "\",\"pass\":\"" + pass + "\"}\n"))
	select {
	case valid := <-validChan:
		fmt.Println(valid)
	case err := <-errChan:
		panic(err)
	}
}

// send the message
func SendMessage(conn net.Conn) {
	fmt.Println("\nEnter reciebver:")
	var res, msg string
	_, err := fmt.Scan(&res)
	fmt.Print("Enter message:")
	_, err = fmt.Scan(&msg)
	if err != nil {
		panic(err)
	}

	m := &MsgQuery{
		Reciever: strings.TrimSpace(res),
		Message:  strings.TrimSpace(msg),
	}
	byt, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	conn.Write([]byte("M"))
	conn.Write(byt)
	conn.Write([]byte("\n")) // Always add \n at the end

	select {
	case valid := <-validChan:
		fmt.Println(valid)
	case err := <-errChan:
		fmt.Println(err)
	}
}

// stream the message results
func StreamMessage(login string, conn net.Conn) {
	fmt.Print("\nEnter message:")
	var msg string
	_, err := fmt.Scan(&msg)
	if err != nil {
		panic(err)
	}
	m := &MsgQuery{
		Reciever: login, // the Sender
		Message:  strings.TrimSpace(msg),
	}
	byt, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("S"))
	conn.Write(byt)
	conn.Write([]byte("\n")) // ALways add \n atthe end !

	select {
	case valid := <-validChan:
		fmt.Println(valid)
	case err := <-errChan:
		fmt.Println(err)

	}
}

// List the messages
func ListMessage() {
	empty := true

LOOP:
	for {
		select {
		case m := <-msgChan:
			fmt.Println("+\n|From: ", m.Reciever, "\n|Content:\n|\t", m.Message, "\n+")
			empty = false
		default:
			if empty {
				fmt.Println("no message")
			}
			break LOOP
		}
	}
}

// Declare
func Declare(conn net.Conn, f string) bool {
	conn.Write([]byte("P{\"func\":\"" + f + "\"}\n"))

	select {
	case valid := <-validChan:
		fmt.Println(valid)
		return true
	case err := <-errChan:
		fmt.Println(err)
		return false
	}
}

func Ready(conn net.Conn) {
	conn.Write([]byte("R\n"))
	for {
		data := <-calcCahn
		file, err := os.OpenFile(dataReadyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			panic(err)
		}
		_, err = file.Write(data)
		if err != nil {
			panic(err)
		}
		out, err := exec.Command(binaryFuncPath, dataReadyPath).Output()
		if err != nil {
			panic(err)
		}

		conn.Write([]byte("D"))
		conn.Write(out)
		conn.Write([]byte("\n"))
	}
}

func CalculateFunc(conn net.Conn) {
	fmt.Printf("\nEnter func name:")
	var f string
	_, err := fmt.Scan(&f)
	fmt.Println("Enter:")

	file, err := os.Create(dataClPath)
	if err != nil {
		panic(err)
	}
	file.Close()
	cmd := exec.Command(binaryDataCalculPath, dataClPath)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	err = cmd.Run()
	if err != nil {
		panic(err)
	}

	file, err = os.OpenFile(dataClPath, os.O_RDONLY, 0666)
	out, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	file.Close()

	c := &calcQuery{
		Function: f,
		Data:     out,
	}

	byt, err := json.Marshal(c)
	if err != nil {
		panic(err)
	}

	conn.Write([]byte("C"))
	conn.Write(byt)
	conn.Write([]byte("\n"))

	select {
	case ok := <-doneChan:
		cmd := exec.Command(binaryResultPath, string(ok))
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			panic(err)
		}
	case err := <-errChan:
		fmt.Println(err)
	}
}

func ReadMessages(reader bufio.Reader) {
	for {
		msgType, err := reader.ReadByte()
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(time.Second)
				panic("server has closed your conn")
			}
			time.Sleep(time.Second)
			panic(err)
		}
		if msgType == 'E' {
			errChan <- string(msg)
		} else if msgType == 'M' {
			m := &MsgQuery{}
			err = json.Unmarshal(msg, m)
			msgChan <- m
		} else if msgType == 'O' {
			validChan <- string(msg)
		} else if msgType == 'D' {
			doneChan <- msg
		} else if msgType == 'C' {
			calcCahn <- msg
		} else {
			fmt.Println("\n[IN] Unexpected message type!", string(msgType), string(msg))
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", address)
	defer conn.Close()
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(conn)
	go ReadMessages(*reader)

	//login first if you have already register
	Login(login, pass, conn)

	declared := false
	ready := false

	for {
		fmt.Printf("\n------------\nYou've logged as %s\nClient menu:\n1) Send message\n2) Check messages\n3) Stream message\n4) Declare and exec my func\n5) Calculate someone's func\n6) Exit\n------------\n\nEnter number: ", login)
		key := 0
		_, err := fmt.Scan(&key)
		if err != nil {
			continue
		}

		switch key {
		case 1:
			SendMessage(conn)
		case 2:
			ListMessage()
		case 3:
			StreamMessage(login, conn)
		case 4:
			if !declared {
				f := "unnamed"
				_, err := fmt.Scan(&f)
				if err != nil {
					panic(err)
				}
				fmt.Println(f)
				declared = Declare(conn, f)
			}
			if !ready {
				ready = true
				go Ready(conn)
			}
		case 5:
			CalculateFunc(conn)
		case 6:
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("No such option!")
		}
	}
}
