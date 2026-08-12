package config

import (
	"path/filepath"
	"testing"

	"github.com/guxiao1976/community-common/v2/pkg/configx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 安全不变式（评审 CRITICAL：RPC 身份伪造）：数据权限身份经 gRPC metadata 传输，
// RPC 必须绑定回环——仅宿主机可信调用方（API 网关 / moderation 回调）可连通，
// 阻断局域网 / Docker 桥接对端注入 user_id。回退 0.0.0.0 即击穿数据权限。
//
// SEE: [[rpc-identity-spoofing-loopback-isolation]] — RPC 身份伪造风险 + 回环隔离缓解
func TestRpcConfig_BindsLoopback(t *testing.T) {
	t.Setenv("MYSQL_USER", "root")
	t.Setenv("MYSQL_PASSWORD", "test")
	t.Setenv("REDIS_PASSWORD", "")

	var c Config
	path := filepath.Join("..", "..", "etc", "communityhub.yaml")
	configx.MustLoad(path, &c)

	require.NotEmpty(t, c.ListenOn, "必须显式配置 ListenOn，不能走零值默认")
	host, _, ok := splitHostPort(t, c.ListenOn)
	require.True(t, ok, "ListenOn 必须为 host:port 形式: %s", c.ListenOn)
	assert.True(t, host == "127.0.0.1" || host == "localhost",
		"RPC 必须绑定回环（127.0.0.1/localhost），当前 host=%q → 允许局域网/Docker 桥接对端伪造身份", host)
}

func splitHostPort(t *testing.T, addr string) (string, string, bool) {
	t.Helper()
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i], addr[i+1:], true
		}
	}
	return "", "", false
}
