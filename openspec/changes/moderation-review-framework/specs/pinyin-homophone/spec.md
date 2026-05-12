## ADDED Requirements

### Requirement: 汉字转拼音
系统 SHALL 将汉字转换为拼音序列，支持多音字（返回所有读音），基于 `mozillazg/go-pinyin` 实现。

#### Scenario: 单音字转换
- **WHEN** 对汉字 "你" 调用 ToPinyin
- **THEN** 返回 [["ni"]]

#### Scenario: 多音字转换
- **WHEN** 对汉字 "重" 调用 ToPinyin
- **THEN** 返回 [["zhong", "chong"]]

#### Scenario: 非汉字保持
- **WHEN** 对字符串 "hello" 调用 ToPinyin
- **THEN** 非汉字字符原样保留

### Requirement: 谐音变体生成
系统 SHALL 对给定的敏感词生成谐音变体列表，通过拼音匹配 + 同音字表替换实现。

#### Scenario: 两字词谐音
- **WHEN** 对 "傻逼" 调用 ExpandHomophones
- **THEN** 返回包含 "shabi" 拼音对应常见同音字的组合（如 "沙壁"、"杀币"等）

#### Scenario: 变体数量限制
- **WHEN** 对某个词调用 ExpandHomophones，可能的变体超过 20 个
- **THEN** 最多返回 20 个变体，按常见程度排序

#### Scenario: 空输入
- **WHEN** 对空字符串调用 ExpandHomophones
- **THEN** 返回空列表

### Requirement: 同音字表
系统 SHALL 内置常用汉字的同音字映射表，覆盖拼音到汉字的反向索引，用于谐音变体生成。

#### Scenario: 查询同音字
- **WHEN** 查询拼音 "sha" 的同音字
- **THEN** 返回包含 "沙"、"杀"、"傻" 等常见字的列表
