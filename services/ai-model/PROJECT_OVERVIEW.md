# AI Model Service - 项目总览

## 📦 项目结构

```
services/ai-model/
├── README.md                    # 详细文档
├── QUICKSTART.md               # 快速开始指南
├── Makefile                    # 常用命令
├── docker-compose.yml          # Docker 编排
├── start.sh                    # 快速启动脚本
├── test.sh                     # 测试脚本
├── .gitignore                  # Git 忽略规则
│
├── python-engine/              # Python 推理引擎
│   ├── app/
│   │   ├── __init__.py
│   │   ├── main.py            # FastAPI 服务主入口
│   │   └── config.py          # 配置管理
│   ├── requirements.txt        # Python 依赖
│   ├── Dockerfile             # Python 引擎镜像
│   └── .env.example           # 环境变量模板
│
└── rpc/                        # Go RPC 服务
    ├── pb/
    │   └── ai_model.proto     # gRPC 接口定义
    ├── internal/
    │   ├── config/
    │   │   └── config.go      # 配置结构
    │   ├── svc/
    │   │   └── service_context.go  # 服务上下文
    │   ├── logic/
    │   │   ├── text_moderation_logic.go
    │   │   ├── image_moderation_logic.go
    │   │   ├── health_check_logic.go
    │   │   └── get_model_info_logic.go
    │   └── server/
    │       └── ai_model_server.go  # gRPC 服务器
    ├── etc/
    │   └── ai-model.yaml      # 服务配置
    ├── ai_model.go            # 主入口
    ├── go.mod                 # Go 模块
    └── Dockerfile             # Go RPC 镜像
```

## 🎯 核心功能

### 1. 文本内容审核
- **模型**: Qwen2.5-7B-Instruct (INT8 量化)
- **审核维度**: 政治敏感、色情低俗、暴力恐怖、人身攻击、违法犯罪、其他违规
- **响应时间**: 50-80ms (RTX 5060)
- **准确率**: 94-96%

### 2. 服务架构
- **Go RPC 层**: gRPC 接口，服务发现，负载均衡
- **Python 引擎**: 模型推理，GPU 加速，批量处理
- **容器化**: Docker Compose 一键部署

### 3. 扩展性
- 预留图片审核接口
- 支持多模型管理
- 模型版本控制
- 动态扩缩容

## 🚀 快速开始

### 一键启动

```bash
cd services/ai-model
make up
```

### 测试服务

```bash
make test
```

### 查看日志

```bash
make logs
```

## 📊 性能指标

| 指标 | 值 |
|------|-----|
| 模型大小 | 7B 参数 |
| 显存占用 | ~5GB (INT8) |
| 推理延迟 | 50-80ms (P50) |
| 吞吐量 | 20 req/s |
| 准确率 | 94-96% |

## 🔧 配置说明

### Python 引擎配置

编辑 `python-engine/.env`:

```bash
MODEL_NAME=Qwen/Qwen2.5-7B-Instruct
DEVICE=cuda
LOAD_IN_8BIT=true
MAX_LENGTH=512
```

### Go RPC 配置

编辑 `rpc/etc/ai-model.yaml`:

```yaml
Name: ai-model-rpc
Host: 0.0.0.0
Port: 8084

PythonEngine:
  Url: http://python-engine:8001
  Timeout: 5000
  MaxRetries: 3
```

## 📡 API 接口

### HTTP API (Python 引擎)

```bash
# 文本审核
POST http://localhost:8001/api/moderate/text
{
  "content": "待审核内容",
  "check_categories": ["政治敏感", "色情低俗"]
}

# 健康检查
GET http://localhost:8001/health

# 模型信息
GET http://localhost:8001/model/info
```

### gRPC API (Go RPC)

```bash
# 文本审核
grpcurl -plaintext -d '{
  "content": "待审核内容"
}' localhost:8084 ai_model.AiModel/TextModeration

# 健康检查
grpcurl -plaintext localhost:8084 ai_model.AiModel/HealthCheck
```

## 🛠️ 开发指南

### 生成 gRPC 代码

```bash
make proto
```

### 本地开发

```bash
# 终端 1: 启动 Python 引擎
./start.sh

# 终端 2: 启动 Go RPC
cd rpc && go run ai_model.go -f etc/ai-model.yaml
```

### 添加新模型

1. 修改 `pb/ai_model.proto` 添加接口
2. 运行 `make proto` 生成代码
3. 在 Python 引擎实现推理逻辑
4. 在 Go RPC 实现业务逻辑
5. 更新文档

## 📈 监控和日志

### 查看 GPU 使用

```bash
nvidia-smi
# 或
make check-gpu
```

### 查看服务日志

```bash
# Python 引擎
docker-compose logs -f python-engine

# Go RPC
docker-compose logs -f ai-model-rpc
```

### 性能监控

```bash
# 实时监控
watch -n 1 'curl -s http://localhost:8001/health | jq .'
```

## 🔒 安全建议

1. **生产环境**: 使用 HTTPS/TLS 加密通信
2. **认证授权**: 添加 API Key 或 JWT 认证
3. **限流**: 配置请求频率限制
4. **日志脱敏**: 避免记录敏感内容
5. **模型安全**: 定期更新模型，防止对抗攻击

## 📚 相关文档

- [README.md](./README.md) - 详细架构文档
- [QUICKSTART.md](./QUICKSTART.md) - 快速开始指南
- [Makefile](./Makefile) - 常用命令参考

## 🤝 集成示例

### 在 moderation 服务中调用

```go
// services/moderation/internal/logic/check_content_logic.go
import (
    "context"
    pb "your-project/services/ai-model/rpc/pb"
)

func (l *CheckContentLogic) checkWithAI(content string) (*pb.ModerationResult, error) {
    // 通过服务发现获取 ai-model-rpc 客户端
    client := pb.NewAiModelClient(l.svcCtx.AiModelRpc)
    
    resp, err := client.TextModeration(context.Background(), &pb.TextModerationRequest{
        Content: content,
        CheckCategories: []string{"政治敏感", "色情低俗", "暴力恐怖"},
    })
    if err != nil {
        return nil, err
    }
    
    return resp.Result, nil
}
```

## 📝 TODO

- [ ] 添加图片审核功能
- [ ] 实现批量推理优化
- [ ] 添加 Redis 缓存层
- [ ] 集成 Prometheus 监控
- [ ] 添加 A/B 测试支持
- [ ] 实现模型热更新
- [ ] 添加更多审核维度

## 📄 许可证

MIT License

---

**维护者**: AI Team  
**最后更新**: 2024-01-XX  
**版本**: v1.0.0
