## ADDED Requirements

### Requirement: 干扰字符检测
系统 SHALL 识别文本中的干扰字符（分隔符），包括 `x X . 。 * 、 空格 - _ / \ | ~ · 丶` 及 Emoji。

#### Scenario: 识别常见干扰
- **WHEN** 文本为 "敏x感x词"
- **THEN** 系统识别 'x' 为干扰字符

#### Scenario: 识别多类型干扰
- **WHEN** 文本为 "敏.感.词" 或 "敏 感 词"
- **THEN** 系统识别 '.' 和 ' ' 为干扰字符

### Requirement: 拆字还原匹配
系统 SHALL 去除干扰字符后将剩余片段拼接，用 AC 自动机匹配还原出的连续文本。

#### Scenario: 单字符干扰
- **WHEN** 文本 "敏x感x词"，去除干扰后为 "敏感词"
- **THEN** AC 自动机匹配到 "敏感词"，返回原文位置映射

#### Scenario: 多字符干扰
- **WHEN** 文本 "敏..感...词"
- **THEN** 去除干扰后拼接为 "敏感词"，匹配成功

#### Scenario: 正常文本不受影响
- **WHEN** 文本 "这是一个正常句子"
- **THEN** 拆字检测不产生误匹配

### Requirement: 配置化干扰字符集
干扰字符集 SHALL 通过配置文件可配置，默认包含 `xX.* 、-_|/~·丶`。

#### Scenario: 自定义干扰字符
- **WHEN** 配置中 SplitSeparators 为 "xX."
- **THEN** 系统仅识别 'x'、'X'、'.' 为干扰字符
