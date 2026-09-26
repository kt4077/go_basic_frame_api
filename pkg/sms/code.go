package sms

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// VerificationCodeLength 系统短信验证码固定长度。
const VerificationCodeLength = 6

// GenerateVerificationCode 使用密码学安全随机源生成固定六位数字验证码，包含前导零。
func GenerateVerificationCode() (string, error) {
	upperBound := new(big.Int).Exp(big.NewInt(10), big.NewInt(VerificationCodeLength), nil)
	value, err := rand.Int(rand.Reader, upperBound)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", VerificationCodeLength, value.Int64()), nil
}
