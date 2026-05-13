# GPU 加速配置指南

## 当前状态

- ✅ 模型服务代码已完成
- ✅ 模型正在下载中（Qwen2.5-7B-Instruct，约 15GB）
- ⚠️ GPU 未启用（NVIDIA Container Toolkit 未配置）
- ⚠️ 当前使用 CPU 模式（速度较慢）

## 问题诊断

### 1. NVIDIA Container Toolkit 未安装

**错误信息**:
```
could not select device driver "nvidia" with capabilities: [[gpu]]
```

**原因**: Docker 无法访问 NVIDIA GPU，需要安装 NVIDIA Container Toolkit。

## 解决方案

### 方案 1: 安装 NVIDIA Container Toolkit（推荐）

```bash
# 1. 添加 NVIDIA 仓库
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | sudo apt-key add -
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | \
  sudo tee /etc/apt/sources.list.d/nvidia-docker.list

# 2. 安装 nvidia-container-toolkit
sudo apt-get update
sudo apt-get install -y nvidia-container-toolkit

# 3. 重启 Docker
sudo systemctl restart docker

# 4. 验证安装
docker run --rm --gpus all nvidia/cuda:12.1.0-base-ubuntu22.04 nvidia-smi

# 5. 启用 GPU 配置
cd /home/jiaoxh/my-code/community-and-home/services/ai-model
# 编辑 docker-compose.yml，取消注释 GPU 配置部分

# 6. 重启服务
docker-compose down
docker-compose up -d
```

### 方案 2: 继续使用 CPU 模式（临时方案）

**优点**: 无需额外配置，可以立即使用
**缺点**: 
- 推理速度慢（200-500ms vs 50-80ms）
- 内存占用高（~15GB vs ~5GB）
- 吞吐量低（2-5 req/s vs 20 req/s）

**当前配置已是 CPU 模式**，等待模型下载和加载完成即可使用。

## 性能对比

| 指标 | CPU 模式 | GPU 模式 (INT8) |
|------|---------|----------------|
| 响应时间 (P50) | 200-500ms | 50-80ms |
| 显存/内存占用 | ~15GB RAM | ~5GB VRAM |
| 吞吐量 | 2-5 req/s | 20 req/s |
| 首次启动时间 | 10-20 分钟 | 1-2 分钟 |

## 当前进度

```bash
# 查看下载进度
docker stats ai-model-python --no-stream

# 查看日志
docker logs ai-model-python -f

# 检查服务状态
curl http://localhost:8001/health
```

## 预计完成时间

- **模型下载**: 还需 5-10 分钟（剩余约 3-4GB）
- **模型加载**: 下载完成后需 10-15 分钟（CPU 模式）
- **总计**: 约 15-25 分钟

## 建议

1. **短期**: 等待当前 CPU 模式部署完成，验证功能正常
2. **中期**: 安装 NVIDIA Container Toolkit，启用 GPU 加速
3. **长期**: 考虑使用更小的模型（如 Qwen2.5-3B）或量化到 INT4

## 验证步骤

模型加载完成后：

```bash
# 1. 健康检查
curl http://localhost:8001/health

# 2. 测试文本审核
curl -X POST http://localhost:8001/api/moderate/text \
  -H "Content-Type: application/json" \
  -d '{"content": "这是一段测试文本"}'

# 3. 运行完整测试
cd /home/jiaoxh/my-code/community-and-home/services/ai-model
./test.sh
```

## 故障排除

### 容器一直重启
- 检查内存是否充足（至少 16GB）
- 增加健康检查的 start_period（已设置为 300s）

### 下载速度慢
- 已配置 HF_ENDPOINT=https://hf-mirror.com（国内镜像）
- 可以手动下载模型到 volume

### 内存不足
- 关闭其他应用释放内存
- 或使用更小的模型

---

**创建时间**: 2026-05-13  
**状态**: 等待模型下载完成
