// Community Pinia Store — manages user community memberships and current active community
import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { getUserMemberships, setCurrentCommunity, getAppState } from '@/api/user';
import type { CommunityMembership } from '@common/types/identity';

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
  // 一次性「待加入小区」：join-choice「其他身份认证」路径写入，我的页申请角色时回退消费
  const pendingCommunityId = ref<string>('');

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

      // 服务端权威的当前小区（跨设备一致）：getAppState 返回后端持久化的 current_community_id。
      // 若存在于 memberships 则采用并保存，修复本地 storage 陈旧导致切换/显示不一致。
      // 必须容错：getAppState 失败/缺失时降级忽略，走本地回退逻辑。
      // SEE: [[frontend-business-rule-hardcode]] — 当前小区权威在后端（app-state），前端以服务端为准
      try {
        const appState = await getAppState();
        const serverCurrentId = appState?.current_community_id;
        if (
          serverCurrentId &&
          serverCurrentId !== '0' &&
          communities.value.some(c => c.communityId === serverCurrentId)
        ) {
          currentCommunityId.value = serverCurrentId;
          saveStoredCommunityId(serverCurrentId);
          return;
        }
      } catch (e) {
        // getAppState 失败/缺失 → 降级忽略，走本地回退逻辑；但必须留痕，否则 app-state 接口故障无任何 trace。
        // SEE: [[verify-api-before-calling]] — 禁止空 catch 静默吞错
        console.error('[community] getAppState 获取失败，降级本地', e);
      }

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

  async function switchCommunity(id: string): Promise<void> {
    const exists = communities.value.some(c => c.communityId === id);
    if (!exists) return;
    // 后端持久化当前小区（并校验数据范围）；失败（如 10015 不在数据范围）抛错，
    // 由调用方提示，本地 currentCommunityId 保持不变。
    // SEE: [[frontend-business-rule-hardcode]]
    await setCurrentCommunity(id);
    currentCommunityId.value = id;
    saveStoredCommunityId(id);
  }

  function setPendingCommunityId(id: string): void {
    pendingCommunityId.value = id;
  }

  function clearPendingCommunityId(): void {
    pendingCommunityId.value = '';
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
    pendingCommunityId,
    // Getters
    currentCommunity,
    hasCommunities,
    communityCount,
    // Actions
    loadMemberships,
    switchCommunity,
    addCommunity,
    removeCommunity,
    setPendingCommunityId,
    clearPendingCommunityId,
  };
});
