package user

import (
	"context"
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
}

