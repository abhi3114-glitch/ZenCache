# ZenCache

A high-performance distributed in-memory cache server inspired by Redis, built from scratch in Go.

## Overview

ZenCache is a lightweight, Redis-compatible caching solution that provides fast key-value storage with advanced features including automatic memory management through LRU eviction, real-time messaging via Pub/Sub, data persistence, and horizontal scaling through replication.

## Features

### Core Capabilities

- **Key-Value Storage**: Thread-safe operations with O(1) average time complexity for SET, GET, and DEL commands
- **LRU Eviction**: Automatic memory management using a doubly linked list and hashmap combination, ensuring O(1) eviction when capacity is exceeded
- **Pub/Sub Messaging**: Real-time channel-based messaging allowing multiple subscribers to receive published messages instantly
- **RDB Persistence**: Snapshot-based persistence using binary encoding, with automatic loading on server startup
- **Master-Replica Replication**: Asynchronous replication with automatic command propagation from master to replicas

### Technical Highlights

- Pure Go implementation with no external dependencies beyond the standard library
- Concurrent connection handling using goroutines
- Thread-safe data structures protected by read-write mutexes
- Simple text-based protocol for easy debugging and integration

## Installation

```bash
git clone https://github.com/abhi3114-glitch/ZenCache.git
cd ZenCache
go build -o zencache.exe .
```

## Usage

### Starting the Server

```bash
# Default configuration (port 6379, capacity 10000 items)
./zencache.exe

# Custom port and capacity
./zencache.exe -port 6380 -capacity 50000
```

### Command Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `-port` | 6379 | TCP port to listen on |
| `-capacity` | 10000 | Maximum number of items before LRU eviction |

### Connecting to the Server

Use netcat, telnet, or any TCP client:

```bash
nc localhost 6379
```

## Command Reference

### Data Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| SET | `SET key value` | Store a key-value pair |
| GET | `GET key` | Retrieve value by key (returns `(nil)` if not found) |
| DEL | `DEL key` | Delete a key (returns count of deleted keys) |
| PING | `PING` | Health check (returns `PONG`) |

### Pub/Sub Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| SUBSCRIBE | `SUBSCRIBE channel` | Subscribe to a channel for messages |
| UNSUBSCRIBE | `UNSUBSCRIBE channel` | Unsubscribe from a channel |
| PUBLISH | `PUBLISH channel message` | Publish a message to all subscribers |

### Persistence Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| SAVE | `SAVE` | Create an RDB snapshot to disk |

### Replication Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| REPLICAOF | `REPLICAOF host port` | Configure this instance as a replica |
| INFO | `INFO` | Display server role and replica count |

### Connection Commands

| Command | Syntax | Description |
|---------|--------|-------------|
| QUIT | `QUIT` | Close the connection |

## Examples

### Basic Key-Value Operations

```
> SET user:1001 "John Doe"
OK
> GET user:1001
John Doe
> DEL user:1001
(integer) 1
> GET user:1001
(nil)
```

### Pub/Sub Messaging

Terminal 1 (Subscriber):
```
> SUBSCRIBE notifications
SUBSCRIBED notifications
MESSAGE notifications Hello subscribers!
```

Terminal 2 (Publisher):
```
> PUBLISH notifications Hello subscribers!
(integer) 1
```

### Setting Up Replication

Master Instance:
```bash
./zencache.exe -port 6379
```

Replica Instance:
```bash
./zencache.exe -port 6380
```

On the replica, run:
```
> REPLICAOF localhost 6379
OK
> INFO
role:replica
replicas:0
```

All write operations (SET, DEL) on the master will automatically propagate to connected replicas.

## Architecture

```
ZenCache/
├── main.go                 # Entry point and CLI flag parsing
├── server/
│   └── server.go           # TCP server and command dispatcher
├── lru/
│   ├── lru.go              # LRU cache implementation
│   └── lru_test.go         # LRU unit tests
├── pubsub/
│   ├── pubsub.go           # Pub/Sub messaging system
│   └── pubsub_test.go      # Pub/Sub unit tests
├── rdb/
│   ├── rdb.go              # RDB persistence layer
│   └── rdb_test.go         # Persistence unit tests
├── repl/
│   └── repl.go             # Replication manager
└── integration_test.go     # End-to-end integration tests
```

### Component Details

- **Server**: Handles TCP connections, parses commands, and routes to appropriate handlers
- **LRU Cache**: Maintains insertion order using a doubly linked list with O(1) access via hashmap
- **Pub/Sub**: Manages channel subscriptions with buffered channels for message delivery
- **RDB**: Serializes cache data using Go's gob encoding for efficient binary storage
- **Replication**: Manages master-replica connections and propagates write commands

## Testing

Run the complete test suite:

```bash
go test -v ./...
```

Run specific package tests:

```bash
go test -v ./lru/...
go test -v ./pubsub/...
go test -v ./rdb/...
```

## Performance Considerations

- LRU operations (get, set, eviction) are O(1) time complexity
- Read operations use read locks allowing concurrent access
- Write operations use exclusive locks for thread safety
- Pub/Sub uses buffered channels (100 messages) to prevent blocking publishers
- Replication is asynchronous to avoid impacting master performance

## Limitations

- No TTL (time-to-live) support for automatic key expiration
- No clustering support (single master only)
- No authentication mechanism
- Persistence is manual (no automatic background saves)
- Protocol is simplified text-based, not full RESP compatible

## License

This project is available for educational and personal use.

## Contributing

Contributions are welcome. Please ensure all tests pass before submitting pull requests.
