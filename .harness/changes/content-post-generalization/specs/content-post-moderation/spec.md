# Content Post Moderation Capability Specification

## Purpose

定义 content_posts 的内容级审核 + Kafka 链路契约：Kafka 基建引入（docker-compose 单节点 KRaft 模式 + 数据卷持久化，D8/Q8）+ `content-review` topic（D17）；发布 submitted 时推送含可再生 file_url 的契约消息（D7/Q7，消费者直接拉取附件内容；契约单源本 capability 且含 version 字段，REVISION）；content_posts 停用 Redis `moderation:task:queue` 只走 Kafka（D3/Q3，lostfound/user 等其他来源仍走 Redis）；moderation-service Redis 消费者对 content_posts 不再回调 NoticeService（D4/Q4，精确跳过判定 source_type="notice"，REVISION）；**Kafka 推送 at-least-once（落库待推标记 + 定时重推，推送失败=该帖审核盲区风险显式登记 + 可观测指标，REVISION）**；moderation-service 扩展消费者（文字先关键字后大模型、图片/pdf 走大模型）后期开发，本期只定契约 + 推送（D16/D18）；本期 submit 即隐式通过 status=approved + published_at=NOW()（无消费者也可见，REVISION）。涉及 docker-compose（Kafka 基建）、community-hub-service（推送方 + 待推标记 + 重推）、moderation-service（消费者扩展后期 + Redis 消费者同步）。

## Requirements

### Requirement: REQ-CPM-1 — Kafka 基建引入（docker-compose 单节点 KRaft + 数据卷持久化 + content-review topic）

The system SHALL provision a Kafka broker in docker-compose as a **single-node KRaft mode** deployment (D8/Q8 kafka-infra-form): no ZooKeeper dependency, `process.roles=broker,controller`, a stable container IP on the app-network (172.19.0.0/24), a data volume for persistence (`./data/kafka-data`), healthcheck, and `depends_on` wiring so dependent services wait for Kafka readiness. The system SHALL create the `content-review` topic (D17). The broker SHALL be reachable from community-hub-service (producer) by a stable advertised listener address resolvable inside the compose network. Startup SHALL be verified by the standard full-stack bring-up (docker compose up + scripts/start.sh) — Kafka must be up before services that push to it start. **The producer-side at-least-once delivery SHALL be part of this change: submission records a pending-push marker on the content post and a reconciliation path (scheduled rescan) retries delivery until acknowledged or quarantined (REVISION REQ-CPB-7 — the reconciliation and the pending-push observable are shipped with the Kafka introduction, not deferred).** The `content-review` topic retention SHALL be set so that messages pushed while no consumer is deployed (this phase) are not lost before the later consumer exists (REVISION — a retention window covering the consumer-rollout gap; exact retention is a config, the contract is that pushed messages SHALL survive until a consumer exists or the pending-push reconciliation re-pushes).

#### Scenario: 全栈启动后 Kafka 可投递 content-review
- **GIVEN** docker compose up has brought up the Kafka broker (KRaft single node) with a persistent data volume
- **WHEN** a producer pushes a message to the `content-review` topic
- **THEN** the message is accepted and can be consumed; the topic exists and the broker remains healthy after a container restart (data volume persists)

#### Scenario: Kafka 未就绪时发布不崩溃但进入待推（at-least-once）
- **GIVEN** the Kafka broker is down or not yet ready when a content post is submitted
- **WHEN** the producer attempts to push the content-review message
- **THEN** the push fails without panicking the publish flow; the post remains persisted and visible (status=approved, D16); the post is recorded as pending-push and the reconciliation scan re-pushes it after the broker recovers until acknowledged or quarantined; the pending-push count is observable via metric/log

### Requirement: REQ-CPM-2 — content-review 消息契约（单源，含 version + 可再生 file_url，REVISION）

