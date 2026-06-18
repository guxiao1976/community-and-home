# file-service Build Guide

See [CLAUDE.md](../CLAUDE.md) for full role and rules.

## Build
```bash
cd services/file-service && go build ./...
```

## Test
```bash
cd services/file-service && go test ./...
```

## Proto (do NOT modify)
Proto definitions are in `api-proto/api/file/v1/`. Changes must go through the global Claude instance.
