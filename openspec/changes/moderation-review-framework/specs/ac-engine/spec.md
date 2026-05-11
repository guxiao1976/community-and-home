## ADDED Requirements

### Requirement: AC 自动机构建
系统 SHALL 从敏感词列表构建 AC 自动机（Aho-Corasick），支持通过 `Build(entries []WordEntry)` 方法一次性构建，支持通过 `Rebuild()` 原子替换旧机器。

#### Scenario: 正常构建
- **WHEN** 传入包含 ["敏感词", "政治", "色情"] 的 WordEntry 列表
- **THEN** AC 自动机构建完成，可对文本进行匹配

#### Scenario: 空词库构建
- **WHEN** 传入空列表
- **THEN** AC 自动机正常构建，后续匹配返回空结果

### Requirement: 单趟匹配
系统 SHALL 对输入文本进行单趟 O(n) 扫描，返回所有命中的敏感词及其位置。

#### Scenario: 单个词命中
- **WHEN** 文本 "这是一个敏感内容" 匹配包含 "敏感" 的 AC 自动机
- **THEN** 返回 MatchResult{Word:"敏感", Start:3, End:5, Category:"...", Severity:1}

#### Scenario: 多个词命中
- **WHEN** 文本 "政治和色情内容" 匹配包含 "政治" 和 "色情" 的 AC 自动机
- **THEN** 返回两个 MatchResult，按文本位置排序

#### Scenario: 无命中
- **WHEN** 文本 "今天天气真好" 匹配包含 "敏感词" 的 AC 自动机
- **THEN** 返回空结果

### Requirement: 并发安全
AC 自动机 SHALL 使用读写锁保护，构建/重建时加写锁，匹配时加读锁，支持并发匹配。

#### Scenario: 并发读写
- **WHEN** 一个 goroutine 执行 Rebuild，多个 goroutine 执行 Match
- **THEN** 不会产生数据竞争，匹配始终使用完整的机器快照

### Requirement: 拼音变体构建
词库构建时 SHALL 对 severity=1 的高风险词自动生成拼音谐音变体，每词最多 20 个变体，变体词与原词共享 Category 和 Severity。

#### Scenario: 高风险词谐音扩展
- **WHEN** 构建词库，其中 severity=1 的词为 "傻逼"
- **THEN** AC 自动机中同时包含原词和谐音变体（如 "沙壁"、"杀币"）

#### Scenario: 低风险词不扩展
- **WHEN** 构建词库，其中 severity=2 的词为 "笨蛋"
- **THEN** AC 自动机中仅包含原词 "笨蛋"，不生成谐音变体
