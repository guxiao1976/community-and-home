# TDD 证据 — community-hub-service content-post-generalization（Task 1.1-1.23）RED 摘录补录

> 生成时间: 2026-08-16
> 背景: 上一轮 QA 判 FAIL——TDD 证据表 RED 列全部 ❌（~20 个有逻辑函数无具体 FAIL 摘录）。本文件为第 4 次复发（`tdd-red-evidence-requires-fail-excerpt` 记忆）的补救：**用 git 回退生产文件复现真实编译失败并持久化摘录**。
> 复现方式: 在父提交 `dca1225`（`git worktree add --detach`）叠加 HEAD 新增的 13 个测试文件 + HEAD go.mod/go.sum（kafka-go 依赖，模拟"测试已写、依赖已加、实现缺失"的 RED 态），`GOWORK=off go test ./...` 捕获真实 `go test` 编译输出。摘录均为 go 编译器实际输出（含 `undefined:` 行号），非注释/口头描述。
> 结构性佐证（辅助，不替代摘录）: `git show HEAD:` 确认 division.go / producer.go / contentcompat.go / helper.go 等新符号在父提交均不存在。

---

## 1. RED — model 包（ContentPostModel 读查询 / ContentPostScopeModel / ContentPostAttachmentModel / IsReviewComplete）

复现命令: `go test ./model/`（父提交 model 生产文件 + HEAD 测试文件）

```
model/content_post_attachment_test.go:21:7: undefined: NewContentPostAttachmentModel
model/content_post_attachment_test.go:24:13: undefined: ContentPostAttachment
model/content_post_attachment_test.go:25:88: undefined: AttachmentReviewApproved
model/content_post_attachment_test.go:26:88: undefined: AttachmentReviewApproved
model/content_post_attachment_test.go:47:7: undefined: NewContentPostAttachmentModel
model/content_post_attachment_test.go:70:7: undefined: NewContentPostAttachmentModel
model/content_post_scope_test.go:21:7: undefined: NewContentPostScopeModel
model/content_post_scope_test.go:39:7: undefined: NewContentPostScopeModel
model/content_post_scope_test.go:53:7: undefined: NewContentPostScopeModel
model/content_post_scope_test.go:71:7: undefined: NewContentPostScopeModel
model/content_post_scope_test.go:71:7: too many errors
```

说明: `content_post_attachment_test.go`/`content_post_scope_test.go` 引用的 `NewContentPostAttachmentModel`/`NewContentPostScopeModel`/`ContentPostAttachment`/`AttachmentReviewApproved` 在实现前不存在（Task 1.3/1.4 新模型）。

---

## 2. RED — contentcompat 包（ResolveReadableCommunityForCompat / ScopeFilter / CodePostNotFound）

复现命令: `go test ./internal/contentcompat/`

```
internal/contentcompat/contentcompat_test.go:16:8: undefined: model.ContentPostModel
internal/contentcompat/contentcompat_test.go:21:86: undefined: model.ContentPost
internal/contentcompat/contentcompat_test.go:32:8: undefined: model.ContentPostScopeModel
internal/contentcompat/contentcompat_test.go:49:12: undefined: ScopeFilter
internal/contentcompat/contentcompat_test.go:71:14: undefined: CodePostNotFound
internal/contentcompat/contentcompat_test.go:103:16: undefined: ResolveReadableCommunityForCompat
internal/contentcompat/contentcompat_test.go:9:2: too many errors
```

说明: Task 1.14/1.23 详情兼容回退 `ResolveReadableCommunityForCompat`（scope 反查 + 逐小区 FilterAllowed）在实现前不存在。

---

## 3. RED — kafkapush 包（Producer / Rescanner / ContentReviewMessage）

复现命令: `go test ./rpc/internal/kafkapush/`

```
rpc/internal/kafkapush/producer_test.go:41:8: undefined: model.ContentPostModel
rpc/internal/kafkapush/producer_test.go:63:32: undefined: model.ContentPost
rpc/internal/kafkapush/producer_test.go:71:33: undefined: model.ContentPostAttachment
rpc/internal/kafkapush/producer_test.go:152:46: undefined: ContentReviewMessage
rpc/internal/kafkapush/rescanner_test.go:30:8: undefined: model.ContentPostAttachmentModel
rpc/internal/kafkapush/rescanner_test.go:35:89: undefined: model.ContentPostAttachment
rpc/internal/kafkapush/producer_test.go:65:16: too many errors
```

说明: Task 1.18/1.19 Kafka at-least-once 推送（`ContentReviewMessage` 契约 + Producer.Push + Rescanner 重推）在实现前不存在。

