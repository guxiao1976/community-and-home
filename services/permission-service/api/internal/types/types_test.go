// Code scaffolded by goctl. Safe to edit.
package types

import (
	"encoding/json"
	"math"
	"testing"
)

// SEE: [[proto-jstype]] — 验证 int64 字段是否正确序列化为字符串以避免 JavaScript 精度丢失
func TestInt64FieldsSerializeAsString(t *testing.T) {
	// 使用超过 JavaScript Number.MAX_SAFE_INTEGER (2^53-1) 的值
	largeID := int64(9007199254740992) // 2^53, 会在 JS Number 中丢失精度

	tests := []struct {
		name     string
		value    interface{}
		jsonPath string
		wantType string // "string" or "number"
	}{
		{
			name: "PageInfo.Total should be string",
			value: PageInfo{
				Page:       1,
				PageSize:   10,
				Total:      largeID,
				TotalPages: 1,
			},
			jsonPath: "total",
			wantType: "string",
		},
		{
			name: "RoleInfo.CreatedAt should be string",
			value: RoleInfo{
				Id:        largeID,
				Code:      "test",
				Name:      "test",
				CreatedAt: largeID,
				UpdatedAt: largeID,
			},
			jsonPath: "createdAt",
			wantType: "string",
		},
		{
			name: "RoleInfo.UpdatedAt should be string",
			value: RoleInfo{
				Id:        largeID,
				Code:      "test",
				Name:      "test",
				CreatedAt: largeID,
				UpdatedAt: largeID,
			},
			jsonPath: "updatedAt",
			wantType: "string",
		},
		{
			name: "CreateRoleReq.PermissionIds should be string array",
			value: CreateRoleReq{
				Code:          "test",
				Name:          "test",
				PermissionIds: Int64Array{largeID, largeID + 1},
			},
			jsonPath: "permissionIds",
			wantType: "string", // array elements should be strings
		},
		{
			name: "UpdateRoleReq.PermissionIds should be string array",
			value: UpdateRoleReq{
				Id:            largeID,
				PermissionIds: Int64Array{largeID, largeID + 1},
			},
			jsonPath: "permissionIds",
			wantType: "string", // array elements should be strings
		},
		{
			name: "PermissionInfo.CreatedAt should be string",
			value: PermissionInfo{
				Id:        largeID,
				ParentId:  0,
				Code:      "test",
				Name:      "test",
				Type:      1,
				CreatedAt: largeID,
				UpdatedAt: largeID,
			},
			jsonPath: "createdAt",
			wantType: "string",
		},
		{
			name: "PermissionInfo.UpdatedAt should be string",
			value: PermissionInfo{
				Id:        largeID,
				ParentId:  0,
				Code:      "test",
				Name:      "test",
				Type:      1,
				CreatedAt: largeID,
				UpdatedAt: largeID,
			},
			jsonPath: "updatedAt",
			wantType: "string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBytes, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var result map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			field := result[tt.jsonPath]
			if field == nil {
				t.Fatalf("Field %s not found in JSON", tt.jsonPath)
			}

			switch v := field.(type) {
			case string:
				if tt.wantType != "string" {
					t.Errorf("Field %s is string, want %s", tt.jsonPath, tt.wantType)
				}
			case float64:
				if tt.wantType != "number" {
					t.Errorf("Field %s is number, want %s. This will cause precision loss in JavaScript!", tt.jsonPath, tt.wantType)
				}
			case []interface{}:
				if tt.wantType == "string" {
					// Check if array elements are strings
					if len(v) > 0 {
						if _, ok := v[0].(string); !ok {
							t.Errorf("Field %s array elements are not strings. This will cause precision loss in JavaScript!", tt.jsonPath)
						}
					}
				}
			default:
				t.Errorf("Field %s has unexpected type: %T", tt.jsonPath, field)
			}
		})
	}
}

// SEE: [[proto-jstype]] — 验证大数字在 JSON 往返过程中不会丢失精度
func TestInt64PrecisionPreservation(t *testing.T) {
	tests := []struct {
		name  string
		value int64
	}{
		{"Max safe integer", int64(math.Pow(2, 53) - 1)},
		{"Above safe integer", int64(math.Pow(2, 53))},
		{"Snowflake ID range", 1234567890123456789},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := RoleInfo{
				Id:        tt.value,
				Code:      "test",
				Name:      "test",
				CreatedAt: tt.value,
				UpdatedAt: tt.value,
			}

			jsonBytes, err := json.Marshal(original)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			var decoded RoleInfo
			if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
				t.Fatalf("Unmarshal failed: %v", err)
			}

			if decoded.CreatedAt != original.CreatedAt {
				t.Errorf("CreatedAt precision lost: got %d, want %d", decoded.CreatedAt, original.CreatedAt)
			}
			if decoded.UpdatedAt != original.UpdatedAt {
				t.Errorf("UpdatedAt precision lost: got %d, want %d", decoded.UpdatedAt, original.UpdatedAt)
			}
		})
	}
}
