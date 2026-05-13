"""
Configuration management for AI Model Service
"""
import os
from typing import Optional
from pydantic import BaseSettings


class Settings(BaseSettings):
    """Application settings"""

    # Model Configuration
    model_name: str = "Qwen/Qwen2.5-7B-Instruct"
    model_cache_dir: str = "/app/models"
    device: str = "cuda"
    load_in_8bit: bool = True

    # Server Configuration
    host: str = "0.0.0.0"
    port: int = 8001
    workers: int = 1
    log_level: str = "info"

    # Inference Configuration
    max_length: int = 512
    temperature: float = 0.7
    top_p: float = 0.9
    top_k: int = 50

    # Performance
    batch_size: int = 1
    max_batch_wait_ms: int = 50

    # Cache (optional)
    enable_cache: bool = False
    redis_host: str = "localhost"
    redis_port: int = 6379
    cache_ttl_seconds: int = 3600

    class Config:
        env_file = ".env"
        case_sensitive = False


# Global settings instance
settings = Settings()
