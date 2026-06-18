# File Service 设计方案

## 一、定位

`file-service` 是社区平台的 **统一文件存储服务**，基于 MinIO 对象存储。RPC+API 双层：gRPC 供其他服务调用，REST API 供前端直接使用。

### 1.1 做什么

| 职责 | 说明 |
|------|------|
| 获取上传 URL | 生成 MinIO 预签名 PUT URL（15 分钟有效），客户端直传 |
| 确认上传完成 | 客户端上传完成后回调，写入 DB 元数据 |
| 获取下载 URL | 生成 MinIO 预签名 GET URL（默认 1 小时，最长 7 天） |
| 文件删除 | 软删除 DB + 移除 MinIO 对象（MinIO 失败不阻塞 DB） |
| 文件列表 | 分页查询，支持 user_id / entity_type / entity_id 过滤 |

### 1.2 不做什么

| 不负责 | 归属 |
|--------|------|
| 图片处理（裁剪/压缩/水印） | —（未来可加） |
| 文件病毒扫描 | — |
| CDN 加速 | 基础设施层 |
| 用户配额管理 | — |

### 1.3 核心设计决策

- **客户端直传模式**：服务生成预签名 URL → 客户端直传 MinIO，文件流不经过本服务
- **DB 优先一致性**：删除时 DB 先成功，MinIO 失败仅记录日志（DB 一致性优先于对象存储清理）
- **软删除**：`is_deleted=1`，不物理删除 DB 记录
- **Object Key 命名**：`uploads/{user_id}/{timestamp_nano}_{filename}` — 用户隔离 + 纳秒时间戳防碰撞
- **双 MinIO 客户端**：common 封装用于下载/删除，原始 minio-go 用于上传预签名（封装未暴露该 API）

---

## 二、数据库设计（`file_db`，1 张表）

### 2.1 `uploaded_file`

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int64 PK | 自增，JSON 输出为 string（Snowflake 兼容） |
| `user_id` | int64 INDEX | 上传用户 |
| `entity_type` | varchar(64) | 业务关联类型：verification/avatar/article |
| `entity_id` | int64 INDEX | 关联的业务实体 ID，0=未关联 |
| `file_name` | varchar(255) | 原始文件名 |
| `file_path` | varchar(512) | MinIO object key |
| `file_size` | int64 | 文件大小（字节） |
| `mime_type` | varchar(128) | MIME 类型 |
| `bucket_name` | varchar(64) | 默认 "community-home" |
| `upload_time` | datetime | 上传完成时间 |
| `is_deleted` | boolean INDEX | 软删除标记，默认 false |

**Model 层**：使用 `sqlx.SqlConn` 原始 SQL（非 GORM），5 个方法：Insert、FindOne、FindByIds、FindPage（动态 WHERE + 分页 + total）、Delete（SET is_deleted=1）。

---

## 三、上传/下载流程

### 3.1 上传（3 步）

```
Step 1: Client → POST /api/files/upload-url
  → gRPC GetUploadUrl → RawMinio.PresignedPutObject(15min)
  → 返回 upload_url, object_key, expire_at

Step 2: Client → HTTP PUT upload_url → MinIO（直传，不经过本服务）

Step 3: Client → POST /api/files/confirm
  → gRPC ConfirmUpload → FileModel.Insert(元数据)
  → 返回 FileInfo（含 id）
```

### 3.2 下载

```
Client → GET /api/files/:id
  → gRPC GetFileUrl
    → FileModel.FindOne(file_id) → MinioCli.GetURL(filePath, expireDuration)
  → 返回 download_url, FileInfo
```

### 3.3 删除

```
Client → DELETE /api/files/:id
  → gRPC DeleteFile
    → FileModel.FindOne → MinioCli.Delete(filePath) [失败只记日志]
    → FileModel.Delete(file_id) [SET is_deleted=1]
```

---

## 四、gRPC 接口

Proto: `api-proto/api/file/v1/file.proto`，5 个 RPC（全部 JWT）：

| RPC | 说明 | 超时 |
|-----|------|:---:|
| `GetUploadUrl` | 获取预签名上传 URL | 2s |
| `ConfirmUpload` | 确认上传完成，写入元数据 | 2s |
| `GetFileUrl` | 获取预签名下载 URL | 1s |
| `DeleteFile` | 软删除 | 1s |
| `ListFiles` | 分页列表（可选过滤条件） | 1s |

## 五、REST API

所有端点前缀 `/api/files`，全部 JWT 保护：

| 方法 | 路径 | 对应 gRPC |
|------|------|-----------|
| POST | `/api/files/upload-url` | GetUploadUrl |
| POST | `/api/files/confirm` | ConfirmUpload |
| GET | `/api/files` | ListFiles |
| GET | `/api/files/:id` | GetFileUrl |
| DELETE | `/api/files/:id` | DeleteFile |

---

## 六、配置

```yaml
# rpc/etc/fileservice.yaml
Name: file.rpc
ListenOn: 0.0.0.0:8085
Etcd:
  Hosts: [127.0.0.1:2379]
  Key: file.rpc
DataSource:
  Driver: mysql
  Source: root:root123456@tcp(127.0.0.1:3306)/file_db?charset=utf8mb4&parseTime=true&loc=Asia%2FShanghai
Minio:
  Endpoint: localhost:9000
  AccessKey: admin
  SecretKey: admin123
  Bucket: community-home
  UseSSL: false
```

---

## 七、依赖

| 依赖 | 说明 |
|------|------|
| `api-proto` (file/v1, common/v1) | Proto 定义 |
| `community-common/v2` (minio, errx, responsex) | MinIO 封装、错误码、响应格式 |
| `minio-go/v7` | 原始 MinIO SDK（上传预签名） |
| MySQL | file_db |
| MinIO | 对象存储 |
| etcd | 服务注册发现 |

**无出站 gRPC 调用** — 纯基础设施服务。

---

## 八、错误码

| 错误码 | 含义 |
|--------|------|
| 070001 | 文件不存在 |
| 070002 | 上传失败 |
| 070003 | 不支持的文件类型 |
| 070004 | 文件大小超限 |
| 070005 | Bucket 不存在 |
