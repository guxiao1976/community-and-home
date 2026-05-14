## 1. Define Category Options

- [x] 1.1 Create category constants array with valid values (政治、色情、暴力、广告, etc.)
- [x] 1.2 Add category options to a shared constants file or inline in the component

## 2. Update Query Form

- [x] 2.1 Replace category text input with el-select dropdown in query form
- [x] 2.2 Populate category dropdown with predefined options
- [x] 2.3 Update severity filter to show only two options: 违规 (value: 3) and 可疑 (value: 1)
- [x] 2.4 Remove the middle severity option (中, value: 2) from query form
- [x] 2.5 Test query form filtering with new category dropdown and severity options

## 3. Update Add/Edit Form

- [x] 3.1 Replace category text input with el-select dropdown in dialog form
- [x] 3.2 Populate category dropdown with the same predefined options
- [x] 3.3 Update severity radio group to show only two options: 违规 (value: 3) and 可疑 (value: 1)
- [x] 3.4 Remove "处理动作" (action) form item entirely from the dialog
- [x] 3.5 Remove action field from form reactive object
- [x] 3.6 Remove action validation rules from rules object
- [x] 3.7 Update form submission to not include action field in API request
- [x] 3.8 Update resetForm to remove action field reset

## 4. Update Table Display

- [x] 4.1 Update severity column to display 违规/可疑 labels instead of 低/中/高
- [x] 4.2 Map severity values: 3 → 违规, 1 → 可疑
- [x] 4.3 Handle legacy severity=2 records (display as 可疑 or add special indicator)
- [x] 4.4 Remove action column from table display

## 5. Testing

- [ ] 5.1 Test query form: select category from dropdown and filter results
- [ ] 5.2 Test query form: select severity (违规/可疑) and filter results
- [ ] 5.3 Test query form: clear filters and verify all records display
- [ ] 5.4 Test add form: select category from dropdown and submit
- [ ] 5.5 Test add form: select severity (违规/可疑) and submit
- [ ] 5.6 Test add form: verify validation errors when required fields are empty
- [ ] 5.7 Test add form: verify action field is not sent in API request
- [ ] 5.8 Test edit form: verify existing records load correctly with new dropdowns
- [ ] 5.9 Verify table displays severity labels correctly (违规/可疑)
- [ ] 5.10 Test end-to-end: create a sensitive word and verify it appears in the table with correct values
