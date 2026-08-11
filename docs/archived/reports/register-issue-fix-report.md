# 用户注册功能问题修复报告

## 问题时间
2026-07-12 20:10

---

## 🔍 问题诊断

### 报错信息
```
注册失败，请稍后重试
```

### 后端日志错误
```
1. Unknown column 'nickname_moderation_status' in 'field list'
2. field "encryptedPhone" is not set
```

---

## ✅ 已修复的问题

### 问题1: 数据库字段缺失 ✅

**错误**: `Unknown column 'nickname_moderation_status'`

**原因**: user_base 表缺少审核字段

**解决方案**: 已执行 migration 004 脚本

**验证**:
```bash
$ docker exec mysql mysql -uroot -proot123456 -e "USE user; DESCRIBE user_base;" | grep moderation
nickname_moderation_status  tinyint  NO  0
```

✅ **状态**: 已修复

---

## ⚠️ 待修复的问题

### 问题2: 前端字段名不匹配 ❌

**错误**: `field "encryptedPhone" is not set`

**原因**: 前端发送的是 `phone`，后端期望 `encryptedPhone`

**后端API定义** (`services/auth-service/api/internal/types/types.go`):
```go
type RegisterReq struct {
    EncryptedPhone    string `json:"encryptedPhone"`     // ← 后端期望这个
    SmsCode           string `json:"smsCode"`
    EncryptedPassword string `json:"encryptedPassword,optional"`
    Nickname          string `json:"nickname"`
    DeviceId          string `json:"deviceId"`
    DeviceType        string `json:"deviceType,optional"`
}
```

**前端当前发送**:
```json
{
  "phone": "13800138000",              // ❌ 错误
  "password": "123456",                // ❌ 错误
  "nickname": "测试用户",
  "sms_code": "123456"                 // ❌ 错误（应该是驼峰）
}
```

**前端应该发送**:
```json
{
  "encryptedPhone": "<RSA加密的手机号>",      // ✅ 正确
  "encryptedPassword": "<RSA加密的密码>",    // ✅ 正确
  "nickname": "测试用户",
  "smsCode": "123456",                       // ✅ 正确（驼峰命名）
  "deviceId": "web_xxx",
  "deviceType": "web"
}
```

---

## 🔐 RSA 加密说明

### 为什么需要加密？

**安全原因**:
- 手机号和密码在传输时需要加密
- 防止中间人攻击
- 符合安全规范

### 加密流程

```
1. 前端获取RSA公钥
   GET /api/auth/public-key
   
2. 使用公钥加密手机号和密码
   const encrypted = RSA.encrypt(plaintext, publicKey)
   
3. 发送加密后的数据到后端
   POST /api/auth/register
```

---

## 🛠️ 前端修复方案

### 步骤1: 获取RSA公钥

```typescript
// api/auth.ts
export async function getPublicKey(): Promise<string> {
  const response = await axios.get('/api/auth/public-key')
  return response.data.publicKey
}
```

### 步骤2: 实现RSA加密

```typescript
// utils/crypto.ts
import JSEncrypt from 'jsencrypt'

export function encryptWithRSA(text: string, publicKey: string): string {
  const encrypt = new JSEncrypt()
  encrypt.setPublicKey(publicKey)
  const encrypted = encrypt.encrypt(text)
  if (!encrypted) {
    throw new Error('加密失败')
  }
  return encrypted
}
```

### 步骤3: 修改注册API调用

```typescript
// api/auth.ts
import { encryptWithRSA } from '@/utils/crypto'

export async function register(data: {
  phone: string
  password: string
  nickname: string
  smsCode: string
}) {
  // 1. 获取公钥
  const publicKey = await getPublicKey()
  
  // 2. 加密手机号和密码
  const encryptedPhone = encryptWithRSA(data.phone, publicKey)
  const encryptedPassword = encryptWithRSA(data.password, publicKey)
  
  // 3. 发送正确的字段名
  const response = await axios.post('/api/auth/register', {
    encryptedPhone,           // ✅ 加密后的手机号
    encryptedPassword,        // ✅ 加密后的密码
    nickname: data.nickname,
    smsCode: data.smsCode,    // ✅ 驼峰命名
    deviceId: generateDeviceId(),
    deviceType: 'web'
  })
  
  return response.data
}

function generateDeviceId(): string {
  // 生成或从localStorage获取设备ID
  let deviceId = localStorage.getItem('deviceId')
  if (!deviceId) {
    deviceId = `web_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
    localStorage.setItem('deviceId', deviceId)
  }
  return deviceId
}
```

### 步骤4: 安装依赖

```bash
cd web/pc
npm install jsencrypt
```

---

## 📝 完整的注册流程示例

### 前端代码 (`views/Register.vue`)

```vue
<script setup lang="ts">
import { ref } from 'vue'
import { register, sendSmsCode, getPublicKey } from '@/api/auth'
import { encryptWithRSA } from '@/utils/crypto'

