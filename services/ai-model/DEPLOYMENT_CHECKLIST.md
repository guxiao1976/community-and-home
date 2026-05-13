# AI Model Service - 部署检查清单

## 📋 部署前检查

### 1. 硬件环境

- [ ] NVIDIA GPU 可用（RTX 5060 8GB 或更高）
- [ ] 显存至少 8GB
- [ ] 内存至少 16GB
- [ ] 磁盘空间至少 20GB

**验证命令**:
```bash
nvidia-smi
free -h
df -h
```

### 2. 软件依赖

- [ ] Docker 20.10+ 已安装
- [ ] Docker Compose 2.0+ 已安装
- [ ] NVIDIA Driver 525+ 已安装
- [ ] NVIDIA Container Toolkit 已安装

**验证命令**:
```bash
docker --version
docker-compose --version
nvidia-smi
docker run --rm --gpus all nvidia/cuda:12.1.0-base-ubuntu22.04 nvidia-smi
```

### 3. 网络环境

- [ ] 可访问 Hugging Face (下载模型)
- [ ] 端口 8001 未被占用 (Python 引擎)
- [ ] 端口 8084 未被占用 (Go RPC)
- [ ] 端口 2379 未被占用 (Etcd)

**验证命令**:
```bash
curl -I https://huggingface.co
netstat -tlnp | grep -E '8001|8084|2379'
```

---

## 🚀 部署步骤

### Step 1: 克隆代码

```bash
cd /path/to/project
git pull origin main
cd services/ai-model
```

- [ ] 代码已更新到最新版本
- [ ] 所有文件完整无缺失

### Step 2: 配置环境

```bash
# 复制环境变量模板
cp python-engine/.env.example python-engine/.env

# 根据实际情况修改配置
vim python-engine/.env
vim rpc/etc/ai-model.yaml
```

- [ ] Python 引擎配置已修改
- [ ] Go RPC 配置已修改
- [ ] 数据库连接配置正确（如需要）

### Step 3: 下载模型（可选，首次启动会自动下载）

```bash
make download-model
```

- [ ] 模型下载完成（~15GB）
- [ ] 模型缓存目录可访问

### Step 4: 构建镜像

```bash
make build
# 或
docker-compose build
```

- [ ] Python 引擎镜像构建成功
- [ ] Go RPC 镜像构建成功
- [ ] 无构建错误

### Step 5: 启动服务

```bash
make up
# 或
docker-compose up -d
```

- [ ] 所有容器启动成功
- [ ] 无启动错误日志

### Step 6: 健康检查

```bash
# 等待 30-60 秒让模型加载完成
sleep 60

# 检查 Python 引擎
curl http://localhost:8001/health

# 检查 Go RPC
curl http://localhost:8084/health
```

- [ ] Python 引擎健康检查通过
- [ ] Go RPC 健康检查通过
- [ ] 模型加载成功

### Step 7: 功能测试

```bash
make test
# 或
./test.sh
```

- [ ] 安全内容测试通过
- [ ] 敏感内容检测正常
- [ ] 响应时间符合预期（<100ms）
- [ ] 所有测试用例通过

---

## 🔍 部署后验证

### 1. 服务状态检查

```bash
# 查看容器状态
docker-compose ps

# 预期输出：所有服务 State 为 Up
```

- [ ] python-engine: Up
- [ ] ai-model-rpc: Up
- [ ] etcd: Up

### 2. GPU 使用检查

```bash
nvidia-smi

# 预期：python-engine 进程占用 GPU，显存约 5GB
```

- [ ] GPU 被正常使用
- [ ] 显存占用合理（~5GB）
- [ ] GPU 利用率正常

### 3. 日志检查

```bash
# Python 引擎日志
docker-compose logs python-engine | tail -50

# Go RPC 日志
docker-compose logs ai-model-rpc | tail -50
```

- [ ] 无 ERROR 级别日志
- [ ] 模型加载成功日志存在
- [ ] 服务启动成功日志存在

### 4. 性能测试

```bash
# 运行性能测试
for i in {1..10}; do
  time curl -s -X POST http://localhost:8001/api/moderate/text \
    -H "Content-Type: application/json" \
    -d '{"content": "测试内容'$i'"}'
done
```

- [ ] 平均响应时间 < 100ms
- [ ] 无超时错误
- [ ] 结果准确

### 5. 集成测试

```bash
# 从其他服务调用 gRPC 接口
grpcurl -plaintext -d '{
  "content": "测试内容"
}' localhost:8084 ai_model.AiModel/TextModeration
```

