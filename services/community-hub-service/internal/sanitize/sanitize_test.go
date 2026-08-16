package sanitize

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContentPostText 白名单净化（REQ-XSS-2 穷举白名单）table-driven：
// 注入剥离 / 合法富文本保留 / scheme 白名单 / rel 归一化 / img+style 剔除 / 幂等。
func TestContentPostText(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		// 正向：合法富文本保留（REQ-XSS-1 Scenario 合法富文本保留）
		{
			name: "合法富文本保留",
			in:   `<p>停水通知</p><p><strong>明早 8 点</strong>恢复供水，<br/>带来不便敬请谅解</p>`,
			want: `<p>停水通知</p><p><strong>明早 8 点</strong>恢复供水，<br/>带来不便敬请谅解</p>`,
		},
		{
			name: "块级+列表保留",
			in:   `<h2>标题</h2><div>容器</div><ul><li>项</li></ul>`,
			want: `<h2>标题</h2><div>容器</div><ul><li>项</li></ul>`,
		},
		// 注入载荷被剥离（REQ-XSS-1 Scenario 注入 payload 被剥离）
		{
			name: "img+script+iframe 注入",
			in:   `<img src=x onerror=alert(1)><script>alert(document.cookie)</script><iframe src="//evil.example"></iframe>`,
			want: ``,
		},
		{
			name: "script 注入单独",
			in:   `<script>alert(1)</script>`,
			want: ``,
		},
		{
			name: "iframe 注入",
			in:   `<iframe src="//evil.example"></iframe>`,
			want: ``,
		},
		// on* 事件属性剥离（REQ-XSS-2 全局剔除）
		{
			name: "onclick 事件属性剥离",
			in:   `<a href="https://x.com" onclick="steal()">x</a>`,
			want: `<a href="https://x.com" rel="noopener noreferrer nofollow">x</a>`,
		},
		// a href scheme 白名单（REQ-XSS-2）
		{
			name: "https 链接保留",
			in:   `<a href="https://example.com/notice">官方通知</a>`,
			want: `<a href="https://example.com/notice" rel="noopener noreferrer nofollow">官方通知</a>`,
		},
		{
			name: "http 链接保留",
			in:   `<a href="http://example.com/notice">官方通知</a>`,
			want: `<a href="http://example.com/notice" rel="noopener noreferrer nofollow">官方通知</a>`,
		},
		{
			name: "mailto 链接保留",
			in:   `<a href="mailto:a@b.com">邮件</a>`,
			want: `<a href="mailto:a@b.com" rel="noopener noreferrer nofollow">邮件</a>`,
		},
		{
			name: "javascript: href 剔除降级",
			in:   `<a href="javascript:alert(1)">点我</a>`,
			want: `点我`,
		},
		{
			name: "data: href 剔除降级",
			in:   `<a href="data:text/html,script">数据</a>`,
			want: `数据`,
		},
		{
			name: "vbscript: href 剔除降级",
			in:   `<a href="vbscript:msgbox(1)">vb</a>`,
			want: `vb`,
		},
		// target 一律剔除（评审钉死：不允许 target 属性）
		{
			name: "target 一律剔除",
			in:   `<a href="https://x.com" target="_blank">链接</a>`,
			want: `<a href="https://x.com" rel="noopener noreferrer nofollow">链接</a>`,
		},
		// rel 归一化：单令牌归一化为完整 noopener noreferrer（+ nofollow）
		{
			name: "rel 单令牌归一化",
			in:   `<a href="https://x.com" rel="noopener">链接</a>`,
			want: `<a href="https://x.com" rel="noopener noreferrer nofollow">链接</a>`,
		},
		{
			name: "rel 完整集保持",
			in:   `<a href="https://x.com" rel="noopener noreferrer">链接</a>`,
			want: `<a href="https://x.com" rel="noopener noreferrer nofollow">链接</a>`,
		},
		// img 全剔除（D6）+ style 属性剔除
		{
			name: "img 全剔除（D6）",
			in:   `<img src="/attachments/1.jpg" onerror="alert(1)">`,
			want: ``,
		},
		{
			name: "img data: 剔除",
			in:   `<img src="data:image/png;base64,AAAA">`,
			want: ``,
		},
		{
			name: "div style 属性剔除",
			in:   `<div style="background:url(javascript:evil())">x</div>`,
			want: `<div>x</div>`,
		},
		{
			name: "span style 属性剔除",
			in:   `<span style="color:red">带样式</span>`,
			want: `<span>带样式</span>`,
		},
		// 白名单外标签剥离后子标签/文本保留（REQ-XSS-2 防御纵深）
		{
			name: "marquee 剥离子标签保留",
			in:   `<marquee><b>滚动标题</b></marquee>`,
			want: `<b>滚动标题</b>`,
		},
		// 纯文本（REQ-XSS-1 边界——纯文本正文，渲染等价）
		{
			name: "纯文本原样",
			in:   `小区明天开展义诊活动`,
			want: `小区明天开展义诊活动`,
		},
		{
			name: "实体转义渲染等价（D13）",
			in:   `A & B < C > D`,
			want: `A &amp; B &lt; C &gt; D`,
		},
		// 空串
		{
			name: "空串",
			in:   ``,
			want: ``,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ContentPostText(tc.in)
			assert.Equal(t, tc.want, got)
			// REQ-XSS-4：净化后不得残留可执行 HTML 特征
			for _, banned := range []string{"<script", "</script", "<iframe", "<img", "onerror=", "onload=", "onclick=", "javascript:", "data:text/html"} {
				assert.NotContains(t, got, banned, "残留可执行特征 %s", banned)
			}
		})
	}
}

// TestContentPostText_Idempotent 幂等（REQ-XSS-3）：s(s(a)) == s(a)，重复净化不产生渐进式漂移。
func TestContentPostText_Idempotent(t *testing.T) {
	inputs := []string{
		`<p>停水通知</p><p><strong>明早 8 点</strong>恢复供水，<br/>带来不便敬请谅解</p>`,
		`<img src=x onerror=alert(1)><script>alert(document.cookie)</script><iframe src="//evil.example"></iframe>`,
		`<a href="https://example.com/notice" target="_blank" rel="noopener">官方通知</a>`,
		`<div style="background:url(javascript:evil())">x</div><marquee><b>滚动标题</b></marquee>`,
		`小区明天开展义诊活动`,
	}
	for _, in := range inputs {
		once := ContentPostText(in)
		twice := ContentPostText(once)
		require.Equal(t, once, twice, "幂等破坏: %q → %q → %q", in, once, twice)
	}
}
