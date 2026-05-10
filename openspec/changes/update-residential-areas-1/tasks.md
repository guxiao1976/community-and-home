## Implementation Tasks

- [x] T1: Modify `List.vue` filters reactive state — replace `county_id` with `province_id`, `city_id`, `county_id` three fields, add `provinceOptions`, `cityOptions`, `districtOptions` refs
- [x] T2: Implement `loadProvinces()` — call `getAdministrativeDivisions({ level: 1 })` on page mount, populate `provinceOptions`
- [x] T3: Implement `handleProvinceChange()` — on province select, clear city/district, call `getAdministrativeDivisions({ parent_id, level: 2 })` to load `cityOptions`
- [x] T4: Implement `handleCityChange()` — on city select, clear district, call `getAdministrativeDivisions({ parent_id, level: 3 })` to load `districtOptions`
- [x] T5: Implement `handleDistrictChange()` — on district select, set `filters.county_id`
- [x] T6: Replace the existing `el-cascader` in template with three `el-select` components (省份/城市/区县), bind options and change handlers
- [x] T7: Add search button disabled state — `:disabled="!filters.city_id"`
- [x] T8: Remove old `loadDivisions()`, `buildDivisionTree()`, `handleDivisionChange()`, and `divisionOptions` code
