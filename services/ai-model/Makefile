.PHONY: help proto build up down logs test clean

help:
	@echo "AI Model Service - Makefile Commands"
	@echo ""
	@echo "Development:"
	@echo "  make proto       - Generate gRPC code from proto files"
	@echo "  make build       - Build Docker images"
	@echo "  make up          - Start all services"
	@echo "  make down        - Stop all services"
	@echo "  make logs        - View service logs"
	@echo "  make test        - Run test suite"
	@echo ""
	@echo "Quick Start:"
	@echo "  make start-dev   - Start Python engine locally"
	@echo "  make start-rpc   - Start Go RPC locally"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean       - Clean up containers and volumes"
	@echo "  make restart     - Restart all services"

# 生成 gRPC 代码
proto:
	@echo "Generating gRPC code..."
	cd rpc && protoc --go_out=. --go-grpc_out=. pb/ai_model.proto
	@echo "✓ Done"

# 构建 Docker 镜像
build:
	@echo "Building Docker images..."
	docker-compose build
	@echo "✓ Done"

# 启动所有服务
up:
	@echo "Starting services..."
	docker-compose up -d
	@echo "✓ Services started"
	@echo ""
	@echo "Check status: make logs"

# 停止所有服务
down:
	@echo "Stopping services..."
	docker-compose down
	@echo "✓ Services stopped"

# 查看日志
logs:
	docker-compose logs -f

# 运行测试
test:
	@echo "Running tests..."
	./test.sh

# 清理
clean:
	@echo "Cleaning up..."
	docker-compose down -v
	rm -rf python-engine/venv
	rm -rf python-engine/__pycache__
	rm -rf python-engine/app/__pycache__
	@echo "✓ Cleaned"

# 重启服务
restart: down up

# 本地开发 - 启动 Python 引擎
start-dev:
	@echo "Starting Python engine locally..."
	./start.sh

# 本地开发 - 启动 Go RPC
start-rpc:
	@echo "Starting Go RPC locally..."
	cd rpc && go run ai_model.go -f etc/ai-model.yaml

# 检查 GPU
check-gpu:
	@echo "GPU Status:"
	@nvidia-smi --query-gpu=index,name,memory.used,memory.total,utilization.gpu --format=csv

# 下载模型
download-model:
	@echo "Downloading Qwen2.5-7B-Instruct model..."
	@echo "This will take 10-20 minutes depending on your network speed..."
	cd python-engine && python -c "from transformers import AutoTokenizer, AutoModelForCausalLM; AutoTokenizer.from_pretrained('Qwen/Qwen2.5-7B-Instruct'); AutoModelForCausalLM.from_pretrained('Qwen/Qwen2.5-7B-Instruct', device_map='auto', load_in_8bit=True)"
	@echo "✓ Model downloaded"

# 健康检查
health:
	@echo "Checking service health..."
	@curl -s http://localhost:8001/health | jq .
	@echo ""
	@curl -s http://localhost:8084/health | jq .
