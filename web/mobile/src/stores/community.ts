// Community Pinia Store — manages user community memberships and current active community
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { getUserMemberships } from '@/api/user';
import type { CommunityMembership } from '@/api/user';

export interface CommunityInfo {
  communityId: string;
  communityName: string;
  address?: string;
}

function loadStoredCommunityId(): string {
  try {
    return uni.getStorageSync('current_community_id') || '';
  } catch {
    return '';
  }
}

function saveStoredCommunityId(id: string): void {
  try {
    uni.setStorageSync('current_community_id', id);
  } catch {
    // ignore storage errors
  }
}

export const useCommunityStore = defineStore('community', () => {
  // --- State ---
  const communities = ref<CommunityInfo[]>([]);
  const currentCommunityId = ref<string>(loadStoredCommunityId());

  // --- Getters ---
  const currentCommunity = computed<CommunityInfo | null>(() => {
    if (!currentCommunityId.value) return null;
    return communities.value.find(c => c.communityId === currentCommunityId.value) || null;
  });

  const hasCommunities = computed(() => communities.value.length > 0);

  const communityCount = computed(() => communities.value.length);

  // --- Actions ---
  async function loadMemberships(): Promise<void> {
    try {
      const memberships = await getUserMemberships();
      // API returns snake_case proto fields (community_id), map to camelCase
      // Note: membership API does NOT return community name/address

      // Resolve community names from master-data-service
      const ids = memberships.map((m: any) => m.community_id || m.communityId || '').filter(Boolean);
      let nameMap: Map<string, string> = new Map();
      if (ids.length > 0) {
        try {
          const { getResidentialAreasByIds } = await import('@/api/user');
          const areas = await getResidentialAreasByIds(ids);
          areas.forEach(a => nameMap.set(a.id, a.name));
        } catch { /* ignore - will use existing names or ID fallback */ }
      }

      const existingMap = new Map(communities.value.map(c => [c.communityId, c]));
      communities.value = memberships.map((m: any) => {
        const cid = m.community_id || m.communityId || '';
        const existing = existingMap.get(cid);
        return {
          communityId: cid,
          // Preserve existing name/address if available, otherwise use ID fallback
          communityName: existing?.communityName || nameMap.get(cid) || m.community_name || m.communityName || ('小区 ' + cid.slice(-6)),
          address: existing?.address || m.address || undefined,
        };
      });

      // If no current selection (first load), pick first community
      if (!currentCommunityId.value && communities.value.length > 0) {
        currentCommunityId.value = communities.value[0].communityId;
        saveStoredCommunityId(currentCommunityId.value);
      }
    } catch {
      communities.value = [];
    }
  }

  function removeCommunity(id: string): void {
    communities.value = communities.value.filter(c => c.communityId !== id);
    // If current community was removed, switch to first remaining or clear
    if (currentCommunityId.value === id) {
      const next = communities.value[0];
      currentCommunityId.value = next?.communityId || '';
      saveStoredCommunityId(currentCommunityId.value);
    }
  }

  function switchCommunity(id: string): void {
    const exists = communities.value.some(c => c.communityId === id);
    if (!exists) return;
    currentCommunityId.value = id;
    saveStoredCommunityId(id);
  }

  function addCommunity(membership: { communityId: string; communityName?: string; address?: string }): void {
    const exists = communities.value.some(c => c.communityId === membership.communityId);
    if (exists) return;

    communities.value.push({
      communityId: membership.communityId,
      communityName: membership.communityName || ('小区 ' + membership.communityId),
      address: membership.address,
    });

    // Auto-select newly joined community
    currentCommunityId.value = membership.communityId;
    saveStoredCommunityId(membership.communityId);
  }

  return {
    // State
    communities,
    currentCommunityId,
    // Getters
    currentCommunity,
    hasCommunities,
    communityCount,
    // Actions
    loadMemberships,
    switchCommunity,
    addCommunity,
    removeCommunity,
  };
});
