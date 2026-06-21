# Community-and-Home 项目架构图

## 整体架构

```mermaid
graph TB
    subgraph "前端层 Frontend Layer"
        PC[管理后台<br/>web/pc<br/>Vue3 + TypeScript]
    end

    subgraph "API 网关层 API Gateway"
        APISIX[APISIX Gateway<br/>172.18.0.5:9080]
    end

    subgraph "微服务层 Microservices"
        subgraph "核心服务 Core Services"
            USER[用户服务<br/>user-service<br/>用户CRUD+手机号加密]
            AUTH[认证服务<br/>auth-service<br/>AT/RT双Token+SMS]
            PERM[权限服务<br/>permission-service<br/>RBAC权限模型]
        end

        subgraph "业务服务 Business Services"
            COMM[社区枢纽服务<br/>community-hub-service<br/>通知/联络/寻失]
            MOD[审核服务<br/>moderation-service<br/>AI内容审核]
            FILE[文件服务<br/>file-service<br/>MinIO上传下载]
        end

        subgraph "支撑服务 Support Services"
            MASTER[主数据服务<br/>master-data-service<br/>系统配置+主数据]
            AI[AI模型服务<br/>ai-model-service<br/>Go+Python混合]
            MONITOR[监控服务<br/>monitoring-service<br/>TCP/Docker/AI监控]
        end
    end

    subgraph "共享组件层 Shared Components"
        PROTO[api-proto<br/>Proto定义+代码生成<br/>Buf v2管理]
        COMMON[common库<br/>community-common/v2<br/>10个工具包]
    end

    subgraph "基础设施层 Infrastructure"
        MYSQL[(MySQL 8.0<br/>172.18.0.2:3306)]
        REDIS[(Redis 7<br/>172.18.0.4:6379)]
        ETCD[(etcd v3.5<br/>172.18.0.3:2379)]
        MINIO[(MinIO<br/>172.18.0.6:9000)]
        NEO4J[(Neo4j 5<br/>172.18.0.7:7687<br/>知识图谱)]
    end

    %% 前端到网关
    PC --> APISIX

    %% 网关到服务
    APISIX --> USER
    APISIX --> AUTH
    APISIX --> PERM
    APISIX --> COMM
    APISIX --> MOD
    APISIX --> FILE
    APISIX --> MASTER
    APISIX --> MONITOR

    %% 服务间gRPC依赖
    AUTH -.gRPC.-> USER
    MOD -.gRPC.-> AI
    MOD -.gRPC.-> MASTER

    %% 服务依赖共享组件
    USER --> PROTO
    AUTH --> PROTO
    PERM --> PROTO
    COMM --> PROTO
    MOD --> PROTO
    FILE --> PROTO
    MASTER --> PROTO
    AI --> PROTO

    USER --> COMMON
    AUTH --> COMMON
    PERM --> COMMON
    COMM --> COMMON
    MOD --> COMMON
    FILE --> COMMON
    MASTER --> COMMON
    MONITOR --> COMMON

    %% 服务到基础设施
    USER --> MYSQL
    AUTH --> MYSQL
    PERM --> MYSQL
    COMM --> MYSQL
    MOD --> MYSQL
    FILE --> MYSQL
    MASTER --> MYSQL
    AI --> MYSQL

    USER --> REDIS
    AUTH --> REDIS
    PERM --> REDIS
    MOD --> REDIS
    MASTER --> REDIS
    AI --> REDIS

    USER --> ETCD
    AUTH --> ETCD
    PERM --> ETCD
    MOD --> ETCD
    FILE --> ETCD
    MASTER --> ETCD
    AI --> ETCD

    FILE --> MINIO

    style PC fill:#e1f5ff
    style APISIX fill:#fff3e0
    style USER fill:#c8e6c9
    style AUTH fill:#c8e6c9
    style PERM fill:#c8e6c9
    style COMM fill:#b2dfdb
    style MOD fill:#b2dfdb
    style FILE fill:#b2dfdb
    style MASTER fill:#d1c4e9
    style AI fill:#d1c4e9
    style MONITOR fill:#d1c4e9
    style PROTO fill:#ffecb3
    style COMMON fill:#ffecb3
    style MYSQL fill:#ffcdd2
    style REDIS fill:#ffcdd2
    style ETCD fill:#ffcdd2
    style MINIO fill:#ffcdd2
    style NEO4J fill:#ffcdd2
```

