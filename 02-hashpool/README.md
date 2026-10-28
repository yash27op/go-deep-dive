# 02 · Hashpool

Compute SHA-256 hashes using a bounded number of file readers. Output is JSON sorted by path; a failed file has an `error` field and makes the CLI exit unsuccessfully after collecting the other results.

```sh
go run ./02-hashpool -workers 2 README.md go.mod
go test -race -v ./02-hashpool
```

## Follow the pipeline

The producer sends paths to `jobs`. Workers consume paths and send results. A coordinator waits for every worker before closing `results`. The caller drains results until closure, then sorts them. Ctrl-C cancels the context.

- Why does only the producer close `jobs`?
- Why must the coordinator wait before closing `results`?
- What happens if the caller waits for workers before reading results?
- Why does each send select on cancellation?
- Why is there no mutex protecting the output slice?
- Is hashing faster with 100 workers? Measure on your disk.

Reads check cancellation between 32 KiB chunks. Cancellation cannot interrupt an operating-system read already in progress. Use regular local files; FIFOs or devices can block. This version accepts explicit paths, retains all results in memory, and does not guarantee a consistent snapshot of a file modified during hashing.

## Make it yours

1. Add recursive directory traversal with `filepath.WalkDir`; explicitly define symlink behavior.
2. Group duplicate files by digest and print wasted bytes without deleting anything.
3. Inject a hashing function so tests can deterministically block workers and exercise mid-flight cancellation.
4. Add streaming output with a `-stream` flag. Explain the tradeoff between deterministic order and bounded result memory.
5. Benchmark worker counts 1, 2, 4, and 8 on the same input corpus. Record file sizes, hardware, and warm-cache effects.

Portfolio evidence: include a pipeline diagram and a table of measured throughput.
