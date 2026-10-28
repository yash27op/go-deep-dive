# 03 · PocketKV

A localhost HTTP API backed by a JSON snapshot. Run from the collection root:

```sh
go run ./03-pocketkv -addr 127.0.0.1:8080 -data store.json
go test -race -v ./03-pocketkv
```

| Request | Result |
| --- | --- |
| `PUT /kv/{key}` with `{"value":"hello"}` | 204 after persistence succeeds |
| `GET /kv/{key}` | 200 with `{"value":"hello"}`, or 404 |
| Other methods on a key | 405 with an Allow header |
| Invalid PUT body | 400 |
| Persistence failure | 500; previous in-memory data survives |

PUT bodies allow only a string `value` field, require exactly one JSON object, and have a 64 KiB size limit. Empty strings are allowed. Keys occupy one URL path segment. Unknown routes return 404. Like Go's standard JSON decoder, duplicate object fields use the last value.

## Understand the transaction

A writer holds the mutex, copies the current map, updates the copy, writes and syncs a temporary file in the snapshot directory, closes it, and renames it over the snapshot. Only then does it publish the new map. Readers use an `RWMutex` and never observe half a transaction.

- Why can't two writers copy outside the lock and both publish safely?
- Why must the temporary file be in the same directory as the snapshot?
- Why publish memory only after persistence succeeds?
- What does atomic rename guarantee, and what doesn't it guarantee about power loss?
- How does `httptest` let you test HTTP behavior without opening a port?
- Why does graceful shutdown have its own context and timeout?

## Scope and tradeoffs

Run one process per snapshot file. There is no cross-process locking, authentication, replication, expiration, or storage-size quota. Keep the default localhost binding for local experiments. Each write copies and serializes the entire map and holds the lock through disk I/O; this intentionally favors a simple transaction model over throughput. Atomic replacement targets local Unix-like filesystems. The parent directory is not fsynced, so power-loss durability is not guaranteed. A malformed existing snapshot fails startup instead of silently discarding data.

## Make it yours

1. Add DELETE with the same failure semantics as PUT.
2. Add a storage interface, then run the same HTTP contract tests against memory and disk implementations.
3. Add optimistic concurrency with versions and `If-Match`; test two writers updating the same version.
4. Add a write-ahead log and recovery tests for a truncated last record. Compare write cost with snapshots.
5. Profile concurrent reads and writes. Explain when `RWMutex` helps and when disk latency dominates.

Portfolio evidence: document your consistency guarantees and demonstrate recovery after restart and a failed write.