## 服务间通信架构

```mermaid
graph LR
    subgraph "gRPC 服务间通信"
        AUTH[auth-service] -.GetUserByPhone<br/>CreateUser<br/>UpdateUser.-> USER[user-service]
        MOD[moderation-service] -.CallModel<br/>ModerateText.-> AI[ai-model-service]
        MOD -.GetConfig.-> MASTER[master-data-service]
    end

    subgraph "独立服务"
        PERM[permission-service<br/>无出站gRPC]
        FILE[file-service]
        COMM[community-hub-service]
        MONITOR[monitoring-service]
    end

    style AUTH fill:#c8e6c9
    style USER fill:#c8e6c9
    style MOD fill:#b2dfdb
    style AI fill:#d1c4e9
    style MASTER fill:#d1c4e9
    style PERM fill:#f8bbd0
    style FILE fill:#f8bbd0
    style COMM fill:#f8bbd0
    style MONITOR fill:#f8bbd0
```

## 数据存储架构

```mermaid
graph TB
    subgraph "数据库分布 Database Distribution"
        MYSQL[(MySQL 8.0)]
        
        subgraph "user_db"
            UT[users表<br/>user_phone_mapping表<br/>user_estate_relation表]
        end
        
        subgraph "auth_db"
            AT[refresh_tokens表<br/>sms_verification_codes表]
        end
        
        subgraph "permission_db"
            PT[rbac_roles表<br/>rbac_permissions表<br/>rbac_role_permissions表<br/>rbac_user_roles表]
        end
        
        subgraph "moderation_db"
            MT[moderation_pipelines表<br/>moderation_logs表<br/>sensitive_words表]
        end
        
        subgraph "masterdata_db"
            MDT[system_configs表<br/>provinces表<br/>cities表<br/>districts表]
        end
        
        subgraph "ai_model_db"
            AIT[model_configs表<br/>api_keys表<br/>templates表]
        end
        
        subgraph "file_db"
            FT[file_metadata表<br/>upload_sessions表]
        end
        
        subgraph "community_hub_db"
            CT[notifications表<br/>contacts表<br/>lost_and_found表]
        end
    end

    MYSQL --> user_db
    MYSQL --> auth_db
    MYSQL --> permission_db
    MYSQL --> moderation_db
    MYSQL --> masterdata_db
    MYSQL --> ai_model_db
    MYSQL --> file_db
    MYSQL --> community_hub_db

    subgraph "缓存层 Cache Layer"
        REDIS[(Redis 7)]
        RC1[用户缓存]
        RC2[Token缓存]
        RC3[权限缓存]
        RC4[系统配置缓存]
        RC5[模型配置缓存]
    end

    REDIS --> RC1
    REDIS --> RC2
    REDIS --> RC3
    REDIS --> RC4
    REDIS --> RC5

    subgraph "对象存储 Object Storage"
        MINIO[(MinIO)]
        MB[用户头像]
        MB2[文件上传]
    end

    MINIO --> MB
    MINIO --> MB2

    subgraph "服务注册 Service Registry"
        ETCD[(etcd v3.5)]
        ES[服务发现]
        EC[配置中心]
    end

    ETCD --> ES
    ETCD --> EC

    subgraph "知识图谱 Knowledge Graph"
        NEO4J[(Neo4j 5)]
        NG[服务依赖关系]
        NG2[实体血缘追踪]
    end

    NEO4J --> NG
    NEO4J --> NG2

    style MYSQL fill:#ffcdd2
    style REDIS fill:#b2ebf2
    style MINIO fill:#c5e1a5
    style ETCD fill:#ffe0b2
    style NEO4J fill:#d1c4e9
```

