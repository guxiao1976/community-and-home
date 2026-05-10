# Phase 5 - Community Management Workflow Test Plan

## Test Environment
- Branch: 002-web-pc-admin
- Date: 2026-05-03
- Tester: Claude

## Test Scenarios

### Scenario 1: Create Community (Provincial Admin)
**Steps:**
1. Login as provincial admin
2. Navigate to Communities (/communities)
3. Click "新建社区" button
4. Fill in form:
   - Select division (within admin's scope)
   - Enter community name
   - Enter address
   - Enter area (optional)
   - Enter population (optional)
   - Select community type
5. Click "创建" button

**Expected Results:**
- Community created with status = Draft (0)
- Redirected to community list
- New community appears in the list
- Success message displayed

**Status:** ⏳ Pending Manual Test

---

### Scenario 2: Submit Community for Review
**Steps:**
1. Login as provincial admin
2. Navigate to Communities list
3. Find a Draft community
4. Click "提交审核" button
5. Confirm submission

**Expected Results:**
- Community status changes to Submitted (1)
- Submit button disappears
- Edit button disappears
- Success message displayed

**Status:** ⏳ Pending Manual Test

---

### Scenario 3: Review and Approve Community (Headquarters Admin)
**Steps:**
1. Login as headquarters admin
2. Navigate to Community Review (/communities/review)
3. Find a Submitted community
4. Click "批准" button
5. Enter optional review notes
6. Click "确定"

**Expected Results:**
- Community status changes to Approved (2)
- Review notes saved
- Reviewer ID and review time recorded
- Success message displayed
- Community removed from pending review list

**Status:** ⏳ Pending Manual Test

---

### Scenario 4: Review and Reject Community (Headquarters Admin)
**Steps:**
1. Login as headquarters admin
2. Navigate to Community Review
3. Find a Submitted community
4. Click "拒绝" button
5. Enter rejection notes (required)
6. Click "确定"

**Expected Results:**
- Community status changes to Rejected (3)
- Rejection notes saved
- Reviewer ID and review time recorded
- Success message displayed
- Provincial admin can see rejection notes

**Status:** ⏳ Pending Manual Test

---

### Scenario 5: Edit Rejected Community (Provincial Admin)
**Steps:**
1. Login as provincial admin
2. Navigate to Communities list
3. Find a Rejected community
4. Click "编辑" button
5. Modify community details
6. Click "更新" button

**Expected Results:**
- Community details updated
- Status remains Rejected (3)
- Can resubmit after editing
- Success message displayed

**Status:** ⏳ Pending Manual Test

---

### Scenario 6: Attempt to Edit Approved Community (Provincial Admin)
**Steps:**
1. Login as provincial admin
2. Navigate to Communities list
3. Find an Approved community
4. Verify "编辑" button is hidden
5. Attempt direct URL access to edit page

**Expected Results:**
- Edit button not visible for Approved communities
- Direct URL access shows warning message
- Redirected back to community list
- No changes allowed

**Status:** ⏳ Pending Manual Test

---

### Scenario 7: Administrative Scope Filtering
**Steps:**
1. Login as provincial admin (scope = province_id)
2. Navigate to Communities list
3. Verify only communities within scope are visible
4. Login as headquarters admin (scope = 'all' or null)
5. Navigate to Communities list
6. Verify all communities are visible

**Expected Results:**
- Provincial admin sees only their division's communities
- Headquarters admin sees all communities
- Filter works correctly with user scope

**Status:** ⏳ Pending Manual Test

---

### Scenario 8: Delete Community
**Steps:**
1. Login as admin
2. Navigate to Communities list
3. Find a Draft or Rejected community
4. Click "删除" button
5. Confirm deletion
6. Verify Approved communities cannot be deleted

**Expected Results:**
- Draft/Rejected communities can be deleted
- Approved communities show no delete button
- Soft delete performed (delete_time set)
- Success message displayed
- Community removed from list

**Status:** ⏳ Pending Manual Test

---

### Scenario 9: View Community Details
**Steps:**
1. Login as any admin
2. Navigate to Communities list
3. Click "查看" button on any community
4. Verify all details displayed
5. Check submission info section
6. Check review notes (if reviewed)

**Expected Results:**
- All community details visible
- Submission status displayed with correct tag
- Submitter ID and time shown
- Reviewer ID and time shown (if reviewed)
- Review notes displayed (if any)

**Status:** ⏳ Pending Manual Test

---

## State Transition Validation

### Valid Transitions
- Draft (0) → Submitted (1) ✅
- Submitted (1) → Approved (2) ✅
- Submitted (1) → Rejected (3) ✅
- Rejected (3) → Submitted (1) ✅ (after editing and resubmitting)

### Invalid Transitions (Should be Prevented)
- Approved (2) → Any other state ❌
- Direct Draft (0) → Approved (2) ❌ (must go through Submitted)

---

## Edge Cases to Test

1. **Concurrent Edits**: Two admins editing same community
2. **Token Expiry**: Token expires during form submission
3. **Network Failure**: Network error during submission
4. **Invalid Division**: Select division outside admin's scope
5. **Empty Review Notes**: Reject without notes (should fail)
6. **Long Text**: Very long community name/address
7. **Special Characters**: Community name with special chars
8. **Pagination**: Navigate through multiple pages
9. **Filter Combinations**: Multiple filters applied together
10. **Direct URL Access**: Access edit page for approved community

---

## API Endpoints Used

- `GET /api/masterdata/communities` - List communities
- `GET /api/masterdata/communities/:id` - Get community details
- `POST /api/masterdata/communities` - Create community
- `PUT /api/masterdata/communities/:id` - Update community
- `POST /api/masterdata/communities/:id/submit` - Submit for review
- `POST /api/masterdata/communities/:id/review` - Review (approve/reject)
- `DELETE /api/masterdata/communities/:id` - Delete community

---

## Test Results Summary

| Scenario | Status | Notes |
|----------|--------|-------|
| Create Community | ⏳ Pending | |
| Submit for Review | ⏳ Pending | |
| Approve Community | ⏳ Pending | |
| Reject Community | ⏳ Pending | |
| Edit Rejected | ⏳ Pending | |
| Edit Approved (Blocked) | ⏳ Pending | |
| Scope Filtering | ⏳ Pending | |
| Delete Community | ⏳ Pending | |
| View Details | ⏳ Pending | |

---

## Notes

- All tests require backend services running (Identity + Masterdata)
- Test users with different scopes needed
- Database should have sample divisions created
- Review workflow requires both provincial and headquarters admin accounts

---

## Next Steps

1. Start backend services
2. Create test users (provincial admin, headquarters admin)
3. Create sample divisions
4. Execute test scenarios manually
5. Document any bugs or issues found
6. Update implementation if needed
