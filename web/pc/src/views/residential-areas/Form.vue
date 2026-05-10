<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createResidentialArea, updateResidentialArea, getResidentialAreaById, getAdministrativeDivisions } from '@/api/masterdata'
import type { AdministrativeDivision } from '@common/types/masterdata'

const router = useRouter()
const route = useRoute()

const isEdit = computed(() => !!route.params.id)
const areaId = computed(() => Number(route.params.id))

const formRef = ref<FormInstance>()
const loading = ref(false)
const submitting = ref(false)

// Division cascader data (only for edit mode)
const divisionOptions = ref<any[]>([])

const formData = reactive({
  county_id: undefined as number | undefined,
  street_id: undefined as number | undefined,
  community_div_id: undefined as number | undefined,
  name: '',
  address: '',
  community_type: 1
})

// Read-only division path from route query (create mode)
const divisionPath = computed(() => {
  const parts: string[] = []
  const q = route.query
  if (q.county_name) parts.push(String(q.county_name))
  if (q.street_name) parts.push(String(q.street_name))
  if (q.community_name) parts.push(String(q.community_name))
  return parts.join(' > ')
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入小区名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  community_type: [{ required: true, message: '请选择小区类型', trigger: 'change' }]
}

// Load divisions (only for edit mode)
const loadDivisions = async () => {
  if (!isEdit.value) return
  try {
    const res = await getAdministrativeDivisions({ page: 1, page_size: 5000 })
    divisionOptions.value = buildDivisionTree(res.list || [])
  } catch (error: any) {
    ElMessage.error(error.message || '加载行政区划失败')
  }
}

const buildDivisionTree = (list: AdministrativeDivision[]): any[] => {
  const map = new Map<number, any>()
  const roots: any[] = []

  list.forEach(d => {
    map.set(d.id, { ...d, children: [] })
  })

  map.forEach(d => {
    if (d.parent_id && map.has(d.parent_id)) {
      map.get(d.parent_id).children.push(d)
    } else if (d.level === 3) {
      roots.push(d)
    }
  })

  return roots
}

// Cascader change handler (edit mode only)
const handleDivisionChange = (value: number[]) => {
  formData.county_id = value[0] || undefined
  formData.street_id = value[1] || undefined
  formData.community_div_id = value[2] || undefined
}

// Load data for edit
const loadArea = async () => {
  if (!isEdit.value) return
  loading.value = true
  try {
    const res = await getResidentialAreaById(areaId.value)
    const area = res.residential_area
    formData.county_id = area.county_id || undefined
    formData.street_id = area.street_id || undefined
    formData.community_div_id = area.community_div_id || undefined
    formData.name = area.name
    formData.address = area.address || ''
    formData.community_type = area.community_type
  } catch (error: any) {
    ElMessage.error(error.message || '加载小区信息失败')
    router.back()
  } finally {
    loading.value = false
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  try {
    await formRef.value.validate()
    submitting.value = true

    if (isEdit.value) {
      await updateResidentialArea(areaId.value, {
        street_id: formData.street_id,
        community_div_id: formData.community_div_id,
        code: undefined as any,
        name: formData.name,
        address: formData.address,
        area: undefined as any,
        population: undefined as any,
        community_type: formData.community_type
      })
      ElMessage.success('更新成功')
    } else {
      const payload: any = {
        county_id: formData.county_id!,
        name: formData.name,
        community_type: formData.community_type
      }
      if (formData.street_id) payload.street_id = formData.street_id
      if (formData.community_div_id) payload.community_div_id = formData.community_div_id
      if (formData.address) payload.address = formData.address
      await createResidentialArea(payload)
      ElMessage.success('创建成功')
    }
    router.push('/masterdata/residential-areas')
  } catch (error: any) {
    if (error !== false) {
      ElMessage.error(error.message || '操作失败')
    }
  } finally {
    submitting.value = false
  }
}

const handleCancel = () => {
  router.back()
}

onMounted(() => {
  if (isEdit.value) {
    loadDivisions()
    loadArea()
  } else {
    // Create mode: read from route query
    const q = route.query
    formData.county_id = q.county_id ? Number(q.county_id) : undefined
    formData.street_id = q.street_id ? Number(q.street_id) : undefined
    formData.community_div_id = q.community_div_id ? Number(q.community_div_id) : undefined
    // Auto-set community type based on whether community is selected
    formData.community_type = formData.community_div_id ? 1 : 2
    // Auto-fill name: if community name ends with "村委会", use "xx村" as default name
    const communityName = String(q.community_name || '')
    if (communityName.endsWith('村委会')) {
      formData.name = communityName.slice(0, -2)
    } else if (communityName) {
      formData.name = ''
    }
  }
})
</script>

<template>
  <div class="residential-area-form">
    <el-card v-loading="loading">
      <template #header>
        <div class="card-header">{{ isEdit ? '编辑小区' : '新建小区' }}</div>
      </template>

      <el-form ref="formRef" :model="formData" :rules="rules" label-width="120px" style="max-width: 800px">
        <!-- Edit mode: cascader selector -->
        <el-form-item v-if="isEdit" label="所属区县" prop="county_id">
          <el-cascader
            :model-value="[formData.county_id, formData.street_id, formData.community_div_id].filter(Boolean)"
            :options="divisionOptions"
            :props="{ checkStrictly: true, value: 'id', label: 'name' }"
            placeholder="选择区县/街道/社区"
            clearable
            style="width: 100%"
            @change="handleDivisionChange"
          />
        </el-form-item>

        <!-- Create mode: read-only division path -->
        <el-form-item v-else label="所属区划">
          <div class="division-path">{{ divisionPath || '未选择' }}</div>
        </el-form-item>

        <el-form-item label="小区/村名" prop="name">
          <el-input v-model="formData.name" placeholder="请输入小区名称" maxlength="100" show-word-limit />
        </el-form-item>

        <el-form-item label="地址">
          <el-input v-model="formData.address" type="textarea" :rows="3" placeholder="请输入小区详细地址（选填）" maxlength="255" show-word-limit />
        </el-form-item>

        <el-form-item label="类型" prop="community_type">
          <el-radio-group v-model="formData.community_type">
            <el-radio :label="1">住宅小区</el-radio>
            <el-radio :label="2">村庄</el-radio>
            <el-radio :label="3">混合型</el-radio>
          </el-radio-group>
        </el-form-item>

        <el-form-item>
          <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '更新' : '创建' }}</el-button>
          <el-button @click="handleCancel">取消</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped lang="scss">
.residential-area-form {
  padding: 20px;
  .card-header { font-size: 18px; font-weight: 600; }
  .division-path {
    padding: 0 11px;
    line-height: 32px;
    color: var(--el-text-color-regular);
    background: var(--el-fill-color-light);
    border-radius: 4px;
    min-height: 32px;
  }
}
</style>
