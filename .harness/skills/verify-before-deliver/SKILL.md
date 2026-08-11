# verify-before-deliver

## 触发条件

修改了后端代码或前端代码，准备让用户验证时，**必须先完成本检查**。触发词：`自验证`、`交付前检查`、`verify before deliver`。

## 角色

你是交付前的最后一道防线。你必须**自己完成所有验证**，确认没有问题，才能对用户说"请刷新试试"。

## 🚫 硬性禁止

**以下行为在完成本 skill 全部步骤前绝对禁止：**

- ❌ 对用户说"请刷新试试"
- ❌ 对用户说"请点击测试"
- ❌ 对用户说"看看现在行不行"

## ✅ 必须完成的 6 步

---

### Step 1: 确认进程运行的是最新代码

```bash
# 检查进程 PID 和启动时间
ps aux | grep "<binary-name>" | grep -v grep

# 检查二进制文件修改时间 vs 进程启动时间
ls -la /tmp/go-build*/exe/<service>
```

**判定**：进程启动时间 < 代码修改时间 → 旧进程，必须重启。

**操作**：
```bash
lsof -ti:<PORT> | xargs kill -9
# 然后用 start.sh 重启对应服务
```

✅ **Pass**：进程启动时间晚于最后一次代码修改。

---

### Step 2: 自查后端 API 返回（不需要用户帮忙）

原则：**用 dev-token 直接调 API，覆盖所有路由类型**。

#### 2a. 获取调试 Token

```bash
TOKEN=$(bash .harness/scripts/dev-token)
# 生产环境此脚本自动拒绝执行
```

#### 2b. 必须测试的路由类型（逐项勾选）

**⚠️ 必须覆盖所有类型的路由，禁止只测列表不测详情**。

```
□ 列表 (GET 无参数)    curl -s "http://<PORT>/api/users?page=1" -H "Authorization: Bearer $TOKEN"
□ 详情 (GET :id)       curl -s "http://<PORT>/api/users/4542136688377323520" -H "Auth..."
□ 创建 (POST)          curl -s -X POST "http://<PORT>/api/perm/roles" -H "Auth..." -d '{...}'
□ 更新 (PUT :id)       curl -s -X PUT "http://<PORT>/api/perm/roles/3" -H "Auth..." -d '{...}'
□ 删除 (DELETE :id)    curl -s -X DELETE "http://<PORT>/api/perm/roles/<temp_id>" -H "Auth..."
```

**关键检查**：`:id` 路由的 URL 中填入真实数字 ID，验证权限系统能正确匹配模式 `GET:/api/users/:id`。

每项返回 `code=0` 才算 ✅。

#### 2c. 失败时怎么办

```
1. 加日志
   l.Logger.Infof("DEBUG: key=%v", value)

2. 查日志
   tail -50 /tmp/<service>.log | grep "DEBUG\|error\|panic"

3. 查缓存格式
   docker exec redis redis-cli SMEMBERS "perm:user:<user_id>"
   # 确认无双重前缀(GET:GET:...) 且 :id 模式匹配正确

4. 修复 → 回到 2b 重新测试
```

✅ **Pass**：所有路由类型（列表/详情/创建/更新/删除）均返回 code=0。

---

### Step 3: 确认数据状态正确

涉及数据库或缓存变更时，**必须直接查库确认**。

```bash
# MySQL
docker exec mysql mysql -uroot -p<PASS> -e "SELECT ... FROM <db>.<table>"

# Redis
docker exec redis redis-cli GET/SMEMBERS/KEYS "<key>"
```

**检查项**：
- [ ] 数据已写入/已变更，值符合预期
- [ ] 无重复记录（UNIQUE 约束冲突）
- [ ] Redis 缓存格式正确、TTL 合理

✅ **Pass**：数据库和缓存状态与预期一致。

---

### Step 4: 确认前端编译通过

```bash
cd web/pc
npx vue-tsc --noEmit 2>&1    # 或 npm run lint
```

- [ ] 无类型错误
- [ ] 无 import 路径错误
- [ ] API 返回类型与前端类型定义一致

✅ **Pass**：类型检查无错误。

---

### Step 5: 输出验证报告

所有步骤通过后，输出：

```
## 验证报告

| 步骤 | 状态 | 详情 |
|------|------|------|
| 1. 进程重启 | ✅ | perm-api PID 1234 (19:30) |
| 2. API 自查 | ✅ | POST /api/xxx → 200, data: [...] |
| 3. 数据状态 | ✅ | sys_permission 21 条 type=3，无重复 |
| 4. 前端编译 | ✅ | vue-tsc exit 0 |
```

**此时才能对用户说：请刷新测试。**

---

### Step 6: 用户反馈后快速诊断

如果用户反馈"还是不行"：

1. **先看浏览器日志**（用户提供或问我）
2. **定位到具体 API 调用**：`GET /api/xxx → [200] { code: xxx, message: "..." }`
3. **区分 HTTP 错误 vs 业务错误**：
   - HTTP 401 → Token 问题
   - HTTP 200 + code=99401 → 权限/缓存问题
   - HTTP 200 + code=0 → 前端逻辑问题
4. **回到 Step 2**，用 curl 复现，加日志，修好再交付
