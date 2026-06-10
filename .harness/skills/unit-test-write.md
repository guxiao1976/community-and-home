# unit-test-write

## 触发条件

为 Go 服务代码编写单元测试。触发词：`写测试`、`补测试`、`单元测试`、`go test`、`测试覆盖`。

## 核心原则：改动驱动测试

改了哪个接口就测哪个接口，不对最上层做一刀切测试。新增/修改的函数必须有对应测试，未改动的代码不补测试（避免范围蔓延）。

## 执行步骤

### Step 1: 确定测试范围

```bash
git diff --name-only HEAD      # 哪些文件变了
git diff HEAD -- <file>         # 具体变了哪些函数
```

只测变更的函数。如果变更涉及多个包，每个包独立写 `_test.go`。

### Step 2: 设计测试用例

按 Go table-driven tests 模式，每个函数覆盖：

| 场景 | 说明 |
|------|------|
| **正常路径** | 典型输入 → 预期输出，至少 1 条 |
| **边界条件** | 空值、零值、最大/最小值、空切片 |
| **错误路径** | 无效输入 → 预期错误，验证错误码 |
| **业务约束** | 状态机边界（如 verf_status 不允许的操作） |

有依赖的用 `gomock` 或手写 stub interface 隔离。

### Step 3: 构造测试数据

优先使用**真实数据**：
- 查 `docs/design.md` 了解业务规则和状态转换
- 查现有代码中的常量定义（错误码、状态枚举）
- 如对接了线上环境，参考真实请求出入参

禁止：随机生成没有业务含义的数据、复制已有测试不改断言。

### Step 4: 写测试

```go
func TestXxx(t *testing.T) {
    tests := []struct {
        name    string
        input   InputType
        want    OutputType
        wantErr bool
    }{
        {name: "正常路径", input: ..., want: ..., wantErr: false},
        {name: "边界-空值", input: ..., want: ..., wantErr: false},
        {name: "错误-无效参数", input: ..., want: ..., wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := FuncUnderTest(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
                return
            }
            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

### Step 5: 验证

```bash
go test ./... -count=1 -cover          # 全跑一遍
go test ./... -count=1 -coverprofile=coverage.out  # 覆盖率
go tool cover -func=coverage.out | grep <changed_file>  # 只看变更文件
```

验证标准：
- 所有测试 PASS
- 变更文件覆盖率 > 70%
- 核心逻辑覆盖率 > 80%

## 禁止

- 只写"调用不报错"的测试——必须 assert 返回值
- 复制已有测试改个名——每条测试有独立业务价值
- 把测试写在上层 Handler 测——测 Logic 层函数
- 为未变更的代码补测试——范围蔓延

## 关联

- 编码规范：`.harness/rules/项目编码规范.md`
- QA 检查：`.harness/skills/qa.md`
- 设计文档：`services/<name>/docs/design.md`（了解业务规则）
