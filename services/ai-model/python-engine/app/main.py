from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import List, Optional, Dict
import torch
from transformers import AutoModelForCausalLM, AutoTokenizer
import json
import re
import time
import logging
from contextlib import asynccontextmanager

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# ============ 全局变量 ============
model = None
tokenizer = None
model_info = {
    "name": "Qwen2.5-3B-Instruct",
    "version": "v1.0",
    "loaded": False,
    "start_time": time.time()
}
stats = {
    "total_requests": 0,
    "total_latency": 0.0
}

# ============ 数据模型 ============
class TextModerationRequest(BaseModel):
    content: str = Field(..., description="待审核文本")
    check_categories: Optional[List[str]] = Field(default=None, description="检查维度")
    context: Optional[str] = Field(default=None, description="上下文")
    metadata: Optional[Dict[str, str]] = Field(default=None)

class TextModerationResponse(BaseModel):
    is_safe: bool
    risk_level: str  # high/medium/low/safe
    categories: List[str]
    reason: str
    confidence: float
    latency_ms: int
    model_version: str

class HealthCheckResponse(BaseModel):
    status: str
    model_loaded: bool
    model_name: str
    model_version: str
    uptime_seconds: int
    total_requests: int
    avg_latency_ms: float
    gpu_available: bool
    gpu_memory_allocated_mb: Optional[float]

class ModelInfoResponse(BaseModel):
    name: str
    version: str
    type: str
    parameters: int
    capabilities: List[str]
    quantization: str

# ============ 模型加载 ============
def load_model():
    global model, tokenizer, model_info

    logger.info("Loading Qwen2.5-3B-Instruct model...")
    model_name = "Qwen/Qwen2.5-3B-Instruct"

    # 使用国内镜像源
    import os
    os.environ['HF_ENDPOINT'] = 'https://hf-mirror.com'

    try:
        tokenizer = AutoTokenizer.from_pretrained(
            model_name,
            mirror='https://hf-mirror.com'
        )

        # 检测是否有 CUDA 可用
        device = "cuda" if torch.cuda.is_available() else "cpu"

        if device == "cuda":
            logger.info("CUDA available, loading with float16 (half precision)")
            model = AutoModelForCausalLM.from_pretrained(
                model_name,
                torch_dtype=torch.float16,
                device_map="auto",
                mirror='https://hf-mirror.com'
            )
        else:
            logger.info("CUDA not available, loading in CPU mode (float32)")
            model = AutoModelForCausalLM.from_pretrained(
                model_name,
                torch_dtype=torch.float32,
                low_cpu_mem_usage=True,
                mirror='https://hf-mirror.com'
            )
            model = model.to(device)
        model_info["loaded"] = True
        logger.info("Model loaded successfully!")
        logger.info(f"Device: {model.device}")
        logger.info(f"Memory allocated: {torch.cuda.memory_allocated() / 1024**2:.2f} MB")
    except Exception as e:
        logger.error(f"Failed to load model: {e}")
        raise

# ============ 生命周期管理 ============
@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时加载模型
    load_model()
    yield
    # 关闭时清理
    logger.info("Shutting down...")

# ============ FastAPI 应用 ============
app = FastAPI(
    title="AI Model Service - Python Engine",
    description="Text moderation using Qwen2.5-7B-Instruct",
    version="1.0.0",
    lifespan=lifespan
)

