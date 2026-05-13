# AI Model Service - 快速开始指南

## 📋 目录

1. [架构概览](#架构概览)
2. [环境要求](#环境要求)
3. [快速启动](#快速启动)
4. [开发调试](#开发调试)
5. [API 使用](#api-使用)
6. [性能优化](#性能优化)
7. [常见问题](#常见问题)

---

## 🏗️ 架构概览

```
┌─────────────────────────────────────────────────────────────┐
│                     AI Model Service                         │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐         ┌──────────────────────────────┐  │
│  │  Go RPC      │  HTTP   │  Python Inference Engine     │  │
│  │  Service     │────────▶│  (FastAPI + Transformers)    │  │
│  │              │         │                              │  │
│  │  - gRPC API  │         │  - Qwen2.5-7B-Instruct      │  │
│  │  - 服务发现   │         │  - INT8 Quantization        │  │
│  │  - 负载均衡   │         │  - GPU Acceleration         │  │
│  └──────────────┘         └──────────────────────────────┘  │
│         ▲                            ▲                       │
│         │                            │                       │
│         │ gRPC                       │ CUDA                  │
│         │                            │                       │
└─────────┼────────────────────────────┼───────────────────────┘
          │                            │
    ┌─────┴─────┐              ┌───────┴────────┐
    │  Client   │              │  NVIDIA GPU    │
    │  Services │              │  (RTX 5060)    │
    └───────────┘              └────────────────┘
```

### 核心特性

- **分层架构**: Go RPC 层 + Python 推理层，职责清晰
- **GPU 加速**: 支持 CUDA，INT8 量化优化显存占用
- **服务发现**: 集成 Etcd，支持动态扩缩容
- **容器化**: Docker Compose 一键部署
- **可扩展**: 预留图片审核等多模型接口

---

## 💻 环境要求

### 硬件要求

- **GPU**: NVIDIA RTX 5060 (8GB) 或更高
- **内存**: 16GB+ RAM
- **存储**: 20GB+ 可用空间（模型缓存）

### 软件要求

```bash
# 必需
- Docker 20.10+
- Docker Compose 2.0+
- NVIDIA Driver 525+
- NVIDIA Container Toolkit

# 开发调试（可选）
- Go 1.21+
- Python 3.10+
- CUDA 12.1+
```

### 验证环境

```bash
# 检查 GPU
nvidia-smi

# 检查 Docker GPU 支持
docker run --rm --gpus all nvidia/cuda:12.1.0-base-ubuntu22.04 nvidia-smi

# 检查 Go
go version

# 检查 Python
python3 --version
```

---

## 🚀 快速启动

### 方式一: Docker Compose（推荐）

```bash
# 1. 进入服务目录
cd services/ai-model

# 2. 构建镜像（首次运行）
make build
# 或
docker-compose build

# 3. 启动服务
make up
# 或
docker-compose up -d

# 4. 查看日志
make logs
# 或
docker-compose logs -f

# 5. 健康检查
curl http://localhost:8001/health  # Python 引擎
curl http://localhost:8084/health  # Go RPC

# 6. 运行测试
make test
# 或
./test.sh
```

### 方式二: 本地开发

```bash
# 1. 启动 Python 引擎
./start.sh
# 或
cd python-engine
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
python -m uvicorn app.main:app --host 0.0.0.0 --port 8001

# 2. 启动 Go RPC（新终端）
cd rpc
go run ai_model.go -f etc/ai-model.yaml
```

### 首次启动注意事项

⚠️ **首次启动会自动下载模型（~15GB），需要 10-20 分钟**

```bash
# 可以提前下载模型
make download-model

# 或手动下载
cd python-engine
python -c "
from transformers import AutoTokenizer, AutoModelForCausalLM
AutoTokenizer.from_pretrained('Qwen/Qwen2.5-7B-Instruct')
AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct', 
                                      device_map='auto', 
                                      load_in_8bit=True)
"
```

---

## 🛠️ 开发调试

### 生成 gRPC 代码

```bash
# 修改 proto 文件后需要重新生成
make proto

# 或手动执行
cd rpc
protoc --go_out=. --go-grpc_out=. pb/ai_model.proto
```

### 查看服务状态

```bash
# 查看容器状态
docker-compose ps

# 查看 GPU 使用情况
nvidia-smi
# 或
make check-gpu

# 查看日志
docker-compose logs -f python-engine  # Python 引擎日志
docker-compose logs -f ai-model-rpc   # Go RPC 日志
```

### 重启服务

```bash
# 重启所有服务
make restart

# 重启单个服务
docker-compose restart python-engine
docker-compose restart ai-model-rpc
```

### 清理环境

```bash
# 停止服务
make down

# 清理所有数据（包括模型缓存）
make clean
```

---

## 📡 API 使用

### 1. Python 引擎 HTTP API

#### 文本审核

```bash
curl -X POST http://localhost:8001/api/moderate/text \
  -H "Content-Type: application/json" \
  -d '{
    "content": "今天天气真好",
    "check_categories": ["政治敏感", "色情低俗", "暴力恐怖"]
  }'
```

**响应示例**:

```json
{
  "is_safe": true,
  "risk_level": "safe",
  "categories": [],
  "reason": "内容健康，未发现违规信息",
  "confidence": 0.98,
  "processing_time_ms": 45
}
```

#### 健康检查

```bash
curl http://localhost:8001/health
```

#### 模型信息

```bash
curl http://localhost:8001/model/info
```

### 2. Go RPC gRPC API

#### Go 客户端示例

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
    pb "your-project/services/ai-model/rpc/pb"
)

func main() {
    // 连接 gRPC 服务
    conn, err := grpc.Dial("localhost:8084", 
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := pb.NewAiModelClient(conn)

    // 文本审核
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    resp, err := client.TextModeration(ctx, &pb.TextModerationRequest{
        Content: "测试内容",
        CheckCategories: []string{"政治敏感", "色情低俗"},
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("审核结果: %+v", resp.Result)
}
```

#### grpcurl 测试

```bash
# 安装 grpcurl
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 列出服务
grpcurl -plaintext localhost:8084 list

# 文本审核
grpcurl -plaintext -d '{
  "content": "测试内容",
  "check_categories": ["政治敏感"]
}' localhost:8084 ai_model.AiModel/TextModeration

# 健康检查
grpcurl -plaintext localhost:8084 ai_model.AiModel/HealthCheck
```

---

## ⚡ 性能优化

### 1. 批量推理

```python
# Python 引擎支持批量请求
contents = ["内容1", "内容2", "内容3"]
results = []
for content in contents:
    result = await moderate_text(content)
    results.append(result)
```

### 2. 模型量化

当前使用 INT8 量化，可进一步优化：

```python
# 修改 python-engine/app/main.py
model = AutoModelForCausalLM.from_pretrained(
    model_name,
    device_map="auto",
    load_in_8bit=True,      # INT8 量化
    # load_in_4bit=True,    # INT4 量化（更省显存，略降精度）
)
```

### 3. 缓存策略

```python
# 添加 Redis 缓存（可选）
from redis import Redis
cache = Redis(host='localhost', port=6379)

def moderate_with_cache(content: str):
    cache_key = f"moderate:{hash(content)}"
    cached = cache.get(cache_key)
    if cached:
        return json.loads(cached)
    
    result = moderate_text(content)
    cache.setex(cache_key, 3600, json.dumps(result))  # 缓存1小时
    return result
```

### 4. GPU 优化

```bash
# 监控 GPU 使用
watch -n 1 nvidia-smi

# 调整 batch size（修改 Python 引擎配置）
# 增大 batch size 可提高吞吐量，但会增加延迟
```

### 性能基准

| 配置 | 延迟 (P50) | 延迟 (P99) | 吞吐量 | 显存占用 |
|------|-----------|-----------|--------|---------|
| RTX 5060 8GB + INT8 | 50ms | 80ms | 20 req/s | 5GB |
| RTX 4090 24GB + FP16 | 30ms | 50ms | 40 req/s | 14GB |
| A100 40GB + FP16 | 15ms | 25ms | 80 req/s | 14GB |

---

## ❓ 常见问题

### Q1: 首次启动很慢？

**A**: 首次启动需要下载模型（~15GB），建议提前下载：

```bash
make download-model
```

### Q2: 显存不足？

**A**: 尝试以下方案：

1. 使用 INT4 量化（修改 `load_in_4bit=True`）
2. 减小 `max_length` 参数
3. 关闭其他 GPU 应用

### Q3: 推理速度慢？

**A**: 检查以下项：

```bash
# 1. 确认使用 GPU
nvidia-smi  # 查看 GPU 利用率

# 2. 检查 CUDA 版本
python -c "import torch; print(torch.cuda.is_available())"

# 3. 启用 Flash Attention（需要 A100/H100）
pip install flash-attn
```

### Q4: gRPC 连接失败？

**A**: 检查服务状态：

```bash
# 查看端口占用
netstat -tlnp | grep 8084

# 查看服务日志
docker-compose logs ai-model-rpc

# 测试连接
grpcurl -plaintext localhost:8084 list
```

### Q5: 如何添加新模型？

**A**: 参考以下步骤：

1. 修改 `pb/ai_model.proto` 添加新接口
2. 重新生成 gRPC 代码：`make proto`
3. 在 Python 引擎添加模型加载和推理逻辑
4. 在 Go RPC 添加对应的 Logic 实现
5. 更新 Docker 镜像：`make build`

---

## 📚 相关文档

- [Qwen2.5 模型文档](https://github.com/QwenLM/Qwen2.5)
- [Transformers 文档](https://huggingface.co/docs/transformers)
- [gRPC Go 教程](https://grpc.io/docs/languages/go/)
- [FastAPI 文档](https://fastapi.tiangolo.com/)

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

---

## 📄 许可证

MIT License
