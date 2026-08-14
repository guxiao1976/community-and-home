# PDF 技能使用模板

## 技能名称
`pdf`

## 功能描述
将 Markdown、文本或结构化数据转换为专业的 PDF 文档，支持丰富的排版和格式化选项。

## 使用场景
- 生成技术文档交付物
- 导出架构设计文档
- 生成测试报告
- 创建 Release Notes
- 生成会议纪要或评审报告

## 输入要求

### 必需信息
- 文档内容（Markdown 或纯文本）
- 文档标题
- 输出文件名

### 可选信息
- 文档作者
- 版本号
- 生成日期
- 封面设计
- 页眉页脚
- 目录生成
- 主题样式

## 输出内容
1. 专业排版的 PDF 文件
2. 可选的书签导航
3. 可选的目录索引
4. 可选的页码

## 调用示例

### 示例 1: 交付文档生成
```
生成 permission-service 的交付文档：

## 文档内容

将以下内容转换为专业的 PDF 交付文档：

### 封面信息
- 服务名称: 权限管理服务 (Permission Service)
- 版本: v1.0.0
- 交付日期: 2026-06-21
- 项目: Community-and-Home

### 第一章: 变更概览

本次交付实现了基于 RBAC 的用户权限管理系统，包括：

**核心功能**
- 角色管理（创建、编辑、删除、列表）
- 权限配置（为角色分配权限，树形结构）
- 用户角色关联（为用户分配多个角色）
- 权限检查中间件（API 层和 gRPC 层）

**技术亮点**
- 使用 Casbin 作为权限引擎
- 权限数据缓存到 Redis (TTL 5分钟)
- 前端树形控件支持全选/反选/搜索

**影响范围**
- 新增 8 个数据表
- 新增 15 个 gRPC 接口
- 新增 12 个 REST API
- 新增 3 个前端页面

### 第二章: 架构设计

#### 系统架构

```
┌─────────────┐
│  前端界面    │
└──────┬──────┘
       │ HTTP
┌──────▼──────┐
│  API 层     │
└──────┬──────┘
       │ gRPC
┌──────▼──────┐
│  RPC 层     │
│  (Casbin)   │
└──────┬──────┘
       │
┌──────▼──────┐
│  MySQL      │
│  Redis      │
└─────────────┘
```

#### 数据模型

**角色表 (rbac_roles)**
- id: bigint (主键)
- name: varchar(50) (角色名)
- description: varchar(200)
- created_at, updated_at

**权限表 (rbac_permissions)**
- id: bigint (主键)
- resource: varchar(100) (资源)
- action: varchar(50) (操作)
- description: varchar(200)

...（省略更多内容）

### 第三章: API 文档

#### 3.1 角色管理 API

**创建角色**
```
POST /api/permission/roles
Content-Type: application/json

请求:
{
  "name": "管理员",
  "description": "系统管理员角色"
}

响应:
{
  "code": 0,
  "msg": "success",
  "data": {
    "id": "1234567890123456789",
    "name": "管理员",
    "description": "系统管理员角色"
  }
}
```

...（省略更多 API）

### 第四章: 测试报告

#### 测试覆盖率
- 单元测试: 28 个用例，覆盖率 85%
- E2E 测试: 8 个场景，全部通过

#### 性能测试
- 权限检查响应时间: < 50ms (P95)
- Redis 缓存命中率: 96.5%
- 并发支持: 1000+ QPS

### 第五章: 部署指南

#### 前置条件
- MySQL 8.0+
- Redis 7+
- etcd v3.5+

#### 部署步骤
1. 数据库迁移
2. 部署 RPC 服务
3. 部署 API 服务
4. 部署前端
5. 验证功能

...（省略详细步骤）

## PDF 格式要求
- 页面尺寸: A4
- 字体: 标题用黑体，正文用宋体
- 代码块: 等宽字体，浅灰背景
- 目录: 自动生成，带页码
- 页眉: 服务名称
- 页脚: 页码
- 书签: 按章节层级
- 封面: 专业设计，包含 Logo

请使用 /pdf 技能生成专业的交付文档。
```

**期望输出**: 
- 文件名: `permission-service-v1.0.0-delivery.pdf`
- 页数: 约 30-50 页
- 包含完整的目录、书签、页眉页脚

### 示例 2: 架构设计文档导出
```
将架构设计 Markdown 转换为 PDF：

## 源文件
docs/specs/permission-design.md

## PDF 要求
- 保留原有 Markdown 格式
- 代码块语法高亮
- 图表清晰可读
- 添加封面和目录
- 页眉显示文档标题
- 页脚显示页码和日期

输出文件: docs/specs/permission-design.pdf
```