---

## 4. RED — scope 包（ExpandDivisionCommunities / ResolveAdminDivision / PublishRolesFrom）

复现命令: `go test ./rpc/internal/logic/scope/`

```
rpc/internal/logic/scope/division_test.go:85:16: undefined: CodeInvalidParam
rpc/internal/logic/scope/division_test.go:120:16: undefined: ExpandDivisionCommunities
rpc/internal/logic/scope/division_test.go:150:20: undefined: RoleCommunityAdmin
rpc/internal/logic/scope/division_test.go:150:66: undefined: UserRoleStatusVerified
rpc/internal/logic/scope/division_test.go:162:20: undefined: RoleCommittee
rpc/internal/logic/scope/division_test.go:166:14: undefined: CodeInvalidParam
rpc/internal/logic/scope/division_test.go:210:14: undefined: CodeInvalidParam
rpc/internal/logic/scope/division_test.go:162:20: too many errors
```

说明: Task 1.7/1.8 社区管理员 division 展开 + 发布角色派生（`ExpandDivisionCommunities`/`RoleCommunityAdmin`/`UserRoleStatusVerified`）在实现前不存在。

---

## 5. RED — notice 逻辑包（Create/Update/List/Get/Delete ContentPost + 附件绑定 + is_pinned 分流）

复现命令: `go test ./rpc/internal/logic/notice/`

```
rpc/internal/logic/notice/notice_helpers_test.go:26:8: undefined: model.ContentPostModel
rpc/internal/logic/notice/notice_helpers_test.go:27:27: undefined: model.ContentPost
rpc/internal/logic/notice/notice_helpers_test.go:28:27: undefined: model.ContentPost
rpc/internal/logic/notice/notice_helpers_test.go:31:29: undefined: model.ContentPost
rpc/internal/logic/notice/notice_helpers_test.go:32:27: undefined: model.ContentPost
rpc/internal/logic/notice/notice_helpers_test.go:32:27: too many errors
```

说明: Task 1.10-1.16 各 Logic 的测试（create/updatecontentpostlogic_test、read_write_logic_test、notice_helpers_test）引用的 `model.ContentPostModel`/`model.ContentPost` 在实现前不存在（旧模型仅 `Notice`/`NoticeModel`）。

---

## 6. RED — API 代理层（Task 1.22/1.23 REST wire 兼容）

复现命令: `go test ./api/internal/logic/notice/`

```
api/internal/svc/servicecontext.go:17:34: undefined: communityv1.NoticeServiceClient
api/internal/svc/servicecontext.go:35:36: undefined: communityv1.NewNoticeServiceClient
FAIL	github.com/guxiao1976/community-hub/api/internal/logic/notice [build failed]
```

说明: proto 改名 `NoticeService → ContentPostService`（Task 0.1）后旧 svc 接线被击穿，正是 Task 1.23 代理层要修复的状态（`api_proxy_test.go` 测试引用的新代理逻辑同样不满足编译，属同一编译失败类）。

---

## GREEN 确认

- `go test ./... -count=1`（工作树 HEAD）: **13 包 / ~119 测试函数全绿，exit 0**（QA 2026-08-16 11:10 现场复跑确认 + 本轮 harness-checks go_test PASS 复证）。
- 行为断言精确性由 QA 复核确认：080005/080006/080002 映射、attachment_count 重算、is_pinned 操作者分流、Kafka 推送 ack/pending、compat 回退均被断言。

---

## 复现命令（every-fresh-run 证据）

```bash
# RED 复现（只读，临时 worktree，不扰动主树）
git worktree add --detach <tmp>/red-repro dca1225
# 叠加 13 个新测试文件 + HEAD go.mod/go.sum，symlink api-proto/common
cd <tmp>/red-repro/services/community-hub-service
GOWORK=off go test ./...   # 上述各包 FAIL（build failed）
git worktree remove <tmp>/red-repro --force

# GREEN 复现
cd services/community-hub-service && go test ./... -count=1 && echo $?   # 0
bash .harness/skills/qa/scripts/harness-checks.sh --service community-hub-service --json   # 19P/0F/2W exit 0
```

---

## 公告正文 XSS 净化（notice-xss-sanitize-and-frontend-fixes / xss-sanitization）RED 摘录

> 背景: 存储型 XSS 修复——新增 `internal/sanitize` 白名单净化器（bluemonday）+ Create/Update(submit) 写路径接入。RED 均为 `go test` 实际输出（先写测试、未实现时执行），非口头描述。

## 1. RED — internal/sanitize 包（编译期：undefined ContentPostText）

