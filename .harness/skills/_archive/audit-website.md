# Audit Website 技能使用模板

## 技能名称
`audit-website`

## 功能描述
全面审计前端应用质量，包括性能、可访问性、SEO 和最佳实践，基于 Google Lighthouse。

## 使用场景
- 功能开发完成，准备上线前
- 性能优化需要基线数据
- 确保符合 WCAG 可访问性标准
- SEO 优化评估
- 定期质量检查

## 输入要求

### 必需信息
- 应用 URL（开发环境或生产环境）
- 审计维度（性能、可访问性、SEO、最佳实践）

### 可选信息
- 特定页面列表（首页、登录页、核心功能页）
- 设备类型（桌面、移动端）
- 网络条件模拟（3G、4G、WiFi）
- 自定义阈值（性能分数 > 90）

## 输出内容
1. Lighthouse 审计报告（JSON + HTML）
2. 问题清单和优先级
3. 优化建议和最佳实践
4. 性能指标对比（如有历史数据）

## 调用示例

### 示例 1: 全面质量审计
```
审计当前前端应用的整体质量：

目标 URL: http://localhost:3003
审计页面：
- 首页 (/)
- 登录页 (/login)
- 用户列表 (/users)
- 仪表盘 (/dashboard)

审计维度：
- 性能 (Performance)
- 可访问性 (Accessibility)
- SEO
- 最佳实践 (Best Practices)

设备：桌面 + 移动端
网络：Fast 3G

输出：
- 完整的 Lighthouse 报告
- 按优先级排序的问题清单
- 每个问题的修复建议
```

### 示例 2: 性能专项审计
```
专注性能优化审计：

目标 URL: http://localhost:3003/dashboard
审计焦点：性能指标

关注指标：
- FCP (First Contentful Paint) - 首次内容绘制
- LCP (Largest Contentful Paint) - 最大内容绘制
- CLS (Cumulative Layout Shift) - 累计布局偏移
- FID (First Input Delay) - 首次输入延迟
- TTI (Time to Interactive) - 可交互时间

期望目标：
- FCP < 1.8s
- LCP < 2.5s
- CLS < 0.1
- FID < 100ms
- Performance Score > 90

输出：
- 性能瓶颈分析
- 资源加载优化建议
- 代码分割建议
```

### 示例 3: 可访问性合规审计
```
确保应用符合 WCAG 2.1 AA 标准：

目标 URL: http://localhost:3003
审计焦点：可访问性

检查项目：
- 键盘导航（Tab 键顺序、焦点管理）
- 屏幕阅读器兼容性（ARIA 标签、alt 文本）
- 颜色对比度（文本/背景对比度 ≥ 4.5:1）
- 表单可访问性（label 关联、错误提示）
- 语义化 HTML（正确使用 header、nav、main 等）

合规级别：WCAG 2.1 AA

输出：
- 不合规项目清单
- 修复优先级排序
- 代码修复示例
```

## 审计维度详解

### 1. 性能 (Performance)
**关键指标**:
- **FCP** - 首次内容绘制 (目标: < 1.8s)
- **LCP** - 最大内容绘制 (目标: < 2.5s)
- **CLS** - 累计布局偏移 (目标: < 0.1)
- **FID** - 首次输入延迟 (目标: < 100ms)
- **Speed Index** - 速度指数 (目标: < 3.4s)
- **TTI** - 可交互时间 (目标: < 3.8s)

**常见问题**:
- 未压缩的图片
- 未使用的 JavaScript
- 阻塞渲染的资源
- 未使用 CDN
- 缺少缓存策略

### 2. 可访问性 (Accessibility)
**WCAG 2.1 检查项**:
- 键盘可访问性
- ARIA 属性正确使用
- 颜色对比度充足
- 表单元素有 label
- 图片有 alt 文本
- 语义化 HTML 结构
- 焦点指示器可见

**常见问题**:
- 缺少 alt 属性
- 颜色对比度不足
- 无键盘焦点指示
- ARIA 属性滥用
- 非语义化标签

### 3. SEO
**检查项**:
- Meta 标签完整性（title, description）
- 结构化数据（Schema.org）
- 移动端友好性
- robots.txt 和 sitemap.xml
- 语义化 HTML
- 页面加载速度
- HTTPS 使用

**常见问题**:
- 缺少 meta description
- title 标签重复
- 图片无 alt 文本
- 链接无描述性文本
- 移动端不友好

### 4. 最佳实践 (Best Practices)
**检查项**:
- HTTPS 使用
- 控制台无错误
- 安全的第三方库
- 现代 JavaScript API
- 避免已废弃的 API
- 正确的 HTTP 状态码
- CSP (Content Security Policy)

**常见问题**:
- HTTP 未升级到 HTTPS
- 控制台有错误或警告
- 使用过时的库版本
- 缺少安全头部

