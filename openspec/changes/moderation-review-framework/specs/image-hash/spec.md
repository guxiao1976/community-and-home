## ADDED Requirements

### Requirement: 感知哈希计算
系统 SHALL 实现图片感知哈希（pHash）计算接口，接受 `image.Image` 输入，返回 `uint64` 哈希值。

#### Scenario: 计算图片哈希
- **WHEN** 传入一张有效的 PNG/JPEG 图片
- **THEN** 返回一个 uint64 感知哈希值

### Requirement: 哈希距离比较
系统 SHALL 计算两个感知哈希之间的汉明距离（Hamming Distance），用于判断图片相似度。

#### Scenario: 相同图片
- **WHEN** 对同一张图片计算两次哈希并比较
- **THEN** 汉明距离为 0

#### Scenario: 相似图片
- **WHEN** 两张图片的汉明距离 <= 配置的阈值（默认 10）
- **THEN** 判定为相似图片

### Requirement: 违规图库管理
系统 SHALL 支持加载已知违规图片的哈希集合，用于快速比对。

#### Scenario: 加载违规图库
- **WHEN** 从数据库加载违规图片哈希列表
- **THEN** 后续审核图片可与之比对

#### Scenario: 命中违规图库
- **WHEN** 待审图片的哈希与违规图库中某个哈希的距离 <= 阈值
- **THEN** 返回匹配结果，标记为违规

### Requirement: 预留接口
当前阶段图像哈希模块为预留接口，基础结构 SHALL 完整但依赖 `goimagehash` 的实际调用可延后实现。

#### Scenario: 接口可用
- **WHEN** 引擎层调用 `ImageHasher.Hash()`
- **THEN** 接口存在且可编译，实际功能待后续填充
