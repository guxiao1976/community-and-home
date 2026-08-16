// Package sanitize 提供 community-hub-service 公告/内容帖正文（content_posts.text）的
// 写入路径 HTML 白名单净化器（REQ-XSS-1..REQ-XSS-8）。
//
// 决策（specs/xss-sanitization/spec.md 评审钉死）：
//   - D4：净化器位于本服务（不引 community-common），REQ-XSS-3 单例复用。
//
// SEE: [[harness-architecture-decisions]] — D3/D4 子服务 leaf node 自建能力而不复制/引入全局 common
//   - D6：正文 img 全剔除（正文图片走 content_post_attachments 附件机制）。
//   - D8：REQ-XSS-2 穷举允许标签/属性；a 仅 http/https/mailto scheme；target 一律剔除；
//     rel 强制 noopener noreferrer（RequireNoFollowOnLinks + 归一化后处理，单令牌归一化为完整集合）。
//   - D13：纯文本断言以渲染等价为准（HTML 实体转义属允许行为）。
package sanitize

import (
	"regexp"
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// 白名单策略单例（REQ-XSS-3）：进程内仅构建一次；Sanitize 为纯函数，并发安全。
var (
	policyOnce sync.Once
	policy     *bluemonday.Policy

	// generatedAnchorRel 匹配 bluemonday 在 `<a href>` 上生成的 rel 属性
	// （RequireNoFollowOnLinks + RequireNoReferrerOnLinks → 恒为 `rel="nofollow noreferrer"`，
	// 输入 rel 不进入 allowlist，故输出值确定唯一）。
	generatedAnchorRel = regexp.MustCompile(`rel="nofollow noreferrer"`)

	// canonicalAnchorRel 归一化后的 rel 固定令牌集（评审钉死：强制 noopener noreferrer + nofollow）。
	canonicalAnchorRel = `rel="noopener noreferrer nofollow"`
)

// policySingleton 构建并返回共享白名单策略（REQ-XSS-2 穷举允许标签/属性）。
func policySingleton() *bluemonday.Policy {
	policyOnce.Do(func() {
		p := bluemonday.NewPolicy()
		// 允许标签（穷举）：块级 + 行内；非白名单标签整体剔除（子节点保留，script/iframe 原始内容一并剔除）
		p.AllowElements(
			"p", "div", "h1", "h2", "h3", "h4", "h5", "h6",
			"blockquote", "ul", "ol", "li", "pre", "hr",
			"strong", "em", "b", "i", "u", "s", "span", "br",
		)
		// a：href（仅 http/https/mailto scheme 白名单）+ title；style/class/id/on*/target 均不在
		// allowlist → 一律移除（REQ-XSS-2 全局剔除 + target 评审钉死）
		p.AllowAttrs("href").OnElements("a")
		p.AllowURLSchemes("http", "https", "mailto")
		p.AllowAttrs("title").OnElements("a")
		// rel：RequireNoFollowOnLinks 强制 nofollow + RequireNoReferrerOnLinks 强制 noreferrer；
		// noopener 由 normalizeAnchorRel 后处理强制（target 已剔除，bluemonday 不会自动补 noopener）
		p.RequireNoFollowOnLinks(true)
		p.RequireNoReferrerOnLinks(true)
		policy = p
	})
	return policy
}

// ContentPostText 白名单净化公告/内容帖正文（content_posts.text），落库前调用。
//
// 净化范围（REQ-XSS-5）：仅正文 content/text；title 等纯文本字段不经本函数。
// 幂等（REQ-XSS-3）：重复净化输出一致，Update/submit 重净化既有正文不产生渐进式漂移。
// 净化后为空返回空串（D7 唯一化，由调用方决定是否落库）。
func ContentPostText(text string) string {
	if text == "" {
		return ""
	}
	out := policySingleton().Sanitize(text)
	return normalizeAnchorRel(out)
}

// normalizeAnchorRel 将 `<a href>` 上的 rel 属性归一化为完整令牌集 noopener noreferrer nofollow。
//
// bluemonday 恒输出 `rel="nofollow noreferrer"`（输入 rel 不进入 allowlist，值确定唯一）；
// 此处替换为含 noopener 的规范集（REQ-XSS-2 a rel 强制项，评审钉死「单令牌归一化为完整 noopener noreferrer」）。
// 替换保持幂等：二次净化 rel 被 bluemonday 重建后再次归一化到同一值。
func normalizeAnchorRel(s string) string {
	return generatedAnchorRel.ReplaceAllString(s, canonicalAnchorRel)
}
