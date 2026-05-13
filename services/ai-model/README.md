# AI Model Service 部署文档

## 服务架构

```
┌─────────────────────────────────────────────────────────┐
│                   AI Model Service                       │
├─────────────────────────────────────────────────────────┤
│                                                          │
│  ┌──────────────────┐         ┌──────────────────┐     │
│  │  Go gRPC Service │ ◄─────► │ Python Engine    │     │
│  │  (Port: 8084)    │  HTTP   │ (Port: 8001)     │     │
│  │                  │         │                  │     │
│  │  - 服务注册      │         │  - Qwen2.5-7B    │     │
│  │  - 负载均衡      │         │  - INT8量化      │     │
│  │  - 重试机制      │         │  - GPU推理       │     │
│  └──────────────────┘         └──────────────────┘     │
│                                                          │
└─────────────────────────────────────────────────────────┘
```

## 一、本地开发部署（推荐用于开发调试）

### 1.1 前置条件

- NVIDIA GPU (RTX 5060 8GB 或更高)
- CUDA 12.1+
- Python 3.10+
- Go 1.21+
- Docker & Docker Compose (可选)

### 1.2 快速启动 Python 引擎

```bash
cd services/ai-model/python-engine

# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate  # Windows: venv\Scripts\activate

# 安装依赖
pip install -r requirements.txt

# 启动服务（首次启动会自动下载模型，约15GB）
python -m uvicorn app.main:app --host 0.0.0.0 --port 8001
```

**首次启动时间：** 约 5-10 分钟（下载模型）  
**后续启动时间：** 约 30-60 秒（加载模型到GPU）

### 1.3 测试 Python 引擎

```bash
# 健康检查
curl http://localhost:8001/health

# 文本审核测试
curl -X POST http://localhost:8001/api/moderate/text \
  -H "Content-Type: application/json" \
  -d '{
    "content": "今天天气真好",
    "check_categories": ["政治敏感", "色情低俗"]
  }'

# 违规内容测试
curl -X POST http://localhost:8001/api/moderate/text \
  -H "Content-Type: application/json" \
  -d '{
    "content": "习近平是傻逼"
  }'
```

### 1.4 启动 Go gRPC 服务

```bash
cd services/ai-model/rpc

# 生成 protobuf 代码
protoc --go_out=. --go-grpc_out=. pb/ai_model.proto

# 启动服务
go run ai_model.go -f etc/ai-model.yaml
```

### 1.5 测试 gRPC 服务

```bash
# 使用 grpcurl 测试
grpcurl -plaintext \
  -d '{
    "content": "打倒共产党",
    "check_categories": ["政治敏感"]
  }' \
  localhost:8084 ai_model.AiModel/ModerateText

# 健康检查
grpcurl -plaintext localhost:8084 ai_model.AiModel/HealthCheck
```

---

## 二、Docker 部署（推荐用于生产环境）

### 2.1 构建镜像

```bash
cd services/ai-model

# 构建 Python 引擎镜像
docker build -t ai-model-python:latest ./python-engine

# 构建 Go RPC 镜像
docker build -t ai-model-rpc:latest ./rpc
```

### 2.2 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f ai-model-python
docker-compose logs -f ai-model-rpc

# 查看状态
docker-compose ps
```

### 2.3 停止服务

```bash
docker-compose down

# 清理模型缓存（慎用）
docker-compose down -v
```

---

## 三、性能优化

### 3.1 显存优化

**当前配置（INT8）：**
- 显存占用：~5GB
- 推理速度：50-80ms/条
- 适合：RTX 5060 8GB

**如果显存不足，可降级到 INT4：**

```python
# python-engine/app/main.py
model = AutoModelForCausalLM.from_pretrained(
    model_name,
    torch_dtype=torch.float16,
    device_map="auto",
    load_in_4bit=True  # 改为 INT4
)
```

- 显存占用：~3GB
- 推理速度：80-120ms/条
- 准确率：略有下降（~2%）

### 3.2 批处理优化

```python
# 修改 main.py，支持批量推理
@app.post("/api/moderate/text/batch")
async def moderate_text_batch(requests: List[TextModerationRequest]):
    # 批量处理，提升吞吐量
    pass
```

### 3.3 缓存优化

在 Go RPC 层添加 Redis 缓存：

```go
// 相同内容 1 小时内直接返回缓存结果
cacheKey := fmt.Sprintf("moderation:%s", hash(content))
if cached, err := redis.Get(cacheKey); err == nil {
    return cached
}
```

---

## 四、监控与告警

### 4.1 健康检查端点

```bash
# Python 引擎
curl http://localhost:8001/health

