# agentcast

CLI tool for recording and sharing agent terminal sessions.

## Build & Run

```bash
make build        # build binary to ./bin/cast
make install      # install to $GOPATH/bin
go run ./cmd/cast  # run directly
```

## Test

```bash
make test          # run all tests
go test ./internal/asciicast/...  # test specific package
```

## Lint

```bash
make lint          # requires golangci-lint
```

## Project Structure

- `cmd/cast/` — entry point
- `internal/cmd/` — cobra command definitions
- `internal/asciicast/` — asciicast v2 format library
- `internal/recorder/` — PTY recording engine
- `internal/player/` — terminal playback engine
- `internal/storage/` — local recording management
- `internal/upload/` — push/share HTTP client
