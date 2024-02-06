package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	DSN       = "file:./database/data.db"
	TableName = "users"
)

var (
	TableCols = []string{"login", "pass"}
)

type postQuery struct {
	Function string `json:"func"`
}

type calcQuery struct {
	Function string `json:"func"`
	Data     []byte `json:"data"`
}

type logQuery struct {
	Login    string `json:"login"`
	Password string `json:"pass"`
}

type msgQuery struct {
	Receiver string `json:"rec"`
	Message  string `json:"msg"`
}

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

// Filters s to get rid of all escape characters as '\n', '\r', etc...
func filterNewLines(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
			return -1
		default:
			return r
		}
	}, s)
}

// Logins or registers user with login and pass
func (s *Server) loginOrRegister(login, pass string, reg bool, u *user) bool /*, error*/ {
	l := &logQuery{}
	err := s.db.QueryRow("SELECT * FROM "+TableName+" WHERE "+TableCols[0]+"=?", login).Scan(&l.Login, &l.Password)

	if err != nil {
		if err == sql.ErrNoRows {
			if !reg { // can't find such user
				return false
			}
			// register new user
			statement := "INSERT INTO " + TableName + " (" + strings.Join(TableCols, ",") + ") VALUES (" + strings.Join(strings.Split(strings.Repeat("?", len(TableCols)), ""), ",") + ")"
			_, err := s.db.Exec(statement, login, pass)
			if err != nil {
				log.Fatalln(err)
			}
			return true
		}
		log.Fatalln(err)
	}

	if reg {
		return false // user already exists
	}
	if l.Password == pass {
		s.muChans.Lock()
		if _, ok := s.activeUsers[login]; ok { // user has already logged in
			s.muChans.Unlock()
			return false
		}
		s.activeUsers[login] = u // add to active users
		s.muChans.Unlock()
		return true // succsessful login
	}
	return false // wrong password
}

// Closes conn and deletes func from funcMap if exists
func (s *Server) closeConn(conn net.Conn, login *string) {
	name := conn.RemoteAddr().String()
	s.muMap.Lock()
	for key, value := range s.funcMap {
		if value == *login {
			delete(s.funcMap, key)
			break
		}
	}
	s.muMap.Unlock()
	s.muChans.Lock()
	delete(s.activeUsers, *login)
	s.muChans.Unlock()
	conn.Close()
	log.Println(name, "- [OK]: disconnected")
}

func (s *Server) readMsg(name string, scanner bufio.Reader, conn net.Conn) (byte, []byte, error) {
	msgType, err := scanner.ReadByte()
	if err == io.EOF {
		log.Println(name, "- [EOF]")
		return ' ', nil, err
	} else if err != nil {
		s.logError(conn, name, "can't read message type")
		return ' ', nil, err
	}
	buf, err := scanner.ReadBytes('\n') // because of this, please always add '\n' at the end of your message
	if err != nil {
		s.logError(conn, name, "can't read message content!\r\n")
		return ' ', nil, err
	}
	return msgType, buf, nil
}

func (s *Server) handlePost(conn net.Conn, login string, buf []byte) error {
	name := conn.RemoteAddr().String()

	u := &postQuery{}
	err := json.Unmarshal(buf, u)
	if err != nil {
		s.logError(conn, name, "can't unmarshal buf!\r\n")
		return errors.New("can't unmarshal buf") //continue LOOP
	}

	s.muMap.Lock()
	defer s.muMap.Unlock()
	if _, ok := s.funcMap[u.Function]; ok { // if func with this name is already in a map
		s.logError(conn, name, "function with this name is already exists!\r\n")
		return errors.New("function is already exists in map")
	}
	s.funcMap[u.Function] = login
	log.Printf("%s - [OK]: added to map:\n%v\n\t", name, s.funcMap)
	conn.Write([]byte("OFunction was registered\r\n"))
	return nil
}

func (s *Server) logError(conn net.Conn, name, err string) {
	log.Println(name, "- [ERR]: "+err)
	conn.SetWriteDeadline(time.Now().Add(time.Second))
	conn.Write([]byte("E" + err + "\r\n"))
}

func (s *Server) checkLogin(login, name, msgType string, conn net.Conn) bool {
	if msgType != "" {
		log.Println(name, "- [OK]: "+msgType+" request")
	}
	if login == "" {
		s.logError(conn, name, "you should login first!\r\n")
		return false
	}
	return true
}

func (s *Server) signUpUser(name string, buf []byte, conn net.Conn) error {
	r := &logQuery{}
	err := json.Unmarshal(buf, r)
	if err != nil {
		s.logError(conn, name, "server can't unmarshal message content!\r\n")
		return err
	}
	result := s.loginOrRegister(r.Login, r.Password, true, nil)
	if !result || r.Login == "" {
		s.logError(conn, name, "this login has already been taken\r\n")
		return err
	}

	log.Println(name, "- [OK]: registered succsessfully")
	conn.Write([]byte("OSuccsessfully signed up\r\n"))
	return nil
}

func (s *Server) signInUser(name string, buf []byte, u *user, conn net.Conn) (string, error) {
	l := &logQuery{}
	err := json.Unmarshal(buf, l)
	if err != nil {
		s.logError(conn, name, "server can't unmarshal message content!\r\n")
		return "", err
	}
	result := s.loginOrRegister(l.Login, l.Password, false, u)
	if !result {
		s.logError(conn, name, "wrong login/pass or user is online already\r\n")
		return "", err
	}

	log.Println(name, "- [OK]: logged in succsessfully")
	conn.Write([]byte("OSuccsessfully logged in\r\n"))
	return l.Login, nil
}

