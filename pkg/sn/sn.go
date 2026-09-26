// Package sn 业务编号生成工具。
package sn

import (
	"crypto/rand"
	"errors"
	"math/big"
)

// snAlphabet 编号字符集：小写字母和数字。
const snAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// Generate 生成业务编号：固定小写前缀 cf + 8 位小写字母和数字，如 cfk3m9x2a。
// 唯一性由调用方的唯一索引兜底，冲突时应重新生成重试。
func Generate() (string, error) {
	out := make([]byte, 8)
	for i := range out {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(snAlphabet))))
		if err != nil {
			return "", errors.New("编号生成失败")
		}
		out[i] = snAlphabet[n.Int64()]
	}
	return "cf" + string(out), nil
}
