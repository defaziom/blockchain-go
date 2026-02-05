# CLAUDE.md - AI Assistant Guide for blockchain-go

## Project Overview

A blockchain implementation in Go that demonstrates peer-to-peer networking, consensus mechanisms, and proof-of-work mining. The project implements a distributed blockchain where peers communicate via TCP and expose a REST API for client interaction.

## Quick Reference

```bash
# Build
go build -o blockchain-go .

# Run
./blockchain-go <http_port> <tcp_port>
# Example: ./blockchain-go 8081 1111

# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./blockchain -v
```

## Architecture

### Component Overview

The application runs three concurrent services:
1. **HTTP Server** - REST API for client interaction (blocks, mining, peer management)
2. **TCP Server** - Peer-to-peer communication for blockchain synchronization
3. **Task Processor** - Handles peer interaction commands via the Command pattern

### Package Structure

```
blockchain-go/
├── main.go              # Entry point, starts all services
├── block/               # Block data structure and hashing
│   ├── block.go         # Block struct, hash calculation, validation
│   └── block_test.go
├── blockchain/          # Blockchain management and consensus
│   ├── blockchain.go    # Chain operations, difficulty adjustment, validation
│   └── blockchain_test.go
├── database/            # In-memory peer storage (hashicorp/go-memdb)
│   ├── database.go      # Singleton DB instance
│   ├── schema.go        # memdb schema definition
│   ├── dao.go           # Data access functions
│   └── models.go        # PeerConnInfo model
├── http/                # REST API
│   ├── server.go        # HTTP server setup and routing
│   ├── handlers.go      # Request handlers (blocks, mining, peers)
│   └── middleware.go    # JSON response, request logging
├── task/                # Command pattern for peer interactions
│   ├── tasks.go         # Job executor, task definitions
│   └── tasks_test.go
└── tcp/                 # TCP peer networking
    ├── tcp_server.go    # TCP listener
    ├── tcp_peer.go      # Peer connection management, dialing
    ├── tcp_msg.go       # Message types and protocol
    ├── tcp_msg_test.go
    └── tcp_peer_test.go
```

## Key Interfaces

### BlockChain Interface (blockchain/blockchain.go:105)

```go
type BlockChain interface {
    MineBlock(data string) *block.Block
    AddBlock(block *block.Block) error
    GetBlocks() *SafeDoublyLinkedBlockList
    GetDifficulty() int
    GetAdjustedDifficulty() int
    GetCumulativeDifficulty() float64
    GetLatestBlock() *block.Block
    ReplaceChain(newChain BlockChain)
}
```

### Peer Interface (tcp/tcp_msg.go:27)

```go
type Peer interface {
    ClosePeer() error
    IsClosed() bool
    ReceiveMsg() (*PeerMsg, error)
    SendResponseBlockChainMsg(blocks []*block.Block) error
    SendQueryAllMsg() error
    SendAckMsg() error
}
```

### Task System Interfaces (task/tasks.go:34-47)

```go
type JobExecutor interface {
    Start() error
}

type Job interface {
    GetNextTask() (Task, error)
}

type Task interface {
    Execute() error
}
```

## Design Patterns

### Command Pattern (task package)
The `task` package implements the Command pattern for peer interactions:
- **JobExecutor** - Runs jobs by executing tasks sequentially
- **Job** - Produces tasks based on incoming peer messages
- **Task** - Encapsulates a single operation (Ack, QueryLatest, QueryAll, ResponseBlockChain)

### Singleton Pattern (database package)
The database uses a singleton pattern via `GetDatabase()` to ensure a single memdb instance.

### Thread-Safe Linked List (blockchain package)
`SafeDoublyLinkedBlockList` provides mutex-protected operations for concurrent blockchain access.

## REST API Endpoints

| Method | Endpoint       | Description                    |
|--------|----------------|--------------------------------|
| GET    | /blocks        | Returns all blocks in chain    |
| POST   | /blocks/mine   | Mine a new block with data     |
| GET    | /peers         | List registered peers          |
| POST   | /peers         | Register a new peer            |

### Request/Response Examples

```bash
# Mine a block
curl -X POST http://localhost:8081/blocks/mine \
  -H "Content-Type: application/json" \
  -d '{"Data": "my block data"}'

# Register a peer
curl -X POST http://localhost:8081/peers \
  -H "Content-Type: application/json" \
  -d '{"Ip": "192.168.1.100", "Port": 1111}'
```

