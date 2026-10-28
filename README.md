# Go Deep Dive

Three runnable, standard-library-only Go projects that progress from streaming I/O to concurrency to persistent HTTP services. Each project is a separate executable in one Go module. Requires **Go 1.22+**.

| Project | Build | Concepts |
| --- | --- | --- |
| [01 · Loglens](01-loglens/) | Streaming log analyzer | Interfaces, maps, scanning, errors, CLI flags |
| [02 · Hashpool](02-hashpool/) | Concurrent SHA-256 file hasher | Goroutines, channels, ownership, cancellation, synchronization |
| [03 · PocketKV](03-pocketkv/) | Persistent HTTP key-value API | HTTP, JSON validation, mutexes, atomic snapshots, shutdown |

## Start here

Run commands from this directory:

```sh
go run ./01-loglens -file ./01-loglens/sample.log
go run ./02-hashpool -workers 2 ./README.md ./go.mod
go test -race ./...
go vet ./...
```

Start the API:

```sh
go run ./03-pocketkv -data store.json
```

In another terminal:

```sh
curl -i -X PUT http://127.0.0.1:8080/kv/name \
  -H 'Content-Type: application/json' -d '{"value":"gopher"}'
curl http://127.0.0.1:8080/kv/name
```

Stop with Ctrl-C, restart, then repeat the GET to see persistence.

## Learn deeply, one change at a time

1. Run each project and its tests. Read its README before reading its source.
2. Draw the data flow. Identify who owns each file, map, channel, and lock.
3. Explain every `defer`, error return, and goroutine aloud.
4. Pick one extension from its README. Write a behavioral test before implementing it.
5. Record a measurement or design tradeoff in your README and commit it.

The implementations are deliberately small enough to understand end to end. They are learning foundations; the exercises are where you make them your own.

## Build binaries

```sh
mkdir -p bin
go build -o bin/loglens ./01-loglens
go build -o bin/hashpool ./02-hashpool
go build -o bin/pocketkv ./03-pocketkv
```

## Share on GitHub

This collection can be one repository. Before publishing, add your name, your own extensions, screenshots or terminal recordings, and a short explanation of a bug you fixed. Choose a license that matches how you want others to use your work.

For a standalone repository, copy this entire directory **outside any existing Git repository**, then run inside the copy:

```sh
git init
git add .
git commit -m "Add three Go learning projects"
git branch -M main
```

Create an empty repository on GitHub, then use its actual URL:

```sh
git remote add origin https://github.com/YOUR_USERNAME/go-deep-dive.git
git push -u origin main
```

The module path is `github.com/yash27op/go-deep-dive`.

## Checks

GitHub Actions runs formatting checks, `go vet`, and race-enabled tests. Locally, run `make check` (or the Go commands directly). No external packages or service credentials are needed.
