package main_test

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"
	"zencache/server"
)

func TestCoreOperations(t *testing.T) {
	// Start server in a goroutine
	port := 6380
	srv := server.NewServer(port)
	go func() {
		srv.Start()
	}()

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		t.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	sendCommand := func(cmd string) string {
		_, err := writer.WriteString(cmd + "\n")
		if err != nil {
			t.Fatalf("Failed to write command: %v", err)
		}
		writer.Flush()
		resp, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("Failed to read response: %v", err)
		}
		return strings.TrimSpace(resp)
	}

	// Test PING
	if resp := sendCommand("PING"); resp != "PONG" {
		t.Errorf("Expected PONG, got %s", resp)
	}

	// Test SET
	if resp := sendCommand("SET mykey myvalue"); resp != "OK" {
		t.Errorf("Expected OK, got %s", resp)
	}

	// Test GET
	if resp := sendCommand("GET mykey"); resp != "myvalue" {
		t.Errorf("Expected myvalue, got %s", resp)
	}

	// Test GET non-existent
	if resp := sendCommand("GET nokey"); resp != "(nil)" {
		t.Errorf("Expected (nil), got %s", resp)
	}

	// Test DEL
	if resp := sendCommand("DEL mykey"); resp != "(integer) 1" {
		t.Errorf("Expected (integer) 1, got %s", resp)
	}

	// Test GET after DEL
	if resp := sendCommand("GET mykey"); resp != "(nil)" {
		t.Errorf("Expected (nil), got %s", resp)
	}
}
