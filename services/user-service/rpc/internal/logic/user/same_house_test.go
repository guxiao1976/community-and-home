package user

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 3.6: GetUser 同屋互见
// =============================================================================

func TestMaskPhone(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"11-digit", "13812341234", "138****1234"},
		{"short-non-phone", "phone_1001", "phone_1001"},
		{"empty", "", ""},
		{"cipher-blob-fallback", "RAW_ENCRYPTED_BLOB_VALUE", "RAW_ENCRYPTED_BLOB_VALUE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, maskPhone(tt.in))
		})
	}
}

func TestIsSameHouse(t *testing.T) {
	t.Run("same community and address", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 11, 2001, 1, 2, 301)
		createTestMembershipAt(t, mm, 2, 12, 2001, 1, 2, 301)

		same, b, u, r, err := isSameHouse(context.Background(), svc, 11, 12)
		require.NoError(t, err)
		assert.True(t, same)
		assert.Equal(t, int32(1), b)
		assert.Equal(t, int32(2), u)
		assert.Equal(t, int32(301), r)
	})

	t.Run("different address", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 21, 2001, 1, 2, 301)
		createTestMembershipAt(t, mm, 2, 22, 2001, 1, 2, 302)

		same, _, _, _, err := isSameHouse(context.Background(), svc, 21, 22)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("different community", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 31, 2001, 1, 2, 301)
		createTestMembershipAt(t, mm, 2, 32, 2002, 1, 2, 301)

		same, _, _, _, err := isSameHouse(context.Background(), svc, 31, 32)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("zero-address members are not same house", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 41, 2001, 0, 0, 0)
		createTestMembershipAt(t, mm, 2, 42, 2001, 0, 0, 0)

		same, _, _, _, err := isSameHouse(context.Background(), svc, 41, 42)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("partial zero address building", func(t *testing.T) {
		// 地址部分为 0（如 building=0）也应跳过，不误判同屋
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 81, 2001, 0, 2, 301)
		createTestMembershipAt(t, mm, 2, 82, 2001, 0, 2, 301)
		same, _, _, _, err := isSameHouse(context.Background(), svc, 81, 82)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("partial zero address unit", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 83, 2001, 1, 0, 301)
		createTestMembershipAt(t, mm, 2, 84, 2001, 1, 0, 301)
		same, _, _, _, err := isSameHouse(context.Background(), svc, 83, 84)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("partial zero address room", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 85, 2001, 1, 2, 0)
		createTestMembershipAt(t, mm, 2, 86, 2001, 1, 2, 0)
		same, _, _, _, err := isSameHouse(context.Background(), svc, 85, 86)
		require.NoError(t, err)
		assert.False(t, same)
	})

	t.Run("viewer membership find error", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		mm.findErr = errors.New("viewer db down")
		_, _, _, _, err := isSameHouse(context.Background(), svc, 51, 52)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "viewer db down")
	})

	t.Run("target membership find error", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 61, 2001, 1, 2, 301)
		mm.findErr = errors.New("target db down")
		// 第一个 FindByUserId（viewer=61）成功后，第二个（target=62）触发 error
		_, _, _, _, err := isSameHouse(context.Background(), svc, 61, 62)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "target db down")
	})
}

func TestOwnHouseInfo(t *testing.T) {
	t.Run("has membership returns first house", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		createTestMembershipAt(t, mm, 1, 71, 2001, 3, 4, 502)
		b, u, r := ownHouseInfo(context.Background(), svc, 71)
		assert.Equal(t, int32(3), b)
		assert.Equal(t, int32(4), u)
		assert.Equal(t, int32(502), r)
	})

	t.Run("no membership returns zeros", func(t *testing.T) {
		svc := testSvc(t)
		b, u, r := ownHouseInfo(context.Background(), svc, 72)
		assert.Equal(t, int32(0), b)
		assert.Equal(t, int32(0), u)
		assert.Equal(t, int32(0), r)
	})

	t.Run("find error returns zeros", func(t *testing.T) {
		svc := testSvc(t)
		mm := membershipModel(svc)
		mm.findErr = errors.New("db down")
		b, u, r := ownHouseInfo(context.Background(), svc, 73)
		assert.Equal(t, int32(0), b)
		assert.Equal(t, int32(0), u)
		assert.Equal(t, int32(0), r)
	})
}