The `content-review` topic message SHALL follow the contract defined HERE as the **single, authoritative payload definition** (REVISION — REQ-CPB-7 references this Requirement and does not re-enumerate fields, eliminating the dual-source drift risk). Top-level fields: `version` (int32, REVISION — the contract version, incremented on any incompatible change so the later consumer and producer stay aligned; currently 1), `post_id` (int64 JS_STRING), `section_code` (string), `text` (string, the post body), `publisher_id` (int64 JS_STRING), and `attachments` (repeated) where each element SHALL carry `file_id` (int64 JS_STRING), `file_type` (string), `review_status` (int32 — **at push time this is the pre-review default value (approved, this change); it is the attachment state snapshot as pushed, not a review verdict** — REVISION note), and a **regenerable `file_url`** (D7/Q7 kafka-contract-file-url) — the consumer SHALL be able to pull attachment content directly from `file_url`, and the url is regenerable via file-service `GetFileUrl(file_id)` so it is not a permanent-link dependency (the url may expire after the presigned window; the consumer regenerates). This contract SHALL be stable for the later consumer implementation (D18); any post-ship change requires a `version` bump (REVISION).

#### Scenario: 发布 submitted 推送符合契约的 JSON（含 version）
- **GIVEN** a content post with section_code=notice, body text, and one attachment {file_id, file_type, review_status}
- **WHEN** the post is submitted (REQ-CPB-7/REQ-CPB-9)
- **THEN** the pushed JSON has version=1, post_id/section_code/text/publisher_id and attachments[0] = {file_id, file_type, review_status, file_url} where file_url is a regenerable presigned URL (obtainable via GetFileUrl(file_id))

#### Scenario: 无附件帖推送空附件数组
- **GIVEN** a content post with no attachments being submitted
- **WHEN** the system pushes the content-review message
- **THEN** the message has `attachments` as an empty array (not null) and a consistent `attachment_count` context on the post row

#### Scenario: 契约变更需 version bump（演进载体）
- **GIVEN** the content-review contract in this Requirement
- **WHEN** a future change adds a field or alters the payload structure
- **THEN** the `version` field SHALL be incremented so the consumer and producer can negotiate the payload version (contract evolvable without a topic rename)

### Requirement: REQ-CPM-3 — content_posts 停 Redis 只推 Kafka（D3/Q3）

The system SHALL stop pushing content_posts to the Redis `moderation:task:queue`; the content-post review path SHALL use Kafka `content-review` only (D3/Q3 redis-kafka-dual-write). Other content sources that currently use the Redis moderation queue (e.g. lostfound, user content) SHALL continue using Redis — the Redis queue mechanism and its existing consumer are retained. The change SHALL NOT remove the Redis queue infrastructure or the consumer code paths for the retained sources.

#### Scenario: content_posts 不写 Redis 队列
- **GIVEN** a content post being submitted
- **WHEN** the submission completes
- **THEN** no `moderation:task:queue` LPUSH occurs for this content post (Kafka only); no orphan Redis moderation task is created

#### Scenario: lostfound 仍走 Redis 队列（双轨过渡一致）
- **GIVEN** a lostfound item being created (a source explicitly retained on Redis)
- **WHEN** the creation completes
- **THEN** the existing Redis `moderation:task:queue` push still occurs (unchanged); the Redis consumer still processes it

### Requirement: REQ-CPM-4 — moderation-service Redis 消费者对 content_posts 不再回调 NoticeService（D4/Q4，精确跳过判定 REVISION）

The system SHALL modify the moderation-service Redis consumer so that content_posts tasks SHALL NOT be dispatched to the old `UpdateNoticeModerationStatus` / NoticeService callback path (D4/Q4 proto-backcompat). **The precise skip guard SHALL be: any `moderation:task:queue` task whose `source_type == "notice"` SHALL be skipped (not routed to the retired NoticeService callback) — this is the exact residual-tag value used by the legacy notice publish path (createnoticelogic.go source_type "notice"), and content_posts no longer pushes to Redis this change, so any such task is stale/residual** (REVISION — the tag value is pinned to `"notice"`, not a new `"content_post"` value, because content_posts never enters the Redis queue). The lostfound/user task routing (`source_type` values for those retained sources) SHALL be unchanged. No new callback contract for content_posts is introduced this change (D18 — the Kafka consumer that writes review results back is later-phase). **The retired `UpdateNoticeModerationStatus` RPC in community.proto SHALL be removed together with the NoticeService rename (D21, REVISION — no callback path exists this phase).**

