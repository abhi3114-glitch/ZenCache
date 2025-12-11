package server

import (
	"bufio"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"zencache/lru"
	"zencache/pubsub"
	"zencache/rdb"
	"zencache/repl"
)

type Server struct {
	port     int
	cache    *lru.Cache
	pubsub   *pubsub.PubSub
	rdb      *rdb.RDB
	repl     *repl.ReplicationManager
	clientID uint64
}

func NewServer(port int) *Server {
	return &Server{
		port:   port,
		cache:  lru.NewCache(10000),
		pubsub: pubsub.NewPubSub(),
		rdb:    rdb.NewRDB("zencache.rdb"),
		repl:   repl.NewReplicationManager(),
	}
}

func NewServerWithCapacity(port int, capacity int) *Server {
	return &Server{
		port:   port,
		cache:  lru.NewCache(capacity),
		pubsub: pubsub.NewPubSub(),
		rdb:    rdb.NewRDB("zencache.rdb"),
		repl:   repl.NewReplicationManager(),
	}
}

func (s *Server) Start() error {
	// Try to load from RDB on startup
	if data, err := s.rdb.Load(); err == nil {
		s.cache.LoadData(data)
		fmt.Println("Loaded data from RDB snapshot")
	}

	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		clientID := atomic.AddUint64(&s.clientID, 1)
		go s.handleConnection(conn, fmt.Sprintf("client-%d", clientID))
	}
}

// ApplyCommand applies a command directly (used for replication).
func (s *Server) ApplyCommand(cmd string) {
	parts := strings.Split(cmd, " ")
	if len(parts) < 1 {
		return
	}
	command := strings.ToUpper(parts[0])

	switch command {
	case "SET":
		if len(parts) >= 3 {
			key := parts[1]
			val := strings.Join(parts[2:], " ")
			s.cache.Set(key, val)
		}
	case "DEL":
		if len(parts) >= 2 {
			s.cache.Del(parts[1])
		}
	}
}

func (s *Server) handleConnection(conn net.Conn, clientID string) {
	defer conn.Close()
	defer s.pubsub.UnsubscribeAll(clientID)

	reader := bufio.NewReader(conn)
	isReplica := false

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			if isReplica {
				s.repl.RemoveReplica(conn)
			}
			return
		}

		message = strings.TrimSpace(message)
		if len(message) == 0 {
			continue
		}

		parts := strings.Split(message, " ")
		cmd := strings.ToUpper(parts[0])
		var output string

		switch cmd {
		case "SET":
			if len(parts) < 3 {
				output = "(error) ERR wrong number of arguments for 'set' command\n"
			} else {
				key := parts[1]
				val := strings.Join(parts[2:], " ")
				s.cache.Set(key, val)
				output = "OK\n"
				// Propagate to replicas
				if s.repl.IsMaster() {
					s.repl.PropagateCommand(message)
				}
			}

		case "GET":
			if len(parts) < 2 {
				output = "(error) ERR wrong number of arguments for 'get' command\n"
			} else {
				val, found := s.cache.Get(parts[1])
				if !found {
					output = "(nil)\n"
				} else {
					output = fmt.Sprintf("%s\n", val)
				}
			}

		case "DEL":
			if len(parts) < 2 {
				output = "(error) ERR wrong number of arguments for 'del' command\n"
			} else {
				deleted := s.cache.Del(parts[1])
				if deleted {
					output = "(integer) 1\n"
					// Propagate to replicas
					if s.repl.IsMaster() {
						s.repl.PropagateCommand(message)
					}
				} else {
					output = "(integer) 0\n"
				}
			}

		case "PING":
			output = "PONG\n"

		case "SUBSCRIBE":
			if len(parts) < 2 {
				output = "(error) ERR wrong number of arguments for 'subscribe' command\n"
			} else {
				channel := parts[1]
				sub := s.pubsub.Subscribe(channel, clientID)
				output = fmt.Sprintf("SUBSCRIBED %s\n", channel)

				go func(sub *pubsub.Subscriber, ch string) {
					for msg := range sub.Messages {
						conn.Write([]byte(fmt.Sprintf("MESSAGE %s %s\n", ch, msg)))
					}
				}(sub, channel)
			}

		case "UNSUBSCRIBE":
			if len(parts) < 2 {
				output = "(error) ERR wrong number of arguments for 'unsubscribe' command\n"
			} else {
				channel := parts[1]
				s.pubsub.Unsubscribe(channel, clientID)
				output = fmt.Sprintf("UNSUBSCRIBED %s\n", channel)
			}

		case "PUBLISH":
			if len(parts) < 3 {
				output = "(error) ERR wrong number of arguments for 'publish' command\n"
			} else {
				channel := parts[1]
				msg := strings.Join(parts[2:], " ")
				count := s.pubsub.Publish(channel, msg)
				output = fmt.Sprintf("(integer) %d\n", count)
			}

		case "SAVE":
			err := s.rdb.Save(s.cache.GetAllData())
			if err != nil {
				output = fmt.Sprintf("(error) %v\n", err)
			} else {
				output = "OK\n"
			}

		case "REPLICAOF":
			if len(parts) < 3 {
				output = "(error) ERR wrong number of arguments for 'replicaof' command\n"
			} else {
				host := parts[1]
				port, err := strconv.Atoi(parts[2])
				if err != nil {
					output = "(error) ERR invalid port\n"
				} else {
					err = s.repl.ConnectToMaster(host, port, s.ApplyCommand)
					if err != nil {
						output = fmt.Sprintf("(error) %v\n", err)
					} else {
						output = "OK\n"
					}
				}
			}

		case "REPLCONF":
			// This is sent by replicas during handshake
			isReplica = true
			s.repl.AddReplica(conn)
			output = "OK\n"

		case "INFO":
			output = fmt.Sprintf("role:%s\nreplicas:%d\n", s.repl.Role(), s.repl.ReplicaCount())

		case "QUIT":
			return

		default:
			output = fmt.Sprintf("(error) ERR unknown command '%s'\n", cmd)
		}

		conn.Write([]byte(output))
	}
}
