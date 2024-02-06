package main

import (
	"database/sql"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

const (
	logPath = "./logs/"
	address = ":8080"
)

// ////////
const (
	DSN       = "file:./database/data.db"
	TableName = "users"
)

type inOutChans struct {
	in  *chan string
	out *chan string
}

type user struct {
	chans *inOutChans
	conn  net.Conn
}

type Server struct {
	db *sql.DB // registeredUsers database

	activeUsers map[string]*user  // login -> it's conn chans and conn itself
	funcMap     map[string]string // func -> login that can handle it

	muMap   *sync.Mutex // locks all r/w operations with funcMap // TODO: RWMutex?
	muChans *sync.Mutex // locks all r/w operations with activeUsers
}

func NewServer() *Server {
	db, err := sql.Open("sqlite3", DSN)
	if err != nil {
		log.Fatalln(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	return &Server{
		db:          db, // registeredUsers database
		activeUsers: make(map[string]*user),
		funcMap:     make(map[string]string),
		muMap:       &sync.Mutex{},
		muChans:     &sync.Mutex{},
	}
}

func setupLogger() *os.File {
	logFile, err := os.OpenFile(logPath+time.Now().String()+".log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("server - [ERR]: Cannot open log file: %v\n", err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	return logFile
}

func main() {
	logFile := setupLogger()
	defer logFile.Close()
	log.Println("Starting server...")

	listner, err := net.Listen("tcp", address)
	if err != nil {
		panic(err)
	}

	server := NewServer()
	defer server.db.Close()
	for {
		conn, err := listner.Accept()
		if err != nil {
			panic(err)
		}
		go server.handleConnection(conn)
	}
}
