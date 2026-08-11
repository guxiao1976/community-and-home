// 全局类型补充 - 解决历史代码类型问题
// 这是临时方案，长期需要逐个文件修复类型

import type { ResidentialArea } from '@/api/residential-areas'
import type { HomeownerVerification } from '@/api/verification'
import type { User } from '@common/types/identity'
import type { Role } from '@/api/roles'
import type { SensitiveWord } from '@/api/sensitive-words'
import type { ModelConfig } from '@/api/aimodel'

declare module 'element-plus' {
  // 扩展 Element Plus 的 DefaultRow 类型，允许它兼容我们的业务类型
  interface DefaultRow
    extends Partial<ResidentialArea>,
      Partial<HomeownerVerification>,
      Partial<User>,
      Partial<Role>,
      Partial<SensitiveWord>,
      Partial<ModelConfig> {
    [key: string]: any
  }
}

export {}
