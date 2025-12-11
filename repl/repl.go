package repl

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

// ReplicationManager handles master-replica communication.
type ReplicationManager struct {
	mu         sync.RWMutex
	role       string // "master" or "replica"
	replicas   []net.Conn
	masterConn net.Conn
}

// NewReplicationManager creates a new replication manager.
func NewReplicationManager() *ReplicationManager {
	return &ReplicationManager{
		role:     "master",
		replicas: make([]net.Conn, 0),
	}
}

// Role returns the current role.
func (r *ReplicationManager) Role() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.role
}

// IsMaster returns true if this is a master node.
func (r *ReplicationManager) IsMaster() bool {
	return r.Role() == "master"
}

// AddReplica adds a replica connection.
func (r *ReplicationManager) AddReplica(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.replicas = append(r.replicas, conn)
}

// RemoveReplica removes a replica connection.
func (r *ReplicationManager) RemoveReplica(conn net.Conn) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, c := range r.replicas {
		if c == conn {
			r.replicas = append(r.replicas[:i], r.replicas[i+1:]...)
			return
		}
	}
}

// PropagateCommand sends a command to all replicas.
func (r *ReplicationManager) PropagateCommand(cmd string) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, conn := range r.replicas {
		go func(c net.Conn) {
			c.Write([]byte(cmd + "\n"))
		}(conn)
	}
}

// ReplicaCount returns the number of connected replicas.
func (r *ReplicationManager) ReplicaCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.replicas)
}

// ConnectToMaster connects to a master server as a replica.
func (r *ReplicationManager) ConnectToMaster(host string, port int, applyCmd func(string)) error {
	r.mu.Lock()
	r.role = "replica"
	r.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.masterConn = conn
	r.mu.Unlock()

	// Send REPLCONF to identify as a replica
	conn.Write([]byte("REPLCONF listening-port\n"))

	// Start goroutine to receive commands from master
	go func() {
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				return
			}
			line = strings.TrimSpace(line)
			if len(line) > 0 {
				applyCmd(line)
			}
		}
	}()

	return nil
}

// Close closes all replica connections.
func (r *ReplicationManager) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, conn := range r.replicas {
		conn.Close()
	}
	if r.masterConn != nil {
		r.masterConn.Close()
	}
}
