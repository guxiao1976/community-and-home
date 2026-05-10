import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { AdministrativeDivision } from '@common/types/masterdata'
import {
  getAdministrativeDivisions,
  getAdministrativeDivisionById,
  createAdministrativeDivision,
  updateAdministrativeDivision,
  deleteAdministrativeDivision
} from '@/api/masterdata'

export const useDivisionStore = defineStore('division', () => {
  // State
  const divisions = ref<AdministrativeDivision[]>([])
  const currentDivision = ref<AdministrativeDivision | null>(null)
  const loading = ref(false)
  const treeData = ref<AdministrativeDivision[]>([])

  // Actions
  const fetchDivisions = async (params?: { parent_id?: number; level?: number }) => {
    loading.value = true
    try {
      const response = await getAdministrativeDivisions(params)
      divisions.value = response.list || []
      return response
    } catch (error) {
      console.error('Failed to fetch divisions:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const fetchDivisionById = async (id: number) => {
    loading.value = true
    try {
      const response = await getAdministrativeDivisionById(id)
      currentDivision.value = response
      return currentDivision.value
    } catch (error) {
      console.error('Failed to fetch division:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const createDivision = async (data: {
    parent_id?: number | null
    level: number
    name: string
    code: string
    sort_order?: number
  }) => {
    loading.value = true
    try {
      const response = await createAdministrativeDivision(data)
      await fetchDivisions() // Refresh list
      return response
    } catch (error) {
      console.error('Failed to create division:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const updateDivision = async (id: number, data: {
    name?: string
    sort_order?: number
    status?: number
  }) => {
    loading.value = true
    try {
      await updateAdministrativeDivision(id, data)
      await fetchDivisions() // Refresh list
    } catch (error) {
      console.error('Failed to update division:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const deleteDivision = async (id: number) => {
    loading.value = true
    try {
      await deleteAdministrativeDivision(id)
      divisions.value = divisions.value.filter(d => d.id !== id)
    } catch (error) {
      console.error('Failed to delete division:', error)
      throw error
    } finally {
      loading.value = false
    }
  }

  const buildTree = (flatList: AdministrativeDivision[]): AdministrativeDivision[] => {
    const map = new Map<number, AdministrativeDivision>()
    const roots: AdministrativeDivision[] = []

    // Create map
    flatList.forEach(item => {
      // level < 5 的节点标记有子节点，触发 el-table lazy 展开箭头
      const hasChildren = item.level < 5
      map.set(item.id, { ...item, children: hasChildren ? undefined : undefined, hasChildren })
    })

    // Build tree
    flatList.forEach(item => {
      const node = map.get(item.id)!
      if (item.parent_id) {
        const parent = map.get(item.parent_id)
        if (parent) {
          if (!parent.children) parent.children = []
          parent.children.push(node)
        }
      } else {
        roots.push(node)
      }
    })

    return roots
  }

  const loadTreeData = async () => {
    const allDivisions = await fetchDivisions()
    treeData.value = buildTree(allDivisions)
  }

  const loadChildren = async (parent: AdministrativeDivision) => {
    try {
      // 直接调 API，不经过 fetchDivisions（避免覆盖 divisions.value 导致树重建）
      const response = await getAdministrativeDivisions({ parent_id: parent.id, page_size: 1000 })
      const children = (response as any).list || response || []
      // 为子节点标记 hasChildren
      const childNodes = children.map((item: any) => ({
        ...item,
        hasChildren: item.level < 5
      }))
      parent.children = childNodes
      return childNodes
    } catch (error) {
      console.error('Failed to load children:', error)
      throw error
    }
  }

  const divisionTree = computed(() => buildTree(divisions.value))

  return {
    // State
    divisions,
    currentDivision,
    loading,
    treeData,
    divisionTree,
    // Actions
    fetchDivisions,
    fetchDivisionById,
    createDivision,
    updateDivision,
    deleteDivision,
    loadTreeData,
    loadChildren
  }
})