# ============ 审核逻辑 ============
def moderate_text_internal(content: str, check_categories: Optional[List[str]] = None) -> dict:
    """
    内部审核逻辑
    """
    start_time = time.time()

    # 构建检查维度
    if check_categories:
        categories_str = "、".join(check_categories)
    else:
        categories_str = "政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销"

    # 构建Prompt - 使用few-shot learning
    user_prompt = f"""请判断以下文本是否违规。

检查维度：{categories_str}

文本内容：{content}

请严格按照JSON格式输出。"""

    messages = [
        {"role": "system", "content": "你是一个专业的内容审核助手，能够准确理解文本的语义和立场。"},
        
        # Few-shot示例1：正面立场
        {"role": "user", "content": '请判断以下文本是否违规。\n\n检查维度：政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销\n\n文本内容：禁止枪支、毒品\n\n请严格按照JSON格式输出。'},
        {"role": "assistant", "content": '{"is_safe": true, "risk_level": "safe", "categories": [], "reason": "文本表达反对违法行为的立场，属于正面宣传", "confidence": 0.95}'},
        
        # Few-shot示例2：负面立场
        {"role": "user", "content": '请判断以下文本是否违规。\n\n检查维度：政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销\n\n文本内容：如何购买枪支和毒品\n\n请严格按照JSON格式输出。'},
        {"role": "assistant", "content": '{"is_safe": false, "risk_level": "high", "categories": ["违法犯罪"], "reason": "文本涉及教唆购买违禁品", "confidence": 0.95}'},
        
        # Few-shot示例3：正常投诉
        {"role": "user", "content": '请判断以下文本是否违规。\n\n检查维度：政治敏感、色情低俗、暴力恐怖、违法犯罪、人身攻击、广告营销\n\n文本内容：小区环境差\n\n请严格按照JSON格式输出。'},
        {"role": "assistant", "content": '{"is_safe": true, "risk_level": "safe", "categories": [], "reason": "正常的投诉反馈", "confidence": 0.95}'},
        
        # 实际要判断的内容
        {"role": "user", "content": user_prompt}
    ]

    # 应用聊天模板
    text = tokenizer.apply_chat_template(
        messages,
        tokenize=False,
        add_generation_prompt=True
    )

    # 编码输入
    inputs = tokenizer([text], return_tensors="pt").to(model.device)

    # 生成
    with torch.no_grad():
        outputs = model.generate(
            **inputs,
            max_new_tokens=300,
            temperature=0.1,
            do_sample=False,
            pad_token_id=tokenizer.eos_token_id
        )

    # 解码输出
    response = tokenizer.decode(
        outputs[0][len(inputs.input_ids[0]):],
        skip_special_tokens=True
    )

    # 解析JSON
    try:
        # 提取JSON部分
        json_match = re.search(r'\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}', response, re.DOTALL)
        if json_match:
            result = json.loads(json_match.group())
        else:
            raise ValueError(f"No valid JSON found in response: {response}")

        # 验证必需字段
        required_fields = ["is_safe", "risk_level", "categories", "reason", "confidence"]
        for field in required_fields:
            if field not in result:
                raise ValueError(f"Missing required field: {field}")

    except Exception as e:
        logger.error(f"Failed to parse model output: {e}")
        logger.error(f"Raw response: {response}")
        # 降级处理：保守策略，标记为需要人工审核
        result = {
            "is_safe": False,
            "risk_level": "medium",
            "categories": ["解析失败"],
            "reason": f"模型输出解析失败，需要人工复审。原始输出：{response[:100]}",
            "confidence": 0.5
        }

    latency_ms = int((time.time() - start_time) * 1000)
    result["latency_ms"] = latency_ms
    result["model_version"] = model_info["version"]

    # 更新统计
    stats["total_requests"] += 1
    stats["total_latency"] += latency_ms

    return result

# ============ API 端点 ============
@app.post("/api/moderate/text", response_model=TextModerationResponse)
async def moderate_text(req: TextModerationRequest):
    """
    文本审核接口
    """
    if not model_info["loaded"]:
        raise HTTPException(status_code=503, detail="Model not loaded")

    try:
        result = moderate_text_internal(req.content, req.check_categories)
        return TextModerationResponse(**result)
    except Exception as e:
        logger.error(f"Moderation error: {e}")
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health", response_model=HealthCheckResponse)
async def health_check():
    """
    健康检查
    """
    uptime = int(time.time() - model_info["start_time"])
    avg_latency = stats["total_latency"] / stats["total_requests"] if stats["total_requests"] > 0 else 0

    gpu_memory = None
    if torch.cuda.is_available():
        gpu_memory = torch.cuda.memory_allocated() / 1024**2

    return HealthCheckResponse(
        status="ok" if model_info["loaded"] else "error",
        model_loaded=model_info["loaded"],
        model_name=model_info["name"],
        model_version=model_info["version"],
        uptime_seconds=uptime,
        total_requests=stats["total_requests"],
        avg_latency_ms=avg_latency,
        gpu_available=torch.cuda.is_available(),
        gpu_memory_allocated_mb=gpu_memory
    )

@app.get("/model/info", response_model=ModelInfoResponse)
async def get_model_info():
    """
    获取模型信息
    """
    return ModelInfoResponse(
        name="Qwen2.5-7B-Instruct",
        version="v1.0",
        type="text",
        parameters=7_000_000_000,
        capabilities=[
            "政治敏感检测",
            "色情低俗检测",
            "暴力恐怖检测",
            "违法犯罪检测",
            "人身攻击检测",
            "广告营销检测"
        ],
        quantization="INT8"
    )

@app.get("/")
async def root():
    return {
        "service": "AI Model Service - Python Engine",
        "version": "1.0.0",
        "model": model_info["name"],
        "status": "running" if model_info["loaded"] else "loading"
    }

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8001, log_level="info")
