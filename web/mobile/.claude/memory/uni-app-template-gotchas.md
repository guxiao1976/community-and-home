---
name: uni-app-template-gotchas
description: Uni-app 模板编译的两个致命坑——嵌套text和内联赋值
metadata:
  type: reference
---

# Uni-app 模板编译坑

## 坑1：嵌套 `<text>` 元素

**症状**：`ReferenceError: temp0 is not defined`

**原因**：Uni-app 编译器不允许 `<text>` 里面嵌套 `<text>`。

**错误写法**：
```html
<text class="outer">已阅并同意<text class="link">《使用协议》</text></text>
```

**正确写法**：外层用 `<view>`：
```html
<view class="outer">已阅并同意<text class="link">《使用协议》</text></view>
```

## 坑2：内联表达式 `@click="x = ''"`

**症状**：同上 `temp0 is not defined`

**原因**：Uni-app 模板编译器对 `@click="phone = ''"` 这类内联赋值支持不好。

**正确写法**：写一个方法：
```typescript
function clearPhone() { phone.value = ''; }
```
```html
@click="clearPhone"
```

## 坑3：构建缓存污染

**症状**：代码明明正确但编译报错或运行异常

**修复**：清缓存重建
```bash
rm -rf dist .uni unpackage && npm run build:h5
```

**Why:** 多次修改代码后，Vite/Uni-app 缓存可能残留旧模板编译产物，导致 `temp0` 等内部变量引用错误。

**How to apply:** 遇到莫名其妙的编译/运行时错误，第一步清缓存。
