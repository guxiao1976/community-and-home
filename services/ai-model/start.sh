#!/bin/bash

set -e

echo "=========================================="
echo "  AI Model Service - Quick Start"
echo "=========================================="
echo ""

# 检查 NVIDIA GPU
if ! command -v nvidia-smi &> /dev/null; then
    echo "❌ NVIDIA GPU not found. Please install NVIDIA drivers."
    exit 1
fi

echo "✓ GPU detected:"
nvidia-smi --query-gpu=name,memory.total --format=csv,noheader

# 检查 Python
if ! command -v python3 &> /dev/null; then
    echo "❌ Python3 not found. Please install Python 3.10+"
    exit 1
fi

echo "✓ Python version: $(python3 --version)"

# 进入 Python 引擎目录
cd "$(dirname "$0")/python-engine"

# 创建虚拟环境
if [ ! -d "venv" ]; then
    echo ""
    echo "Creating virtual environment..."
    python3 -m venv venv
fi

# 激活虚拟环境
source venv/bin/activate

# 安装依赖
echo ""
echo "Installing dependencies..."
pip install -q --upgrade pip
pip install -q -r requirements.txt

# 启动服务
echo ""
echo "=========================================="
echo "  Starting Python Engine..."
echo "=========================================="
echo ""
echo "📝 Note: First run will download model (~15GB)"
echo "⏱️  Model loading takes 30-60 seconds"
echo ""
echo "🌐 Service will be available at:"
echo "   - Health: http://localhost:8001/health"
echo "   - API: http://localhost:8001/api/moderate/text"
echo ""

python -m uvicorn app.main:app --host 0.0.0.0 --port 8001