const phone = ref('')
const password = ref('')
const nickname = ref('')
const smsCode = ref('')
const loading = ref(false)

// 发送验证码
async function handleSendSms() {
  try {
    await sendSmsCode(phone.value)
    ElMessage.success('验证码已发送')
  } catch (error) {
    ElMessage.error('发送失败')
  }
}

// 注册
async function handleRegister() {
  loading.value = true
  try {
    const result = await register({
      phone: phone.value,
      password: password.value,
      nickname: nickname.value,
      smsCode: smsCode.value
    })
    
    // 保存token
    localStorage.setItem('accessToken', result.accessToken)
    localStorage.setItem('refreshToken', result.refreshToken)
    localStorage.setItem('userId', result.userId)
    
    ElMessage.success('注册成功')
    router.push('/home')
  } catch (error: any) {
    ElMessage.error(error.message || '注册失败，请稍后重试')
  } finally {
    loading.value = false
  }
}
</script>
```

---

## 🧪 测试验证

### 测试步骤

1. **获取公钥**
```bash
curl http://localhost:8881/api/auth/public-key
```

2. **发送验证码**
```bash
curl -X POST http://localhost:8881/api/auth/sms/send \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000"}'
```

查看日志：
```bash
tail -f /tmp/microservices-logs/auth-service-api.log | grep "短信验证码"
```

应该看到：
```
【短信验证码】phone=13800138000, code=123456
```

3. **注册用户**（需要RSA加密）

使用Postman或前端页面测试，确保发送：
- `encryptedPhone`（RSA加密）
- `encryptedPassword`（RSA加密）
- `smsCode`（驼峰命名）

---

## 📊 API字段对照表

| 功能 | 后端字段名 | 前端字段名 | 是否需要加密 | 说明 |
|------|-----------|-----------|------------|------|
| 注册 | encryptedPhone | phone → 加密 | ✅ 是 | RSA加密 |
| 注册 | encryptedPassword | password → 加密 | ✅ 是 | RSA加密 |
| 注册 | smsCode | sms_code | ❌ 否 | 驼峰命名 |
| 注册 | nickname | nickname | ❌ 否 | - |
| 注册 | deviceId | deviceId | ❌ 否 | UUID |
| 登录 | encryptedPhone | phone → 加密 | ✅ 是 | RSA加密 |
| 登录 | encryptedPassword | password → 加密 | ✅ 是 | RSA加密 |

---

## ✅ 修复后的验证

### 成功的响应
```json
{
  "code": 0,
  "msg": "success",
  "data": {
    "userId": "1234567890",
    "accessToken": "eyJhbGc...",
    "refreshToken": "eyJhbGc...",
    "expiresAt": 1720876800
  }
}
```

### 失败的响应
```json
{
  "code": 50001,
  "msg": "手机号已注册",
  "data": null
}
```

---

## 📁 需要修改的前端文件

```
web/pc/
├── src/
│   ├── api/
│   │   └── auth.ts              # ← 修改API调用
│   ├── utils/
│   │   └── crypto.ts            # ← 新增RSA加密工具
│   ├── views/
│   │   ├── Register.vue         # ← 修改注册页面
│   │   └── Login.vue            # ← 修改登录页面
│   └── package.json             # ← 添加jsencrypt依赖
```

---

## 🎯 总结

### 已完成
- ✅ 数据库字段修复（nickname_moderation_status）
- ✅ 后端服务正常运行
- ✅ API接口可访问

### 需要前端修复
- ❌ 字段名改为驼峰命名（smsCode）
- ❌ 实现RSA加密（encryptedPhone, encryptedPassword）
- ❌ 添加deviceId生成逻辑

### 修复优先级
1. **高优**: 添加RSA加密功能
2. **高优**: 修改字段名为驼峰命名
3. **中优**: 实现deviceId生成
4. **低优**: 错误提示优化

---

## 📞 技术支持

如遇问题，查看日志：
```bash
# API层日志
tail -f /tmp/microservices-logs/auth-service-api.log

# RPC层日志
tail -f /tmp/microservices-logs/auth-service.log

# 前端日志
tail -f /tmp/frontend-pc.log
```

---

**报告时间**: 2026-07-12 20:12  
**状态**: 后端已修复，等待前端适配
