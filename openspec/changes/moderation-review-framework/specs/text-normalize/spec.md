## ADDED Requirements

### Requirement: 归一化链组合
系统 SHALL 支持将多个归一化策略组合为链式处理器，按顺序依次对每个字符执行归一化。

#### Scenario: 全链路归一化
- **WHEN** 文本 "ＨｅｌｌｏＷｏｒｌｄ" 经过 IgnoreWidth + IgnoreCase 链
- **THEN** 归一化结果为 "helloworld"

#### Scenario: 空链
- **WHEN** 创建不包含任何策略的归一化器
- **THEN** 文本原样返回

### Requirement: 全角半角转换
系统 SHALL 将全角英文字母、数字转换为半角，转换映射基于 Unicode 范围（0xFF01-0xFF5E → 0x0021-0x007E）。

#### Scenario: 全角字母
- **WHEN** 字符 'Ａ' 经过 IgnoreWidth
- **THEN** 转换为 'A'

#### Scenario: 全角数字
- **WHEN** 字符 '１' 经过 IgnoreWidth
- **THEN** 转换为 '1'

#### Scenario: 已是半角
- **WHEN** 字符 'A' 经过 IgnoreWidth
- **THEN** 保持 'A' 不变

### Requirement: 大小写转换
系统 SHALL 将所有大写英文字母转换为小写。

#### Scenario: 大写转小写
- **WHEN** 字符 'F' 经过 IgnoreCase
- **THEN** 转换为 'f'

### Requirement: 繁简转换
系统 SHALL 内置常见繁简映射表（3500+ 汉字），将繁体汉字转换为简体。无外部依赖。

#### Scenario: 繁体转简体
- **WHEN** 字符 '龜' 经过 IgnoreChinese
- **THEN** 转换为 '龟'

#### Scenario: 简体保持
- **WHEN** 字符 '龟' 经过 IgnoreChinese
- **THEN** 保持 '龟' 不变

#### Scenario: 非汉字保持
- **WHEN** 字符 'A' 经过 IgnoreChinese
- **THEN** 保持 'A' 不变

### Requirement: 数字格式归一化
系统 SHALL 将各种 Unicode 数字变体（①②③、壹贰叁、全角１２３等）转换为标准阿拉伯数字。

#### Scenario: 带圈数字
- **WHEN** 字符 '①' 经过 IgnoreNumStyle
- **THEN** 转换为 '1'

#### Scenario: 中文数字
- **WHEN** 字符 '壹' 经过 IgnoreNumStyle
- **THEN** 转换为 '1'

### Requirement: 英文特殊字符归一化
系统 SHALL 将特殊 Unicode 英文字符（ⒶⒷⓐⓑ等）转换为标准 ASCII 字母。

#### Scenario: 特殊字符转换
- **WHEN** 字符 'Ⓕ' 经过 IgnoreEnglishStyle
- **THEN** 转换为 'f'

### Requirement: 重复字符压缩
系统 SHALL 将连续重复的相同字符压缩为单个字符，仅对连续 3 个及以上重复生效。

#### Scenario: 三连重复
- **WHEN** 文本 "你你你好" 经过 IgnoreRepeat
- **THEN** 压缩为 "你好"

#### Scenario: 不压缩短重复
- **WHEN** 文本 "你好" 经过 IgnoreRepeat
- **THEN** 保持 "你好" 不变

### Requirement: 位置映射
归一化 SHALL 同时返回原文位置到归一化文本位置的映射，支持将匹配结果映射回原文位置。

#### Scenario: 映射正确性
- **WHEN** 文本 "ＡＢＣ" 归一化为 "abc"
- **THEN** 位置映射能将归一化文本的位置 0,1,2 正确映射回原文位置 0,1,2
