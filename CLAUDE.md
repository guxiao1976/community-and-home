# community-and-home Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-05-03

## Active Technologies
- Vue 3.4+ with TypeScript 5.0+ + Element Plus (UI), Axios (HTTP), Pinia (state), Vue Router (routing), Vite (build) (002-web-pc-admin)
- Browser localStorage for JWT tokens (access 24h, refresh 7d), sessionStorage for temporary state (002-web-pc-admin)

- Go 1.21+ + go-zero 1.6+, gRPC, JWT (golang-jwt/jwt), bcrypt (golang.org/x/crypto/bcrypt) (001-identity-masterdata)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.21+

## Code Style

Go 1.21+: Follow standard conventions

### Handler 响应格式约定

所有 handler 必须使用 `responsex.Response(w, resp, err)` 返回响应，**不要**使用 `httpx.OkJsonCtx`。`responsex.Response` 会统一包装为 `{ code: 0, message: "success", data: {...} }` 格式，前端 Axios 拦截器依赖此格式解构 `code` 和 `data`。

### goctl 代码生成注意

- `.api` 文件中的 `group` 名称不能包含连字符（如 `deleted-items`），否则生成的目录和 import 别名是非法 Go 标识符，应使用下划线（如 `deleted_items`）
- goctl 会覆盖 `types.go`，手动添加的字段需同步更新到 `.api` 文件中

## Recent Changes
- 002-web-pc-admin: Added Vue 3.4+ with TypeScript 5.0+ + Element Plus (UI), Axios (HTTP), Pinia (state), Vue Router (routing), Vite (build)

- 001-identity-masterdata: Added Go 1.21+ + go-zero 1.6+, gRPC, JWT (golang-jwt/jwt), bcrypt (golang.org/x/crypto/bcrypt)

## API Documentation

Frontend API documentation is available in `docs/api/`:

- **docs/api/README.md** - API overview, architecture, authentication guide
- **docs/api/quick-start.md** - Quick start guide with curl examples
- **docs/api/constants.md** - All enums, constants, validation rules
- **docs/api/identity-service.md** - Identity Service API reference (28 endpoints)
- **docs/api/masterdata-service.md** - Masterdata Service API reference (18 endpoints)

Default credentials: Phone `13800000000`, Password `Admin@123456`

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
