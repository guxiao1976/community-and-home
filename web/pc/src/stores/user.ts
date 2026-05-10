// User management store

import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User, UserFilter } from '@common/types/identity'

export const useUserStore = defineStore('user', () => {
  // State
  const users = ref<User[]>([])
  const total = ref(0)
  const loading = ref(false)

  // Filters
  const filters = ref<UserFilter & { page: number; page_size: number }>({
    page: 1,
    page_size: 20
  })

  // Computed
  const hasUsers = computed(() => users.value.length > 0)

  // Actions
  const setUsers = (list: User[], totalCount: number) => {
    users.value = list
    total.value = totalCount
  }

  const updateUserInList = (userId: number, updates: Partial<User>) => {
    const index = users.value.findIndex(u => u.id === userId)
    if (index !== -1) {
      users.value[index] = { ...users.value[index], ...updates }
    }
  }

  const removeUserFromList = (userId: number) => {
    const index = users.value.findIndex(u => u.id === userId)
    if (index !== -1) {
      users.value.splice(index, 1)
      total.value--
    }
  }

  const setFilters = (newFilters: Partial<typeof filters.value>) => {
    filters.value = { ...filters.value, ...newFilters }
  }

  const resetFilters = () => {
    filters.value = {
      page: 1,
      page_size: 20
    }
  }

  const setLoading = (value: boolean) => {
    loading.value = value
  }

  return {
    // State
    users,
    total,
    loading,
    filters,
    // Computed
    hasUsers,
    // Actions
    setUsers,
    updateUserInList,
    removeUserFromList,
    setFilters,
    resetFilters,
    setLoading
  }
})
