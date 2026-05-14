# community-and-home Development Guidelines

## Tech Stack

**Backend:**
- Go 1.21+ + go-zero 1.6+ (microservices framework)
- MySQL 8.0 (database) + Redis 7.0 (cache)
- gRPC (service communication) + JWT (authentication)

**Frontend:**
- Vue 3.4+ + TypeScript 5.0+ + Element Plus
- Vite 6.0+ (build) + Pinia (state) + Vue Router + Axios

## Project Structure

```text
community-and-home/
├── services/          # Microservices (identity, masterdata, ai-model, moderation)
├── web/pc/           # Vue3 admin frontend
├── gateway/          # API Gateway
├── common/           # Shared libraries (errorx, jwtx, responsex)
└── docs/             # Documentation
```

See `PROJECT_STRUCTURE.md` for detailed structure and service descriptions.

## Quick Commands

```bash
# Start all services
docker-compose up -d

# Build and run a service
cd services/{service}/api && go build && ./{service}-api

# Frontend development
cd web/pc && npm run dev

# Generate API code from .api file
goctl api go -api {service}.api -dir .

# Generate RPC code from .proto file
goctl rpc protoc {service}.proto --go_out=. --go-grpc_out=. --zrpc_out=.
```

## Code Style

**Go:** Follow standard conventions, use gofmt

**Vue/TypeScript:** Composition API, strict mode, PascalCase components

## Critical Rules

1. **Handler响应格式（强制）**
   - 使用 `responsex.Response(w, resp, err)`，不用 `httpx.OkJsonCtx`
   - 统一返回格式：`{ code: 0, message: "success", data: {...} }`

2. **goctl代码生成**
   - `.api` 文件的 `group` 名称用下划线（`deleted_items`），不用连字符（`deleted-items`）
   - goctl 会覆盖 `types.go`，手动字段需同步到 `.api` 文件

3. **时间字段处理**
   - 创建记录时使用 `time.Now()`，不用零值 `time.Time{}`
   - 避免 MySQL datetime 错误 `'0000-00-00'`

4. **JSON字段存储**
   - MySQL JSON 类型必须用 `json.Marshal()` 转换，不能直接存字符串

5. **缓存失效**
   - 更新/删除操作后必须调用 `InvalidateCache(id)`

## Naming Conventions

- **API层**: 驼峰 (`CostPer1KInputTokens`)
- **RPC层**: 下划线 (`CostPer_1KInputTokens`)
- **数据库**: 蛇形 (`cost_per_1k_input_tokens`)
- **表名**: 前缀+下划线 (`id_user`, `md_administrative_division`)
- **时间字段**: `created_time`, `updated_time`, `delete_time`

## API Documentation

API documentation: `docs/api/` (identity-service.md, masterdata-service.md)

Default test credentials: Phone `13800000000`, Password `Admin@123456`

## Service Ports

- API Gateway: 8080 (public entry)
- Identity API: 8888 | RPC: 8080
- Masterdata API: 8889
- AI-Model API: 8891 | RPC: 8084
- Frontend PC: 3003
- MySQL: 3306 | Redis: 6379

<!-- MANUAL ADDITIONS START -->

## Development Tips

**Time Fields:**
```go
// Wrong: time.Time{} causes MySQL '0000-00-00' error
// Right: use time.Now()
config.CreatedTime = time.Now()
```

**JSON Fields:**
```go
// Wrong: storing plain string "chat,streaming"
// Right: marshal to JSON array
features := strings.Split(input, ",")
jsonBytes, _ := json.Marshal(features)  // ["chat","streaming"]
```

**Cache Invalidation:**
```go
func (m *Manager) UpdateModel(ctx context.Context, model *Model) error {
    err := m.modelModel.Update(ctx, model)
    m.InvalidateCache(model.Id)  // Must clear cache
    return err
}
```

<!-- MANUAL ADDITIONS END -->
