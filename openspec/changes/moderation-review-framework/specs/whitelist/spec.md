## ADDED Requirements

### Requirement: 白名单 AC 自动机
系统 SHALL 维护独立的白名单 AC 自动机，白名单词从 `md_sensitive_word` 表中 `word_type=2` 的记录加载。

#### Scenario: 加载白名单
- **WHEN** 数据库中存在 word_type=2 的记录 ["五星红旗", "天安门广场"]
- **THEN** 白名单 AC 自动机构建完成，包含这两个词

### Requirement: 白名单优先匹配
文本审核时 SHALL 同时查询黑名单和白名单，当白名单匹配长度 >= 黑名单匹配长度时，跳过该段（不标记为敏感词）。

#### Scenario: 长白黑短
- **WHEN** 文本 "我爱五星红旗"，黑名单包含 "五星"，白名单包含 "五星红旗"
- **THEN** 白名单匹配长度 4 > 黑名单匹配长度 2，该段被跳过

#### Scenario: 仅有黑名单
- **WHEN** 文本 "敏感内容"，黑名单包含 "敏感"，白名单无匹配
- **THEN** 正常标记为敏感词命中

#### Scenario: 等长匹配
- **WHEN** 白名单和黑名单匹配长度相同
- **THEN** 白名单优先，跳过该段

### Requirement: LongestMatch 查询
系统 SHALL 提供 `LongestMatch(text)` 方法返回最长白名单匹配词及其长度。

#### Scenario: 最长匹配
- **WHEN** 文本 "五星红旗飘扬"，白名单包含 "五星" 和 "五星红旗"
- **THEN** 返回 word="五星红旗", length=4
