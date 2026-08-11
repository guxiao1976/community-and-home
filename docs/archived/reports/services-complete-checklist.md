# 系统服务完整清单与健康检查

## 服务架构完整清单

### API网关层（REST接口）

| 服务 | 端口 | 配置文件 | 状态 |
|------|------|----------|------|
| auth-service API | 8881 | services/auth-service/api/etc/auth-api.yaml | ✅ |
| file-service API | 8884 | services/file-service/api/etc/file-api.yaml | ✅ |
| moderation-service API | 8886 | services/moderation-service/api/etc/moderation-api.yaml | ✅ |
| community-hub-service API | 8887 | services/community-hub-service/api/etc/communityhub-api.yaml | ✅ |
| **master-data-service API** | **8889** | services/master-data-service/api/etc/masterdata-api.yaml | ✅ |
| user-service API | 8882 | services/user-service/api/etc/user-api.yaml | ⏸️ |
| ai-model-service API | 8883 | services/ai-model-service/api/etc/aimodelapi.yaml | ⏸️ |
| monitoring-service API | 8888 | services/monitoring-service/api/etc/monitoring-api.yaml | ⏸️ |

### RPC服务层（gRPC接口）

| 服务 | 端口 | 配置文件 | 状态 |
|------|------|----------|------|
| auth-service RPC | 8083 | services/auth-service/rpc/etc/authservice.yaml | ✅ |
| user-service RPC | 8084 | services/user-service/rpc/etc/userservice.yaml | ✅ |
| file-service RPC | 8085 | services/file-service/rpc/etc/fileservice.yaml | ✅ |
| moderation-service RPC | 8086 | services/moderation-service/rpc/etc/moderation.yaml | ❌ |
| master-data-service RPC | 8087 | services/master-data-service/rpc/etc/masterdata.yaml | ✅ |
| community-hub-service RPC | 8088 | services/community-hub-service/rpc/etc/communityhub.yaml | ✅ |

### 前端层

| 应用 | 端口 | 目录 | 状态 |
|------|------|------|------|
| PC端 | 3003 | web/pc | ✅ |
| 移动端 | 3004 | web/mobile | ⏸️ |

---

## 为什么会遗漏 master-data-service API？

### 1. 初始启动策略
启动服务时采用了"最小可用集"策略，优先启动核心用户功能：
- ✅ 认证服务（用户登录必需）
- ✅ 社区服务（核心业务功能）
- ✅ 文件服务（文件上传必需）
- ✅ 审核服务（内容安全必需）

而主数据管理被误判为"后台管理功能"，不在"最小可用集"中。

### 2. 测试范围不完整
初始测试重点：
- ✅ 用户注册/登录
- ✅ 短信验证码
- ✅ 文件上传
- ❌ 后台管理功能（未测试）

### 3. 服务依赖关系理解不足
误认为主数据管理服务只被管理员使用，实际上：
- 行政区划数据：用户注册时选择地址
- 住宅小区数据：用户绑定小区
- 敏感词库：内容审核依赖
- 系统配置：多个服务读取

这些都是系统运行的基础数据，应该在启动时就准备好。

---

## 502错误的技术细节

### 什么是502错误？

**502 Bad Gateway** = 网关收到上游服务器的无效响应

### 请求流程

```
前端浏览器 (localhost:3003)
    ↓ 发起请求
    ↓ GET /masterdata/divisions
    ↓
前端开发服务器 (Vite)
    ↓ 代理转发
    ↓ proxy: { '/api': 'http://localhost' }
    ↓
目标服务器 (localhost:8889)
    ↓
    ✗ 连接被拒绝 (Connection refused)
    ↓
    ← 502 Bad Gateway
```

### 为什么是502而不是404？

- **404 Not Found**: 服务器在线，但找不到资源
- **502 Bad Gateway**: 网关无法连接到上游服务器
- **503 Service Unavailable**: 服务器过载或维护中

本例中，端口8889没有服务监听，网关连接失败 → 502

---

## 如何避免类似问题？

### 方法1: 完整的服务清单

创建 `docs/services-checklist.md`，列出所有服务：

```markdown
## 必须启动的服务

### API网关层 (5个)
- [ ] auth-service API (8881)
- [ ] file-service API (8884)
- [ ] moderation-service API (8886)
- [ ] community-hub-service API (8887)
- [ ] master-data-service API (8889)

### RPC服务层 (5个)
- [ ] auth-service RPC (8083)
- [ ] user-service RPC (8084)
- [ ] file-service RPC (8085)
- [ ] master-data-service RPC (8087)
- [ ] community-hub-service RPC (8088)
```

### 方法2: 自动化健康检查

创建 `scripts/health-check.sh`：

```bash
#!/bin/bash
# 检查所有必需服务

REQUIRED_PORTS=(8881 8884 8886 8887 8889 8083 8084 8085 8087 8088 3003)

for port in "${REQUIRED_PORTS[@]}"; do
    if netstat -tln | grep ":$port " > /dev/null; then
        echo "✅ 端口 $port 正常"
    else
        echo "❌ 端口 $port 未监听"
    fi
done
```

### 方法3: 功能测试清单

启动后测试所有主要功能：

```
用户功能:
- [ ] 用户注册
- [ ] 用户登录
- [ ] 个人信息

社区功能:
- [ ] 公告列表
- [ ] 便民联络
- [ ] 寻失互助

后台管理:
- [ ] 行政区划  ← 之前遗漏
- [ ] 住宅小区  ← 之前遗漏
- [ ] 敏感词管理 ← 之前遗漏
```

---

## 已采取的改进措施

### 1. ✅ 更新启动脚本
修改 `scripts/start-all-apis.sh`，添加 master-data-service

### 2. ✅ 创建测试脚本
创建 `/tmp/test_masterdata_apis.sh` 测试主数据接口

### 3. ✅ 更新文档
创建 `docs/masterdata-502-fix-report.md` 记录问题和解决方案

### 4. ✅ 完整的服务清单
本文档列出所有服务，避免再次遗漏

---

## 总结

### 根本原因
**启动脚本遗漏了 master-data-service API**

### 触发条件
用户点击主数据管理菜单（行政区划、住宅小区、敏感词）

### 错误表现
```javascript
Object { 
  message: "Request failed with status code 502", 
  status: 502, 
  data: "" 
}
```

### 解决方案
启动 master-data-service API (端口 8889)

### 预防措施
1. 使用完整的服务清单
2. 执行系统化的功能测试
3. 使用自动化健康检查脚本

---

**经验教训**: 系统启动时应该启动所有核心服务，而不是"最小可用集"。后台管理功能的数据服务往往也是前台功能的依赖。