# Go RPC 服务
grpcurl -plaintext localhost:8084 ai_model.AiModel/HealthCheck
```

### 4.2 关键指标

| 指标 | 正常值 | 告警阈值 |
|------|--------|----------|
| GPU 显存占用 | 5-6GB | >7GB |
| 平均推理延迟 | 50-80ms | >200ms |
| 模型加载状态 | loaded=true | loaded=false |
| 错误率 | <1% | >5% |

### 4.3 日志查看

```bash
# Python 引擎日志
docker logs -f ai-model-python

# Go RPC 日志
docker logs -f ai-model-rpc

# 过滤错误日志
docker logs ai-model-python 2>&1 | grep ERROR
```

---

## 五、故障排查

### 5.1 模型加载失败

**症状：** 启动后一直显示 "Loading model..."

**原因：**
1. 网络问题，无法下载模型
2. 显存不足
3. CUDA 版本不匹配

**解决：**
```bash
# 手动下载模型
cd python-engine
python -c "
from transformers import AutoModelForCausalLM, AutoTokenizer
AutoTokenizer.from_pretrained('Qwen/Qwen2.5-7B-Instruct')
AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct')
"

# 检查 CUDA
nvidia-smi
python -c "import torch; print(torch.cuda.is_available())"
```

### 5.2 推理速度慢

**症状：** 延迟 >500ms

**原因：**
1. 使用 CPU 推理（未检测到 GPU）
2. 模型未量化
3. 批处理大小不合理

**解决：**
```bash
# 确认 GPU 可用
docker exec -it ai-model-python nvidia-smi

# 检查模型配置
docker exec -it ai-model-python python -c "
import torch
from transformers import AutoModelForCausalLM
model = AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct', load_in_8bit=True)
print(f'Device: {model.device}')
print(f'Dtype: {model.dtype}')
"
```

### 5.3 Go RPC 连接失败

**症状：** `python engine unavailable`

**原因：**
1. Python 引擎未启动
2. 端口配置错误
3. 网络不通

**解决：**
```bash
# 检查 Python 引擎状态
curl http://localhost:8001/health

# 检查端口
netstat -tuln | grep 8001

# 检查 Docker 网络
docker network inspect ai-model-network
```

---

## 六、扩展指南

### 6.1 添加新模型

1. 在 `python-engine/app/models/` 下创建新模型类
2. 在 `main.py` 中注册新端点
3. 更新 `ai_model.proto` 添加新接口
4. 在 Go RPC 中实现对应 Logic

### 6.2 添加图片审核

```python
# python-engine/app/models/image_model.py
from transformers import AutoModel, AutoProcessor

class ImageModerationModel:
    def __init__(self):
        self.model = AutoModel.from_pretrained("...")
        self.processor = AutoProcessor.from_pretrained("...")
    
    def moderate(self, image_data: bytes) -> dict:
        # 实现图片审核逻辑
        pass
```

### 6.3 多模型负载均衡

```yaml
# docker-compose.yml
services:
  ai-model-python-1:
    ...
  ai-model-python-2:
    ...
  
  ai-model-nginx:
    image: nginx:alpine
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "8001:8001"
```

---

## 七、生产环境建议

### 7.1 硬件配置

| 场景 | GPU | 显存 | 并发 | 成本/月 |
|------|-----|------|------|---------|
| 开发测试 | RTX 5060 | 8GB | 5 QPS | ¥0 |
| 小规模生产 | T4 | 16GB | 20 QPS | ¥2,000 |
| 中规模生产 | A10 | 24GB | 50 QPS | ¥5,000 |
| 大规模生产 | A100 | 40GB | 100 QPS | ¥15,000 |

### 7.2 高可用部署

```
┌─────────────┐
│   Nginx LB  │
└──────┬──────┘
       │
   ┌───┴───┬───────┬───────┐
   │       │       │       │
┌──▼──┐ ┌──▼──┐ ┌──▼──┐ ┌──▼──┐
│ GPU1│ │ GPU2│ │ GPU3│ │ GPU4│
└─────┘ └─────┘ └─────┘ └─────┘
```

### 7.3 成本优化

1. **使用 Spot 实例**：成本降低 70%
2. **按需启动**：非高峰期关闭部分实例
3. **模型缓存**：避免重复推理
4. **分层审核**：AC自动机 → 小模型 → 大模型

---

## 八、常用命令

```bash
# 启动服务
docker-compose up -d

# 重启服务
docker-compose restart

# 查看日志
docker-compose logs -f

# 进入容器
docker exec -it ai-model-python bash

# 查看 GPU 使用
nvidia-smi

# 性能测试
ab -n 100 -c 10 -p test.json -T application/json http://localhost:8001/api/moderate/text

# 清理缓存
docker system prune -a
```

---

## 九、联系方式

- 技术支持：[GitHub Issues](https://github.com/your-repo/issues)
- 文档更新：2026-05-12