## 最佳实践

### 1. 定期审计
建议频率：
- 开发阶段：每次大功能完成后
- 测试阶段：发版前必审
- 生产环境：每月一次

### 2. 设置基线
首次审计后设定目标：
```
性能目标：
- Performance Score > 90
- LCP < 2.5s
- CLS < 0.1

可访问性目标：
- Accessibility Score > 95
- 0 个 WCAG 违规项

SEO 目标：
- SEO Score > 90
- 所有页面有 meta 标签
```

### 3. 持续监控
使用 CI/CD 集成：
```bash
# 在 CI 中运行 Lighthouse
npm run lighthouse -- --url=http://staging.example.com
```

### 4. 优先级排序
按影响程度排序：
1. **Critical** - 阻止用户使用（可访问性严重问题）
2. **High** - 显著影响体验（性能问题）
3. **Medium** - 影响部分用户（SEO 问题）
4. **Low** - 优化建议（最佳实践）

## 与其他技能配合

### audit-website + frontend-design
```
1. frontend-design 生成 UI
2. audit-website 审计质量
3. 根据审计结果调整设计
4. 重新审计验证
```

### audit-website + webapp-testing
```
1. webapp-testing 验证功能
2. audit-website 审计质量
3. 综合优化性能和体验
```

## 优化建议示例

### 性能优化
```typescript
// ❌ 未优化
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'

// ✅ 按需导入
import { ElButton, ElTable } from 'element-plus'
```

### 可访问性优化
```vue
<!-- ❌ 缺少 alt 和 label -->
<img src="/avatar.jpg">
<input type="text">

<!-- ✅ 完整的可访问性属性 -->
<img src="/avatar.jpg" alt="用户头像">
<label for="username">用户名</label>
<input id="username" type="text" aria-label="请输入用户名">
```

### SEO 优化
```html
<!-- ❌ 缺少 meta -->
<head>
  <title>管理后台</title>
</head>

<!-- ✅ 完整的 meta -->
<head>
  <title>用户管理 - 社区管理系统</title>
  <meta name="description" content="管理社区用户，查看用户信息，编辑用户权限">
  <meta name="viewport" content="width=device-width, initial-scale=1">
</head>
```

## 注意事项

1. **审计环境**
   - 使用与生产环境相似的配置
   - 确保数据已加载（避免空页面）
   - 关闭浏览器扩展

2. **结果解读**
   - 分数只是参考，关注具体问题
   - 某些问题可能不适用（如 SEO 对管理后台）
   - 优先修复影响用户的问题

3. **性能测试**
   - 多次测试取平均值
   - 模拟真实网络条件
   - 考虑服务器响应时间

4. **可访问性**
   - 自动化工具只能检测部分问题
   - 需要手动测试键盘导航
   - 最好找真实用户测试

## 常见问题

### Q: Lighthouse 分数波动很大怎么办？
A: 多次运行取平均值，或使用 Lighthouse CI 获得更稳定的结果。

### Q: 如何提升 Performance 分数？
A: 关注 FCP 和 LCP，优化图片、代码分割、使用 CDN、启用缓存。

### Q: 可访问性分数低怎么办？
A: 重点修复 contrast、alt、label 等基础问题，再处理复杂的 ARIA。

### Q: 管理后台需要 SEO 吗？
A: 一般不需要，可以忽略 SEO 分数，但良好的 meta 标签有助于书签和分享。

## 报告示例

```
=== Lighthouse 审计报告 ===
URL: http://localhost:3003/dashboard
日期: 2026-06-21

【总体得分】
✅ Performance: 92/100
⚠️ Accessibility: 85/100
✅ Best Practices: 95/100
⚠️ SEO: 78/100

【关键指标】
✅ FCP: 1.2s (目标: < 1.8s)
✅ LCP: 2.1s (目标: < 2.5s)
❌ CLS: 0.15 (目标: < 0.1)
✅ FID: 45ms (目标: < 100ms)

【需要修复】(按优先级)
1. [High] 颜色对比度不足 (Accessibility)
   - 影响: 3 个按钮
   - 建议: 增加对比度至 4.5:1
   
2. [High] 累计布局偏移过大 (Performance)
   - 原因: 图片未指定尺寸
   - 建议: 为 img 添加 width/height

3. [Medium] 缺少 meta description (SEO)
   - 建议: 为每个页面添加描述性 meta
```

## 相关文档
- [Lighthouse 文档](https://developer.chrome.com/docs/lighthouse/)
- [Web Vitals](https://web.dev/vitals/)
- [WCAG 2.1 指南](https://www.w3.org/WAI/WCAG21/quickref/)
- [Chrome DevTools Performance](https://developer.chrome.com/docs/devtools/performance/)
