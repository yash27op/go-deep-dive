# 01 · Loglens

Analyze `LEVEL message` logs without loading the full file into memory.

```sh
go run ./01-loglens -file ./01-loglens/sample.log
printf 'INFO hello\nERROR failed\n' | go run ./01-loglens
go test -v ./01-loglens
```

Commands assume the collection root. The sample produces six lines, two INFO, one DEBUG, one WARN, one ERROR, and one malformed line. Levels are case sensitive. Blank lines and messages without a supported level count as malformed. Lines larger than the scanner's roughly 1 MiB buffer limit return an error and the CLI exits unsuccessfully.

## Read the code through these questions

- Why does `Analyze` accept `io.Reader` instead of a filename?
- Why initialize the map before incrementing a value? What does a missing map entry return?
- Why check `scanner.Err()` after the loop?
- What memory grows with input size? What remains bounded?
- Why does `run` own file cleanup while `main` owns the exit code?

## Make it yours

1. Add `-strict` to fail on the first malformed line; test the line number.
2. Introduce a `Parser` interface and implement both plain text and JSON log parsers.
3. Add timestamps and per-minute error rates. Decide how to handle out-of-order input.
4. Write `BenchmarkAnalyze` using `testing.B`; run `go test -bench . -benchmem ./01-loglens`. Compare `strings.Fields` with `strings.Cut` and explain the allocation difference.

Portfolio evidence: include before/after benchmark results and document your accepted input grammar.
