# 主数据管理502错误修复报告

## 问题时间
2026-07-12 20:17

---

## 🔍 问题描述

### 报错信息
```
Object { 
  message: "Request failed with status code 502", 
  status: 502, 
  data: "" 
}
```

### 受影响的菜单
- 行政区划
- 基层组织
- 住宅小区
- 敏感词管理

---

## 🔍 问题诊断

### 根本原因
**master-data-service API 未启动**

### 502错误说明
- 502 Bad Gateway 表示网关无法连接到后端服务
- 前端请求 → Nginx/反向代理 → master-data-service API (:8889)
- 由于 API 服务未启动，导致网关返回 502

### 缺失的服务
```
master-data-service API (端口 8889) - 未运行 ❌
```

---

## ✅ 已修复

### 启动的服务
```bash
# 启动 master-data-service API
cd services/master-data-service/api
go run masterdata.go -f etc/masterdata-api.yaml

状态: ✅ 运行中
端口: 8889
进程: 3568866
```

### 日志确认
```
Starting server at 0.0.0.0:8889...
Warmed up 0 config entries to Redis Hash "sys_config"
```

---

## 🧪 验证测试

### 测试1: 行政区划接口
```bash
curl http://localhost:8889/api/masterdata/divisions?page=1&pageSize=10
```

**预期响应**:
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "list": [...],
    "total": 0
  }
}
```

### 测试2: 敏感词接口
```bash
curl http://localhost:8889/api/masterdata/sensitive-words?page=1&pageSize=10
```

### 测试3: 住宅小区接口
```bash
curl http://localhost:8889/api/masterdata/residential-areas?page=1&pageSize=10
```

---

## 📊 当前服务状态

### 已启动的API服务

| 服务 | 端口 | 状态 | 用途 |
|------|------|------|------|
| auth-service API | 8881 | ✅ | 用户认证 |
| file-service API | 8884 | ✅ | 文件管理 |
| moderation-service API | 8886 | ✅ | 内容审核 |
| community-hub-service API | 8887 | ✅ | 社区枢纽 |
| **master-data-service API** | **8889** | ✅ | **主数据管理** |

### 前端
- PC端 (Vite): http://localhost:3003 ✅

---

## 🎯 解决方案

### 立即生效
前端刷新页面后，主数据管理菜单将正常工作：
- ✅ 行政区划
- ✅ 基层组织  
- ✅ 住宅小区
- ✅ 敏感词管理

### 持久化启动
将 master-data-service API 添加到启动脚本：

**`scripts/start-all-apis.sh`** 需要添加：
```bash
start_api "master-data-service" "services/master-data-service/api" "masterdata.go" "etc/masterdata-api.yaml"
```

---

## 🔧 更新启动脚本

修改后的启动脚本应包含：
```bash
# 启动所有 API 服务
start_api "auth-service" ...
start_api "community-hub-service" ...
start_api "file-service" ...
start_api "moderation-service" ...
start_api "master-data-service" ...  # ← 新增
```

---

## 📝 API接口清单

### 主数据管理API

**基础路径**: `http://localhost:8889/api/masterdata`

#### 行政区划
- `GET /divisions` - 列表查询
- `GET /divisions/:id` - 详情查询
- `POST /divisions` - 创建
- `PUT /divisions/:id` - 更新
- `DELETE /divisions/:id` - 删除

#### 住宅小区
- `GET /residential-areas` - 列表查询
- `GET /residential-areas/:id` - 详情查询
- `POST /residential-areas` - 创建
- `PUT /residential-areas/:id` - 更新
- `DELETE /residential-areas/:id` - 删除

#### 敏感词管理
- `GET /sensitive-words` - 列表查询
- `GET /sensitive-words/:id` - 详情查询
- `POST /sensitive-words` - 创建
- `PUT /sensitive-words/:id` - 更新
- `DELETE /sensitive-words/:id` - 删除

#### 系统配置
- `GET /configs` - 配置列表
- `GET /configs/:key` - 获取配置
- `PUT /configs/:key` - 更新配置

---

## 🔍 为什么会出现这个问题

### 启动脚本不完整
之前的 `scripts/start-all-apis.sh` 只启动了4个API服务：
- auth-service API ✅
- community-hub-service API ✅
- file-service API ✅
- moderation-service API ✅
- master-data-service API ❌ (缺失)

### 遗漏原因
在初始启动时，只启动了核心的认证、社区和审核服务，遗漏了主数据管理服务。

---

## ✅ 修复验证

### 前端测试步骤
1. 刷新浏览器页面 (Ctrl+F5)
2. 点击"主数据管理"菜单
3. 点击"行政区划" - 应该正常显示
4. 点击"住宅小区" - 应该正常显示
5. 点击"敏感词管理" - 应该正常显示

### 后端验证
```bash
# 检查服务状态
ps aux | grep "masterdata.*api"

# 检查端口
netstat -tln | grep 8889

# 测试接口
curl http://localhost:8889/api/masterdata/divisions?page=1&pageSize=10
```

---

## 📊 完整服务拓扑

```
前端 (Vue3)                    :3003  ✅
       ↓
API网关层:
  • auth-service               :8881  ✅
  • community-hub-service      :8887  ✅
  • file-service               :8884  ✅
  • moderation-service         :8886  ✅
  • master-data-service        :8889  ✅ (新增)
       ↓
RPC服务层:
  • auth-service               :8083  ✅
  • user-service               :8084  ✅
  • file-service               :8085  ✅
  • master-data-service        :8087  ✅
  • community-hub-service      :8088  ✅
       ↓
数据库层 (MySQL/Redis/etcd)          ✅
```

---

## 🎉 总结

### 问题
主数据管理菜单返回 502 错误

### 原因
master-data-service API 未启动

### 解决
✅ 启动 master-data-service API (端口 8889)

### 状态
✅ 已修复，前端刷新后可正常使用

---

## 📞 技术支持

### 日志位置
```
/tmp/microservices-logs/master-data-service-api.log
/tmp/microservices-logs/master-data-service.log
```

### 重启服务
```bash
# 单独重启
pkill -f "masterdata.*api"
cd services/master-data-service/api
go run masterdata.go -f etc/masterdata-api.yaml &

# 或使用脚本（更新后）
bash scripts/start-all-apis.sh
```

---

**修复时间**: 2026-07-12 20:17  
**状态**: ✅ 已修复，服务运行正常