#### Scenario: source_type="notice" 的 Redis 任务被跳过
- **GIVEN** a stale or residual Redis task message in `moderation:task:queue` whose `source_type` equals "notice" (the legacy content-post/notice tag; content_posts no longer enters Redis)
- **WHEN** the moderation-service Redis consumer processes it
- **THEN** the consumer skips the task (no call to the retired NoticeService UpdateNoticeModerationStatus path); other source types (lostfound/user) are still processed normally

#### Scenario: 无内容帖 Redis 任务时的正常消费不变
- **GIVEN** a Redis task message for a lostfound item in `moderation:task:queue`
- **WHEN** the moderation-service Redis consumer processes it
- **THEN** the task is processed through the existing lostfound path unchanged (retained source, D3)

### Requirement: REQ-CPM-5 — 审核消费者后期开发（本期只定契约 + 推送，submit 即隐式通过）

The Kafka `content-review` consumer (text: keyword-filter-then-LLM; image/pdf: LLM; result write-back: body→content_posts.status, attachment→review_status) SHALL be later-phase development (D18) and is NOT implemented by this change. This change SHALL implement only the Kafka producer push (REQ-CPM-2), the stable message contract, and the at-least-once reconciliation (REQ-CPM-1/REQ-CPB-7). **Because no consumer exists this change, submission SHALL be treated as implicit approval: the submit action SHALL set new content posts' `status` to `approved(2)` AND `published_at` to NOW() (REVISION — this is the single, consistent rule replacing the prior ambiguous "status 默认 approved"; attachments' `review_status` SHALL default to `approved`), so content is visible without a consumer and the review-completeness predicate (REQ-CPB-8) is satisfied for posts whose attachments are all approved and whose body is approved.** The later consumer, once developed, SHALL overwrite `status`/`review_status` and `published_at` per the review outcome (REUSE:notice-D27 approval anchoring).

#### Scenario: 无消费者时内容直接可见（本期验收）
- **GIVEN** a content post submitted this change (status=approved, published_at=NOW(), all-approved attachments) and no Kafka consumer deployed
- **WHEN** a user lists or reads the post
- **THEN** the post is visible (审核完整性成立：已审附件数==attachment_count 且 正文 approved)；本期验收即「发布→推 Kafka→（无消费者）→内容直接可见」

#### Scenario: 契约在消费者开发前保持稳定
- **GIVEN** the content-review contract defined in REQ-CPM-2 (version=1)
- **WHEN** a producer pushes messages during this change's lifecycle (before the consumer is built)
- **THEN** the messages conform to the contract unchanged; the later consumer implementation SHALL consume this same contract (契约本期冻结，后续消费者不改推送方；变更走 version bump)

## 服务职责边界

- **docker-compose**: Kafka 单节点 KRaft + 数据卷持久化 + content-review topic + retention 覆盖消费者上线空窗（D8/D17）
- **community-hub-service**: content_posts 提交时推送 content-review 契约消息（at-least-once，落库待推标记 + 定时重推，D20）；停 Redis 队列推送（D3）
- **moderation-service**: Redis 消费者对 content_posts 不再回调 NoticeService（跳过 source_type="notice"，D4）；lostfound/user Redis 消费保留（D3）；Kafka content-review 消费者后期开发（D18，本期不实现）
- **file-service**: file_url 可再生载体（GetFileUrl(file_id)）供契约 file_url 与消费者拉取附件内容（REUSE:notice-D24）
