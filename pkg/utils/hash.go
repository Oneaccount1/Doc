package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword 使用 bcrypt 哈希密码
// bcrypt 是一种安全的密码哈希算法，具有以下特点：
// 1. 自适应：可以调整计算复杂度
// 2. 加盐：自动生成随机盐值
// 3. 慢速：故意设计为计算缓慢，防止暴力破解
// 参数说明：
// - password: 明文密码
// 返回值：
// - string: 哈希后的密码
// - error: 哈希过程中的错误
func HashPassword(password string) (string, error) {
	// 使用 bcrypt.DefaultCost (10) 作为成本参数
	// 成本越高，计算越慢，安全性越高，但用户体验越差
	// DefaultCost 在安全性和性能之间取得了良好平衡
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashedBytes), nil
}

// CheckPassword 验证密码是否正确
// 将明文密码与存储的哈希值进行比较
// 参数说明：
// - password: 用户输入的明文密码
// - hashedPassword: 数据库中存储的哈希密码
// 返回值：
// - bool: 密码是否匹配
func CheckPassword(password, hashedPassword string) bool {
	// bcrypt.CompareHashAndPassword 会自动处理盐值
	// 如果密码匹配返回 nil，不匹配返回错误
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// GenerateRandomCode 生成指定长度的随机数字验证码
// 用于邮箱验证、短信验证等场景
// 参数说明：
// - length: 验证码长度
// 返回值：
// - string: 生成的验证码
// - error: 生成过程中的错误
func GenerateRandomCode(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	// 定义数字字符集
	const digits = "0123456789"

	// 使用 crypto/rand 生成安全的随机数
	// 比 math/rand 更安全，适用于密码学场景
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		// 生成 0 到 len(digits)-1 的随机数
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random number: %w", err)
		}
		result[i] = digits[num.Int64()]
	}

	return string(result), nil
}

// GenerateRandomString 生成指定长度的随机字符串
// 包含大小写字母和数字，用于生成随机用户名、临时密码等
// 参数说明：
// - length: 字符串长度
// 返回值：
// - string: 生成的随机字符串
// - error: 生成过程中的错误
func GenerateRandomString(length int) (string, error) {
	if length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}

	// 定义字符集：大小写字母和数字
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random string: %w", err)
		}
		result[i] = charset[num.Int64()]
	}

	return string(result), nil
}

// MD5Hash 计算字符串的 MD5 哈希值
// 主要用于生成缓存键、文件指纹等非安全场景
// 注意：MD5 不适用于密码哈希，仅用于数据完整性检查
// 参数说明：
// - text: 待哈希的字符串
// 返回值：
// - string: MD5 哈希值（32位十六进制字符串）
func MD5Hash(text string) string {
	// 创建 MD5 哈希器
	hasher := md5.New()

	// 写入数据
	hasher.Write([]byte(text))

	// 计算哈希值并转换为十六进制字符串
	return hex.EncodeToString(hasher.Sum(nil))
}

// GenerateRedisKey 生成 Redis 缓存键
// 使用 MD5 哈希确保键名的一致性和安全性
// 参数说明：
// - prefix: 键前缀，用于区分不同类型的缓存
// - identifier: 标识符，通常是用户邮箱、ID等
// 返回值：
// - string: 生成的 Redis 键名
func GenerateRedisKey(prefix, identifier string) string {
	// 对标识符进行 MD5 哈希，避免特殊字符和长度问题
	hashedIdentifier := MD5Hash(identifier)
	return fmt.Sprintf("%s:%s", prefix, hashedIdentifier)
}

// ValidatePasswordStrength 验证密码强度
// 检查密码是否符合安全要求
// 参数说明：
// - password: 待验证的密码
// 返回值：
// - bool: 密码是否符合强度要求
// - string: 不符合要求时的错误信息
func ValidatePasswordStrength(password string) (bool, string) {
	// 检查密码长度
	if len(password) < 8 {
		return false, "密码长度至少为8位"
	}

	if len(password) > 128 {
		return false, "密码长度不能超过128位"
	}

	// 检查是否包含数字
	hasDigit := false
	// 检查是否包含字母
	hasLetter := false
	// 检查是否包含特殊字符
	hasSpecial := false

	for _, char := range password {
		switch {
		case char >= '0' && char <= '9':
			hasDigit = true
		case (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z'):
			hasLetter = true
		case char >= 33 && char <= 126: // 可打印的特殊字符
			if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z')) {
				hasSpecial = true
			}
		}
	}

	// 至少包含两种类型的字符
	typeCount := 0
	if hasDigit {
		typeCount++
	}
	if hasLetter {
		typeCount++
	}
	if hasSpecial {
		typeCount++
	}

	if typeCount < 2 {
		return false, "密码必须包含至少两种类型的字符（数字、字母、特殊字符）"
	}

	return true, ""
}