## TCP Message Protocol

Messages are JSON-encoded with newline terminator (`\n`). Types defined in `tcp/tcp_msg.go`:

| Type               | Value | Purpose                              |
|--------------------|-------|--------------------------------------|
| ACK                | 0     | Signals end of communication         |
| QUERY_LATEST       | 1     | Request latest block                 |
| QUERY_ALL          | 2     | Request entire blockchain            |
| RESPONSE_BLOCKCHAIN| 3     | Contains block(s) in response        |

## Blockchain Mechanics

### Proof of Work
- Blocks require hash prefix of `Difficulty` zeros
- Mining iterates nonce until valid hash found
- Hash: SHA-256 of (timestamp + index + data + nonce)

### Difficulty Adjustment
- Adjusts every `DifficultyAdjustmentIntervalBlocks` (5 blocks)
- Target block time: `BlockGenerationIntervalSec` (0.5 seconds)
- Increases difficulty if blocks generated too fast (< expected/2)
- Decreases difficulty if blocks generated too slow (> expected*2)

### Chain Selection
Longest valid chain wins, measured by cumulative difficulty: `sum(2^difficulty)` for all blocks.

## Code Conventions

### Package Organization
- One package per directory
- Package name matches directory name
- Test files use `_test.go` suffix in same package

### Naming
- Interfaces: descriptive names without `I` prefix (`BlockChain`, `Peer`, `Task`)
- Implementations: suffix with `Impl` or descriptive name (`BlockChainIml`, `PeerConn`)
- Exported functions: PascalCase
- Private functions: camelCase

### Error Handling
- Return errors from functions, handle at call site
- Log errors with context using `log.Println()`
- HTTP handlers return appropriate status codes (400, 500)

### Concurrency
- Use channels for goroutine communication (`chan tcp.Peer`)
- Use `sync.Mutex` for shared data protection
- Start long-running services in goroutines from main

## Testing Conventions

### Test Framework
- Standard `testing` package
- `github.com/stretchr/testify/assert` for assertions
- `github.com/stretchr/testify/mock` for mocking

### Mock Naming
- Prefix mock types with `Mock` (e.g., `MockPeer`, `MockBlockChain`)
- Embed the interface being mocked

### Test Structure
```go
func TestFunctionName(t *testing.T) {
    // Setup mocks
    mPeer := &MockPeer{}
    mPeer.On("Method").Return(value, nil)

    // Execute
    result := FunctionUnderTest(mPeer)

    // Assert
    assert.Equal(t, expected, result)
    mPeer.AssertExpectations(t)
}
```

### Subtests
Use `t.Run()` for testing multiple scenarios:
```go
t.Run("scenario name", func(t *testing.T) {
    // test code
})
```

## Dependencies

- `github.com/hashicorp/go-memdb` - In-memory database for peer storage
- `github.com/stretchr/testify` - Testing assertions and mocking

## Common Development Tasks

### Adding a New Task Type
1. Define new `PeerMsgType` constant in `tcp/tcp_msg.go`
2. Create task struct embedding `*PeerMsgTask` in `task/tasks.go`
3. Implement `Execute() error` method
4. Add case to `GetNextTask()` switch statement
5. Add corresponding `Send*Msg()` method to `Peer` interface if needed

### Adding a New API Endpoint
1. Create handler function in `http/handlers.go` returning `http.Handler`
2. Register route in `http/server.go` with middleware chain
3. Add to Postman collection if applicable

### Modifying Block Structure
1. Update `Block` struct in `block/block.go`
2. Update `CalculateBlockHash()` if new field affects hash
3. Update validation in `IsNewBlockValid()`
4. Update tests in `block/block_test.go`

## File Locations Quick Reference

| Concern                  | File                          |
|--------------------------|-------------------------------|
| Entry point              | main.go                       |
| Block structure          | block/block.go:9              |
| Blockchain interface     | blockchain/blockchain.go:105  |
| Genesis block            | blockchain/blockchain.go:88   |
| Difficulty constants     | blockchain/blockchain.go:98   |
| Peer interface           | tcp/tcp_msg.go:27             |
| Message types            | tcp/tcp_msg.go:13             |
| Task interface           | task/tasks.go:44              |
| HTTP routes              | http/server.go:12             |
| Database schema          | database/schema.go:5          |
