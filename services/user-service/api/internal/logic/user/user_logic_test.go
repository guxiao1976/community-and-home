package user

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	userv1 "github.com/guxiao1976/api-proto/gen/go/user/v1"
	usermocks "github.com/guxiao1976/api-proto/gen/go/user/v1/mocks"
	"github.com/guxiao1976/community-user/api/internal/svc"
	"github.com/guxiao1976/community-user/api/internal/types"
)

// TestCreateUserLogic_CreateUser tests the CreateUser logic
func TestCreateUserLogic_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.CreateUserReq
		mockSetup func(*gomock.Controller) *svc.ServiceContext
		want      *types.CreateUserResp
		wantErr   bool
	}{
		{
			name: "success - 正常创建用户",
			req: &types.CreateUserReq{
				Phone:    "13800138000",
				Nickname: "测试用户",
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				// 设置期望调用
				mockUserRpc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(&userv1.CreateUserResponse{UserId: 1}, nil)

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			want: &types.CreateUserResp{
				UserId: 1, // int64 with json:",string" tag
			},
			wantErr: false,
		},
		{
			name: "error - 手机号为空",
			req: &types.CreateUserReq{
				Phone:    "",
				Nickname: "测试用户",
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				return &svc.ServiceContext{}
			},
			wantErr: true,
		},
		{
			name: "error - 手机号已存在",
			req: &types.CreateUserReq{
				Phone:    "13800138000",
				Nickname: "测试用户",
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				// Mock: UserRpc.CreateUser 返回"手机号已存在"错误
				mockUserRpc.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("手机号已存在"))

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svcCtx := tt.mockSetup(ctrl)
			l := NewCreateUserLogic(context.Background(), svcCtx)

			// Execute
			got, err := l.CreateUser(tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			if tt.want != nil {
				assert.Equal(t, tt.want.UserId, got.UserId)
			}
		})
	}
}

// TestGetUserLogic_GetUser tests the GetUser logic
func TestGetUserLogic_GetUser(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GetUserReq
		mockSetup func(*gomock.Controller) *svc.ServiceContext
		want      *types.GetUserResp
		wantErr   bool
	}{
		{
			name: "success - 正常获取用户",
			req: &types.GetUserReq{
				Id: 1,
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				mockUserRpc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(&userv1.GetUserResponse{
						User: &userv1.User{
							Id:       1,
							Nickname: "测试用户",
							Phone:    "13800138000",
						},
					}, nil)

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			want: &types.GetUserResp{
				User: types.UserInfo{
					Id:       1,
					Nickname: "测试用户",
					Phone:    "13800138000",
				},
			},
			wantErr: false,
		},
		{
			name: "error - 用户不存在",
			req: &types.GetUserReq{
				Id: 999,
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				// Mock: UserRpc.GetUser 返回"用户不存在"错误
				mockUserRpc.EXPECT().
					GetUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("用户不存在"))

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			wantErr: true,
		},
		{
			name: "error - ID 为零",
			req: &types.GetUserReq{
				Id: 0,
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				return &svc.ServiceContext{}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svcCtx := tt.mockSetup(ctrl)
			l := NewGetUserLogic(context.Background(), svcCtx)

			// Execute
			got, err := l.GetUser(tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			if tt.want != nil {
				assert.Equal(t, tt.want.User.Id, got.User.Id)
				assert.Equal(t, tt.want.User.Nickname, got.User.Nickname)
			}
		})
	}
}

// TestUpdateUserLogic_UpdateUser tests the UpdateUser logic
func TestUpdateUserLogic_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.UpdateUserReq
		mockSetup func(*gomock.Controller) *svc.ServiceContext
		want      *types.UpdateUserResp
		wantErr   bool
	}{
		{
			name: "success - 更新昵称",
			req: &types.UpdateUserReq{
				Id:       1,
				Nickname: strPtr("新昵称"),
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				mockUserRpc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(&userv1.UpdateUserResponse{
						User: &userv1.User{
							Id:       1,
							Nickname: "新昵称",
						},
					}, nil)

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			want: &types.UpdateUserResp{
				User: types.UserInfo{
					Id:       1,
					Nickname: "新昵称",
				},
			},
			wantErr: false,
		},
		{
			name: "error - 用户不存在",
			req: &types.UpdateUserReq{
				Id:       999,
				Nickname: strPtr("新昵称"),
			},
			mockSetup: func(ctrl *gomock.Controller) *svc.ServiceContext {
				mockUserRpc := usermocks.NewMockUserServiceClient(ctrl)

				mockUserRpc.EXPECT().
					UpdateUser(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("用户不存在"))

				return &svc.ServiceContext{
					UserRpc: mockUserRpc,
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svcCtx := tt.mockSetup(ctrl)
			l := NewUpdateUserLogic(context.Background(), svcCtx)

			// Execute
			got, err := l.UpdateUser(tt.req)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, got)
			if tt.want != nil {
				assert.Equal(t, tt.want.User.Id, got.User.Id)
			}
		})
	}
}

// Helper functions
func strPtr(s string) *string {
	return &s
}