### 示例 3: Release Notes 生成
```
生成 v2.0.0 的 Release Notes PDF：

## 版本信息
- 版本: v2.0.0
- 发布日期: 2026-07-01
- 上一版本: v1.5.3

## 主要变更

### 新功能
- ✨ 支持多租户隔离
- ✨ 添加审计日志功能
- ✨ 支持自定义权限规则

### 改进
- ⚡ 权限检查性能提升 50%
- 📈 Redis 缓存策略优化
- 🎨 前端界面改版

### 修复
- 🐛 修复角色删除时的级联问题
- 🐛 修复并发修改权限配置的竞态
- 🐛 修复缓存失效延迟问题

### 破坏性变更
- ⚠️ API `/api/permission/roles` 响应格式变更
- ⚠️ 配置文件新增 `TenantMode` 字段

### 升级指南
1. 备份数据库
2. 执行迁移脚本
3. 更新配置文件
4. 重启服务

## PDF 样式
- 使用明亮配色
- Emoji 图标保留
- 变更按类型分组
- 突出破坏性变更

输出: releases/v2.0.0-release-notes.pdf
```

## 最佳实践

### 1. 内容结构化
```markdown
# 一级标题（章节）
## 二级标题（小节）
### 三级标题（要点）

- 列表项
- 列表项

**加粗重点**
*斜体强调*
`代码`
```

### 2. 代码块格式化
````markdown
```go
// 使用语言标识符，启用语法高亮
func CheckPermission(user, resource, action string) bool {
    return enforcer.Enforce(user, resource, action)
}
```
````

### 3. 表格使用
```markdown
| 列1 | 列2 | 列3 |
|-----|-----|-----|
| 值1 | 值2 | 值3 |
```

### 4. 图片嵌入
```markdown
![架构图](./architecture.png)
```

### 5. 链接处理
```markdown
[内部章节](#第二章)
[外部链接](https://example.com)
```

## PDF 样式选项

### 主题
- **专业**: 蓝灰配色，适合技术文档
- **明亮**: 鲜艳配色，适合展示材料
- **简洁**: 黑白配色，适合打印

### 字体
- **中文**: 宋体、黑体、楷体
- **英文**: Times New Roman, Arial, Courier
- **代码**: Consolas, Monaco, Courier New

### 页面布局
- **边距**: 上下左右各 2.5cm
- **页眉高度**: 1.5cm
- **页脚高度**: 1.5cm

## 与其他技能配合

### pdf + writing-plans
```
1. writing-plans 生成任务列表
2. pdf 导出为 PDF 格式
3. 分发给团队成员
```

### pdf + requesting-code-review
```
1. requesting-code-review 生成 PR 描述
2. pdf 导出为归档文档
3. 保存到文档库
```

### pdf + xlsx
```
1. xlsx 生成测试报告表格
2. pdf 生成完整交付文档
3. 在 PDF 中引用 Excel 报告
```

## 注意事项

### 1. 文件大小
- 避免过大的图片（建议 < 1MB）
- 压缩 PNG 图片
- 使用适当的图片分辨率（300 DPI 打印，72 DPI 屏幕）

### 2. 字体兼容性
- 使用常见字体，避免特殊字体
- 中文内容使用标准中文字体

### 3. 链接处理
- 内部链接（锚点）在 PDF 中保持可用
- 外部链接在 PDF 中可点击

### 4. 分页控制
```markdown
<!-- 强制分页 -->
<div style="page-break-after: always;"></div>
```

### 5. 目录生成
- 自动从标题层级生成
- 最多支持 3 级目录
- 带页码跳转

## 常见问题

### Q: 如何控制 PDF 的页数？
A: 调整内容详细程度，使用简洁的表述，避免大量截图。

### Q: 代码块在 PDF 中不够清晰怎么办？
A: 使用等宽字体，增加字号，设置浅色背景。

### Q: 如何处理超长表格？
A: 表格跨页自动处理，或拆分为多个小表格。

### Q: PDF 文件过大怎么办？
A: 压缩图片，减少高分辨率图片，使用矢量图（SVG）。

### Q: 如何添加水印？
A: 在 prompt 中指定水印内容和位置。

## 输出格式说明

### 文件命名
```
{服务名}-{文档类型}-{版本}.pdf

示例:
- permission-service-delivery-v1.0.0.pdf
- architecture-design-v2.0.pdf
- release-notes-v2.0.0.pdf
```

### 元数据
PDF 包含以下元数据：
- 标题 (Title)
- 作者 (Author)
- 主题 (Subject)
- 关键词 (Keywords)
- 创建日期 (Creation Date)

### 书签结构
```
第一章: 变更概览
├─ 1.1 核心功能
├─ 1.2 技术亮点
└─ 1.3 影响范围

第二章: 架构设计
├─ 2.1 系统架构
└─ 2.2 数据模型
```

## 示例模板

### 技术文档模板
```markdown
# {服务名称} 技术文档

**版本**: {版本号}  
**日期**: {日期}  
**作者**: {作者}

---

## 目录
<!-- 自动生成 -->

---

## 第一章: 概述
...

## 第二章: 架构设计
...

## 第三章: API 文档
...

## 附录
...
```

## 相关文档
- [Markdown 语法指南](https://www.markdownguide.org/)
- [PDF 最佳实践](https://www.adobe.com/acrobat/resources.html)
- [技术文档写作规范](https://developers.google.com/style)
