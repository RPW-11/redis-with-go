# redis-with-go

A Redis server implementation from scratch in Go. Supports a subset of the Redis command set over a TCP connection using the RESP3 protocol, with LRU eviction and append-only file (AOF) persistence.

## Getting started

**Requirements:** Go 1.26+

```bash
make build
make run
```

By default the server listens on port `6379`. Available flags:

| Flag | Default | Description |
|---|---|---|
| `-port` | `6379` | TCP port to listen on |
| `-max-cap` | `1000000` | Max number of keys before LRU eviction |
| `-aof-dir` | `.` | Directory to write `data.aof` (also replayed on startup) |
| `-no-aof` | `false` | Disable AOF persistence |

```bash
./main -port 6380 -max-cap 500000 -aof-dir /var/lib/redis
```

Connect with any Redis client or `redis-cli`:

```bash
redis-cli -p 6379
```

## Supported commands

| Command | Syntax | Description |
|---|---|---|
| `SET` | `SET key value` | Set a key to a string value |
| `GET` | `GET key` | Get the value of a key |
| `DEL` | `DEL key` | Delete a key |
| `EXPIRE` | `EXPIRE key seconds` | Set a TTL on a key in seconds |
| `EXPIREAT` | `EXPIREAT key unix-ts` | Set expiry as a Unix timestamp |
| `TTL` | `TTL key` | Get remaining TTL; `-1` if no expiry, `-2` if key missing |

## Architecture

```
TCP listener
    └── handle goroutine (per connection)
            └── bufio.Reader (one per connection, reused across commands)
            └── command.Handle — parses one RESP3 command, dispatches, returns *resp.Value
                    └── store.Store — mutex + LRU eviction + AOF logging
                            └── LRUEngine — pure LRU mechanics (doubly linked list + map)
                            └── AofLogger — appends RESP-encoded commands to data.aof
```

Key design decisions:

- **`Store`** owns the mutex, AOF logging, and lazy expiration — all in one locked section per operation.
- **`LRUEngine`** has no locking or I/O awareness; it is only called from within `Store`'s locked sections.
- **`AofLogger`** is write-only and needs no mutex of its own for the same reason.
- **AOF replay** runs before the logger is attached (`s.aof == nil`), so replayed writes never get re-appended to the file.
- **`command.Handle`** takes a `*bufio.Reader` owned by the caller. The server creates one per connection and reuses it across commands, avoiding buffer loss between calls.
- The server catches `SIGTERM`/`SIGINT` for graceful shutdown: the listener is closed, in-flight commands complete, and idle connections are dropped.

## Testing

```bash
make test
```

Tests run with the `-race` flag enabled by default.
