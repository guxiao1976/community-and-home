package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// B2-01: UserInfo JSON marshal — ID 应为字符串
func TestUserInfo_MarshalJSON_IDString(t *testing.T) {
	u := UserInfo{Id: 1234567890123456789, Phone: "13800138000"}
	data, err := json.Marshal(u)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"id":"1234567890123456789"`)
}

// B2-02: UserInfo JSON unmarshal — 接受字符串形式的 ID
func TestUserInfo_UnmarshalJSON_StringID(t *testing.T) {
	input := `{"id":"1234567890123456789","phone":"13800138000"}`
	var u UserInfo
	err := json.Unmarshal([]byte(input), &u)
	require.NoError(t, err)
	assert.Equal(t, int64(1234567890123456789), u.Id)
}

// B2-03: CreateUserResp JSON marshal — user_id 应为字符串
func TestCreateUserResp_MarshalJSON(t *testing.T) {
	resp := CreateUserResp{UserId: 1111111111111111111}
	data, err := json.Marshal(resp)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"user_id":"1111111111111111111"`)
}

// B2-04: 往返测试 — marshal → unmarshal ID 不变
func TestUserInfo_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		id   int64
	}{
		{"small", 1},
		{"max_safe", 9007199254740991},
		{"snowflake_17", 12345678901234567},
		{"snowflake_18", 123456789012345678},
		{"snowflake_19", 1234567890123456789},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			original := UserInfo{Id: tc.id, Phone: "13800138000"}
			data, err := json.Marshal(original)
			require.NoError(t, err)

			var roundtripped UserInfo
			err = json.Unmarshal(data, &roundtripped)
			require.NoError(t, err)

			assert.Equal(t, original.Id, roundtripped.Id)
		})
	}
}

// B2-05: 验证 json:",string" 标签拒绝数字形式的 ID
func TestUserInfo_UnmarshalJSON_NumberID_Rejected(t *testing.T) {
	// When json:",string" is set, encoding/json rejects numeric literals for int64
	input := `{"id":1234567890123456789,"phone":"13800138000"}`
	var u UserInfo
	err := json.Unmarshal([]byte(input), &u)
	assert.Error(t, err, "expected unmarshal failure when sending number to string-tagged int64 field")
}
