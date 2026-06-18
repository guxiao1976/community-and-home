package types

// LoginReq 密码登录请求
type LoginReq struct {
	EncryptedPhone    string `json:"encryptedPhone"`
	EncryptedPassword string `json:"encryptedPassword"`
	DeviceId          string `json:"deviceId"`
	DeviceType        string `json:"deviceType,optional"`
}

// LoginResp 登录响应
type LoginResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
	UserId       int64  `json:"userId,string"`
}

// LoginSmsReq 短信验证码登录请求
type LoginSmsReq struct {
	EncryptedPhone string `json:"encryptedPhone"`
	SmsCode        string `json:"smsCode"`
	DeviceId       string `json:"deviceId"`
	DeviceType     string `json:"deviceType,optional"`
}

// RegisterReq 注册请求
type RegisterReq struct {
	EncryptedPhone    string `json:"encryptedPhone"`
	SmsCode           string `json:"smsCode"`
	EncryptedPassword string `json:"encryptedPassword,optional"`
	Nickname          string `json:"nickname"`
	DeviceId          string `json:"deviceId"`
	DeviceType        string `json:"deviceType,optional"`
}

// RegisterResp 注册响应
type RegisterResp struct {
	UserId       int64  `json:"userId,string"`
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// RefreshTokenReq 刷新令牌请求
type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

// RefreshTokenResp 刷新令牌响应
type RefreshTokenResp struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresAt    int64  `json:"expiresAt"`
}

// LogoutReq 注销请求（JWT 认证后，从 Authorization header 提取 access_token）
type LogoutReq struct {
	DeviceId       string `json:"deviceId"`
	KickAllDevices bool   `json:"kickAllDevices,optional"`
}

// SmsSendReq 短信验证码发送请求
type SmsSendReq struct {
	Phone string `json:"phone"`
}

// PublicKeyResp RSA 公钥响应
type PublicKeyResp struct {
	PublicKey string `json:"publicKey"`
}
