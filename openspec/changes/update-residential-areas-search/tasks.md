## Implementation Tasks

- [x] T1: Add `street_id`, `community_div_id` to filters reactive state, add `streetOptions` and `communityOptions` refs
- [x] T2: Implement `handleDistrictChange()` — clear street/community, call `getAdministrativeDivisions({ parent_id, level: 4 })` to load `streetOptions`
- [x] T3: Implement `handleStreetChange()` — clear community, call `getAdministrativeDivisions({ parent_id, level: 5 })` to load `communityOptions`
- [x] T4: Implement `handleCommunityChange()` — set `filters.community_div_id`
- [x] T5: Add two `el-select` in template (街道/乡镇, 社区), bind options and change handlers, disabled states chained to parent selection
- [x] T6: Update `handleReset()` to clear `street_id`, `community_div_id` and reset `streetOptions`, `communityOptions`
- [x] T7: Update `loadResidentialAreas()` — pass `community_div_id` when set, otherwise fall back to `county_id`