## 服务详细信息

| 服务 | 端口 | 职责 | 依赖服务 | 数据库 |
|------|------|------|---------|--------|
| **user-service** | RPC:8081, API:8001 | 用户CRUD、手机号加密、小区归属 | - | user_db |
| **auth-service** | RPC:8082 | AT/RT双Token、SMS验证码、RSA加密 | user-service | auth_db |
| **permission-service** | RPC:8083 | RBAC权限模型、角色权限管理 | - | permission_db |
| **master-data-service** | RPC:8085, API:8005 | 系统配置、主数据管理、Outbox模式 | - | masterdata_db |
| **moderation-service** | RPC:8086, API:8006 | AI内容审核、Pipeline引擎 | ai-model-service, master-data-service | moderation_db |
| **ai-model-service** | RPC:8084, API:8004 | Go+Python混合、模型管理、API密钥 | - | ai_model_db |
| **file-service** | RPC:8087, API:8007 | MinIO上传下载、文件元数据 | - | file_db |
| **community-hub-service** | RPC:8088, API:8008 | 通知、联络、寻失 | - | community_hub_db |
| **monitoring-service** | API:8009 | TCP/Docker/AI三层监控 | - | - |

## 技术栈

### 后端
- **语言**: Go 1.23.4
- **框架**: go-zero (REST + gRPC)
- **ORM**: GORM
- **Proto管理**: Buf v2
- **服务注册**: etcd v3.5

### 前端
- **框架**: Vue 3 + TypeScript
- **构建**: Vite
- **路由**: Vue Router
- **状态管理**: Pinia

### 基础设施
- **数据库**: MySQL 8.0
- **缓存**: Redis 7
- **对象存储**: MinIO
- **API网关**: Apache APISIX
- **知识图谱**: Neo4j 5
- **容器编排**: Docker Compose

### 开发工具
- **代码生成**: goctl (go-zero官方工具)
- **Proto生成**: protoc + buf
- **工作区**: Go Workspace (go.work)
- **AI辅助**: Claude Code + Harness 架构

## 错误码规范

格式：5位 `XXYYY`
- `XX` = 服务中心
  - 99: Common
  - 10: User
  - 50: Auth
  - 06: Permission
  - 07: File
  - 30: AI Model
  - 40: Moderation
- `YYY` = 具体错误（001-999）
- `0` = 成功

示例：
- `10001`: 用户服务错误
- `50400`: 认证失败
- `06403`: 权限不足
- `30500`: AI模型服务内部错误

## 部署架构

```mermaid
graph TB
    subgraph "本地开发环境"
        DEV[开发者机器]
        DC[Docker Compose<br/>所有中间件]
    end

    subgraph "生产环境（未来）"
        K8S[Kubernetes集群]
        
        subgraph "服务Pod"
            POD1[user-service pods]
            POD2[auth-service pods]
            POD3[其他服务 pods]
        end
        
        subgraph "中间件"
            MYSQL_PROD[(MySQL主从)]
            REDIS_PROD[(Redis Cluster)]
            ETCD_PROD[(etcd Cluster)]
        end
    end

    DEV --> DC
    DC --> MYSQL_PROD
    K8S --> POD1
    K8S --> POD2
    K8S --> POD3
    POD1 --> MYSQL_PROD
    POD1 --> REDIS_PROD
    POD1 --> ETCD_PROD

    style DEV fill:#e1f5ff
    style DC fill:#fff3e0
    style K8S fill:#c8e6c9
    style MYSQL_PROD fill:#ffcdd2
    style REDIS_PROD fill:#ffcdd2
    style ETCD_PROD fill:#ffcdd2
```

---

**生成时间**: 2026-06-21  
**维护者**: Global Architecture Coordinator (Kiro)  
**数据来源**: `.harness/rules/工程结构.md`, `.harness/knowledge/INDEX.md`, 各服务 `docs/design.md`