复现命令: `go test ./internal/sanitize/`（仅测试文件、无实现）

```
# github.com/guxiao1976/community-hub/internal/sanitize [github.com/guxiao1976/community-hub/internal/sanitize.test]
internal/sanitize/sanitize_test.go:146:11: undefined: ContentPostText
internal/sanitize/sanitize_test.go:166:11: undefined: ContentPostText
internal/sanitize/sanitize_test.go:167:12: undefined: ContentPostText
FAIL	github.com/guxiao1976/community-hub/internal/sanitize [build failed]
FAIL
```

## 2. RED — CreateContentPost 落库未净化（行为断言失败）

复现命令: `go test ./rpc/internal/logic/notice/ -run TestCreateContentPost_SanitizesText -count=1`

```
--- FAIL: TestCreateContentPost_SanitizesText (0.00s)
    createcontentpostlogic_test.go:189:
        	Error:      	Not equal:
        	            	expected: "安全文本"
        	            	actual  : "<script>alert(document.cookie)</script><img src=x onerror=alert(1)>安全文本"
    createcontentpostlogic_test.go:190:
        	Error:      	"<script>alert(document.cookie)</script><img src=x onerror=alert(1)>安全文本" should not contain "<script"
    createcontentpostlogic_test.go:191:
        	Error:      	"<script>alert(document.cookie)</script><img src=x onerror=alert(1)>安全文本" should not contain "onerror="
```

## 3. RED — UpdateContentPost 内容编辑分支落库未净化

复现命令: `go test ./rpc/internal/logic/notice/ -run TestUpdateContentPost_ContentEdit_SanitizesText -count=1`

```
--- FAIL: TestUpdateContentPost_ContentEdit_SanitizesText (0.00s)
    updatecontentpostlogic_test.go:165:
        	Error:      	Not equal:
        	            	expected: "净化后正文"
        	            	actual  : "<script>alert(1)</script><iframe src=x></iframe>净化后正文"
    updatecontentpostlogic_test.go:166:
        	Error:      	"<script>alert(1)</script><iframe src=x></iframe>净化后正文" should not contain "<script"
    updatecontentpostlogic_test.go:167:
        	Error:      	"<script>alert(1)</script><iframe src=x></iframe>净化后正文" should not contain "<iframe"
```

## 4. RED — UpdateContentPost submit 发布分支存量 draft 未净化

复现命令: `go test ./rpc/internal/logic/notice/ -run TestUpdateContentPost_Submit_SanitizesDraftText -count=1`

```
--- FAIL: TestUpdateContentPost_Submit_SanitizesDraftText (0.00s)
    updatecontentpostlogic_test.go:221:
        	Error:      	Should be true
        	Messages:   	置公开前先净化存量 draft 正文（UpdateContentTx）
    updatecontentpostlogic_test.go:223:
        	Error:      	Not equal:
        	            	expected: "存量正文"
        	            	actual  : ""
        	Messages:   	净化后正文写入同一事务
```

## GREEN 确认

- `go test ./internal/sanitize/` : `TestContentPostText`（23 用例）+ `TestContentPostText_Idempotent` 全 PASS。
- `go test ./rpc/internal/logic/notice/ -run 'Sanitizes|AlreadySanitized|TextNotPresent'` : 5 用例全 PASS（create 净化 / 内容编辑净化 / 未携带不重净化 / submit 净化 + 置公开 / 幂等不二次改写）。
- 全量 `go build ./...` + `go vet ./...` + `go test ./... -count=1` : 13 包全绿，exit 0。

## 2026-08-16 — Kafka 推送 payload 修正（Review must-follow 跟进）

### RED — TestUpdateContentPost_Submit_SanitizesTextBeforeKafkaPush
复现: 新增测试后运行（修复前 submit 分支 Push 复用 FindOne 未净化快照）。

```
--- FAIL: TestUpdateContentPost_Submit_SanitizesTextBeforeKafkaPush (0.00s)
        	Messages:   	推送 payload 不得含 script
        	Error:      	Not equal:
        	            	expected: "<p>你好</p>"
        	            	actual  : "<p>你好</p><img src=x onerror=alert(1)><script>alert(2)</script>"
        	Messages:   	推送 payload == 落库净化值
FAIL
```

### GREEN
修复（提交后 `post.Text = sanitizedText` 再 Push）后：
```
--- PASS: TestUpdateContentPost_Submit_SanitizesTextBeforeKafkaPush (0.00s)
ok  	github.com/guxiao1976/community-hub/rpc/internal/logic/notice	0.029s
```
