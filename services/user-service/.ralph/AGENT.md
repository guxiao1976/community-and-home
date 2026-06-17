# user-service Build Guide

See [CLAUDE.md](../CLAUDE.md) for full role and rules.

## Build
```bash
cd services/user-service && go build ./...
```

## Test
```bash
cd services/user-service && go test ./...
```

## Proto (do NOT modify)
Proto definitions are in `api-proto/api/user/v1/`. Changes must go through the global Claude instance.