- [ ] gRPC 调用成功
- [ ] 返回结果正确
- [ ] 服务发现正常（如使用 Etcd）

---

## 🐛 常见问题排查

### 问题 1: 容器启动失败

**症状**: `docker-compose up` 报错

**排查步骤**:
```bash
# 查看详细日志
docker-compose logs

# 检查端口占用
netstat -tlnp | grep -E '8001|8084|2379'

# 检查磁盘空间
df -h
```

**解决方案**:
- [ ] 释放被占用的端口
- [ ] 清理磁盘空间
- [ ] 检查 Docker 配置

### 问题 2: 模型加载失败

**症状**: Python 引擎日志显示模型加载错误

**排查步骤**:
```bash
# 检查网络连接
curl -I https://huggingface.co

# 检查磁盘空间
df -h

# 手动下载模型
make download-model
```

**解决方案**:
- [ ] 配置代理访问 Hugging Face
- [ ] 清理磁盘空间
- [ ] 手动下载模型到缓存目录

### 问题 3: GPU 不可用

**症状**: 日志显示 "CUDA not available"

**排查步骤**:
```bash
# 检查 NVIDIA 驱动
nvidia-smi

# 检查 Docker GPU 支持
docker run --rm --gpus all nvidia/cuda:12.1.0-base-ubuntu22.04 nvidia-smi

# 检查容器 GPU 配置
docker inspect ai-model-python | grep -i gpu
```

**解决方案**:
- [ ] 安装/更新 NVIDIA 驱动
- [ ] 安装 NVIDIA Container Toolkit
- [ ] 检查 docker-compose.yml 中的 GPU 配置

### 问题 4: 推理速度慢

**症状**: 响应时间 > 200ms

**排查步骤**:
```bash
# 检查 GPU 利用率
nvidia-smi

# 检查系统负载
top

# 检查模型量化配置
docker-compose logs python-engine | grep "load_in_8bit"
```

**解决方案**:
- [ ] 确认使用 INT8 量化
- [ ] 减小 max_length 参数
- [ ] 关闭其他 GPU 应用
- [ ] 启用批量推理

### 问题 5: 内存不足

**症状**: OOM (Out of Memory) 错误

**排查步骤**:
```bash
# 检查显存使用
nvidia-smi

# 检查内存使用
free -h
```

**解决方案**:
- [ ] 使用 INT4 量化（修改配置）
- [ ] 减小 batch_size
- [ ] 减小 max_length
- [ ] 升级 GPU

---

## 📊 监控指标

### 关键指标

| 指标 | 正常范围 | 告警阈值 |
|------|---------|---------|
| 响应时间 (P50) | < 80ms | > 150ms |
| 响应时间 (P99) | < 150ms | > 300ms |
| GPU 利用率 | 30-70% | > 90% |
| 显存占用 | ~5GB | > 7GB |
| 错误率 | < 0.1% | > 1% |
| QPS | 10-20 | < 5 |

### 监控命令

```bash
# 实时监控 GPU
watch -n 1 nvidia-smi

# 实时监控服务健康
watch -n 5 'curl -s http://localhost:8001/health | jq .'

# 查看请求统计
docker-compose logs python-engine | grep "total_requests"
```

---

## 🔄 回滚方案

### 快速回滚

```bash
# 停止当前服务
make down

# 切换到上一个版本
git checkout <previous-commit>

# 重新构建和启动
make build
make up
```

### 数据备份

```bash
# 备份模型缓存（如有自定义模型）
tar -czf models-backup.tar.gz python-engine/models/

# 备份配置文件
cp rpc/etc/ai-model.yaml rpc/etc/ai-model.yaml.bak
cp python-engine/.env python-engine/.env.bak
```

---

## ✅ 部署完成确认

- [ ] 所有容器正常运行
- [ ] 健康检查全部通过
- [ ] 功能测试全部通过
- [ ] 性能指标符合预期
- [ ] 监控和日志正常
- [ ] 文档已更新
- [ ] 团队已通知

---

## 📞 支持联系

- **技术支持**: tech-support@example.com
- **紧急联系**: oncall@example.com
- **文档**: [PROJECT_OVERVIEW.md](./PROJECT_OVERVIEW.md)

---

**检查人**: ___________  
**检查日期**: ___________  
**部署环境**: [ ] 开发 [ ] 测试 [ ] 生产  
**版本号**: ___________
