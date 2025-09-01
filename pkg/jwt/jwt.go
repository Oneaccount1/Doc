package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims JWT 声明结构
// 包含用户基本信息和标准 JWT 字段
type Claims struct {
	UserID               int64  `json:"user_id"`  // 用户ID
	Username             string `json:"username"` // 用户名
	Email                string `json:"email"`    // 邮箱
	jwt.RegisteredClaims        // 标准 JWT 字段
}

// JWTManager JWT 管理器
// 负责 JWT token 的生成、验证和解析
type JWTManager struct {
	secretKey     []byte        // JWT 签名密钥
	tokenDuration time.Duration // Token 有效期
}

// NewJWTManager 创建新的 JWT 管理器
// secretKey: JWT 签名密钥，用于签名和验证 token
// tokenDuration: token 有效期，过期后需要重新登录
func NewJWTManager(secretKey string, tokenDuration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		tokenDuration: tokenDuration,
	}
}

// GenerateToken 生成 JWT token
// 根据用户信息生成包含用户身份的 JWT token
// 参数说明：
// - userID: 用户唯一标识
// - username: 用户名
// - email: 用户邮箱
// 返回值：
// - string: 生成的 JWT token 字符串
// - error: 生成过程中的错误
func (j *JWTManager) GenerateToken(userID int64, username, email string) (string, error) {
	// 设置 token 过期时间
	expirationTime := time.Now().Add(j.tokenDuration)

	// 创建 JWT 声明
	// 包含用户信息和标准的 JWT 字段
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), // 过期时间
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // 签发时间
			NotBefore: jwt.NewNumericDate(time.Now()),     // 生效时间
			Issuer:    "InkwaveDocNet",                    // 签发者
			Subject:   "user_auth",                        // 主题
		},
	}

	// 使用 HS256 算法创建 token
	// HS256 是对称加密算法，使用相同的密钥进行签名和验证
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 使用密钥签名 token
	tokenString, err := token.SignedString(j.secretKey)

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证 JWT token
// 验证 token 的有效性并解析出用户信息
// 参数说明：
// - tokenString: 待验证的 JWT token 字符串
// 返回值：
// - *Claims: 解析出的用户声明信息
// - error: 验证过程中的错误
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	// 解析 token
	// 使用密钥验证 token 的签名
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		// 确保使用的是预期的 HMAC 算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// 检查 token 是否有效
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 提取声明信息
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// RefreshToken 刷新 JWT token
// 为即将过期的 token 生成新的 token
// 参数说明：
// - tokenString: 当前的 JWT token
// 返回值：
// - string: 新的 JWT token
// - error: 刷新过程中的错误
func (j *JWTManager) RefreshToken(tokenString string) (string, error) {
	// 首先验证当前 token
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// 检查 token 是否即将过期（在最后 30 分钟内）
	// 只有即将过期的 token 才允许刷新，防止滥用
	if time.Until(claims.ExpiresAt.Time) > 30*time.Minute {
		return "", errors.New("token is not eligible for refresh")
	}

	// 使用原有的用户信息生成新 token
	return j.GenerateToken(claims.UserID, claims.Username, claims.Email)
}

// ExtractUserID 从 token 中提取用户ID
// 便捷方法，用于快速获取当前用户ID
func (j *JWTManager) ExtractUserID(tokenString string) (int64, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}

// ExtractUserInfo 从 token 中提取用户信息
// 便捷方法，用于获取当前用户的基本信息
func (j *JWTManager) ExtractUserInfo(tokenString string) (int64, string, string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return 0, "", "", err
	}
	return claims.UserID, claims.Username, claims.Email, nil
}