func (s *Server) sendMessage(name, login string, buf []byte, conn net.Conn) error {
	m := &msgQuery{}
	err := json.Unmarshal(buf, m)
	if err != nil {
		s.logError(conn, name, "server can't unmarshal message content!\r\n")
		return err
	}
	s.muChans.Lock()
	resChan, ok := s.activeUsers[m.Receiver]
	if !ok { // if user-receiver is not logged in
		log.Println(name, "- [ERR]: user", m.Receiver, "is not logged in!")
		conn.Write([]byte("EReceiver is not logged in!\r\n"))
		s.muChans.Unlock()
		return nil
	}
	m.Receiver = login
	msg, _ := json.Marshal(m)

	resChan.conn.Write([]byte("M"))
	resChan.conn.Write(msg) // error check
	resChan.conn.Write([]byte("\n"))

	s.muChans.Unlock()
	conn.Write([]byte("OMessage was sent\r\n"))
	return nil
}

func (s *Server) streamMessage(login string, buf []byte, conn net.Conn) {
	s.muChans.Lock()
	for key, value := range s.activeUsers {
		if key == login {
			continue
		}
		value.conn.Write([]byte("M"))
		value.conn.Write(buf) // error check here?
	}
	s.muChans.Unlock()
	conn.Write([]byte("OMessage was streamed\r\n"))
}

func (s *Server) handleCalc(name string, buf []byte, dataChan chan string, conn net.Conn) error {
	c := &calcQuery{}
	err := json.Unmarshal(buf, c) // get params
	if err != nil {
		s.logError(conn, name, "server can't unmarshal message content!\r\n")
		return err
	}

	s.muMap.Lock()
	handler, ok := s.funcMap[c.Function] // find func
	if !ok {
		s.muMap.Unlock()
		log.Println(name, "- [ERR]: can't find function", c.Function)
		conn.Write([]byte("EThis function wasn't registered on server!\r\n"))
		return nil
	}
	s.muChans.Lock() // mutex here, so nobody can't rebind outChan until operation is done
	s.activeUsers[handler].chans.out = &dataChan
	sendChan := s.activeUsers[handler].chans.in
	s.muMap.Unlock()

	*sendChan <- /*"C" + */ string(c.Data) + "\n" // send params to func holder
	// TODO timeout here
	// select time.After
	answer := <-dataChan // get answer
	s.muChans.Unlock()

	log.Println(name, "- [OK]: operation was done")
	conn.Write([]byte(answer)) // send answer to conn with 'D' header
	return nil
}

func (s *Server) handleReady(conn net.Conn, dataChan chan string, resultChan chan []byte, scanner bufio.Reader, u *user) {
	for {
		data := <-dataChan // get values from chan
		conn.Write([]byte("C" + data))
		// timeout?
		ans := <-resultChan
		*u.chans.out <- "D" + string(ans) // send ans to out chan
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	name := conn.RemoteAddr().String()
	log.Println(name, "- [OK]: connected")

	scanner := bufio.NewReader(conn)
	dataChan := make(chan string)
	resultChan := make(chan []byte)
	u := &user{
		conn:  conn,
		chans: &inOutChans{in: &dataChan, out: nil},
	}
	var login string
	ready := false
	defer s.closeConn(conn, &login)
LOOP:
	for {
		msgType, buf, err := s.readMsg(name, *scanner, conn)
		if err != nil {
			break
		}
		log.Printf("%s - type: %s, data: %s\n", name, string(msgType), filterNewLines(string(buf)))

		switch msgType {
		case 'D': // DONE
			resultChan <- buf

		case 'U': // SIGN UP
			log.Println(name, "- [OK]: UP request")
			if err := s.signUpUser(name, buf, conn); err != nil {
				break LOOP
			}

		case 'I': // SIGN IN
			log.Println(name, "- [OK]: IN request")
			if l, err := s.signInUser(name, buf, u, conn); err != nil {
				break LOOP
			} else {
				login = l
			}

		case 'M': // MESSAGE
			if ok := s.checkLogin(login, name, "MSG", conn); !ok {
				break LOOP
			}
			if err := s.sendMessage(name, login, buf, conn); err != nil {
				break LOOP
			}

		case 'S': // STREAM
			if ok := s.checkLogin(login, name, "STR", conn); !ok {
				break LOOP
			}
			s.streamMessage(login, buf, conn)

		case 'C': // CALC - conn asks to calculate calcQuery.Function with parameters stored in calcQuery.Data
			if ok := s.checkLogin(login, name, "CALC", conn); !ok {
				break LOOP
			}
			if err := s.handleCalc(name, buf, dataChan, conn); err != nil {
				break LOOP
			}

		case 'P': // POST - conn tries to declare it's postQuery.Function on server
			if ok := s.checkLogin(login, name, "POST", conn); !ok {
				break LOOP
			}
			if err := s.handlePost(conn, login, buf); err != nil {
				if err.Error() != "function is already exists in map" {
					break LOOP
				}
			}

		case 'R': // READY - conn is now ready to execute others CALC requests
			if ok := s.checkLogin(login, name, "READY", conn); !ok {
				break LOOP
			}

			if ready {
				continue LOOP
			}
			ready = true

			go s.handleReady(conn, dataChan, resultChan, *scanner, u)

		default:
			s.logError(conn, name, "wrong message type!\r\n")
		}
	}
}
