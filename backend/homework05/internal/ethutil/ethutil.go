// Package ethutil 封装 RPC 连接与私钥加载，供三个命令行程序共用。
package ethutil

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// RPCURL 依次尝试：SEPOLIA_RPC_URL -> 用 INFURA_API_KEY 组装 Infura 地址 -> 公共节点。
func RPCURL() string {
	if v := strings.TrimSpace(os.Getenv("SEPOLIA_RPC_URL")); v != "" {
		return v
	}
	if key := strings.TrimSpace(os.Getenv("INFURA_API_KEY")); key != "" {
		return "https://sepolia.infura.io/v3/" + key
	}
	return "https://ethereum-sepolia-rpc.publicnode.com"
}

// Dial 建立 ethclient 连接，并返回不含凭据的 RPC 提供商标签。
func Dial(ctx context.Context) (*ethclient.Client, string, error) {
	url := RPCURL()
	label := rpcLabel(url)
	client, err := ethclient.DialContext(ctx, url)
	if err != nil {
		return nil, label, fmt.Errorf("连接 %s 失败: %s", label, sanitizeRPCError(err, url))
	}
	return client, label, nil
}

// rpcLabel 只返回 endpoint 的提供商标签或主机名，避免日志输出 URL 凭据。
func rpcLabel(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" {
		return "custom RPC"
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "sepolia.infura.io" {
		return "Infura Sepolia (API key hidden)"
	}
	return host
}

// sanitizeRPCError 保留连接错误信息，同时移除 endpoint 中的路径、查询参数和用户信息。
func sanitizeRPCError(err error, raw string) string {
	message := err.Error()
	parsed, parseErr := url.Parse(raw)
	if parseErr != nil {
		return "RPC 连接失败（endpoint 详情已隐藏）"
	}
	for _, secret := range []string{raw, parsed.String(), parsed.Path, parsed.EscapedPath(), parsed.RawQuery} {
		if secret != "" {
			message = strings.ReplaceAll(message, secret, "[REDACTED]")
		}
	}
	if parsed.User != nil {
		message = strings.ReplaceAll(message, parsed.User.String(), "[REDACTED]")
	}
	return message
}

// PrivateKey 从 PRIVATE_KEY 环境变量解析私钥，并返回对应地址。
func PrivateKey() (*ecdsa.PrivateKey, common.Address, error) {
	raw := strings.TrimSpace(os.Getenv("PRIVATE_KEY"))
	if raw == "" {
		return nil, common.Address{}, errors.New("请先设置 PRIVATE_KEY 环境变量（0x 开头的 64 位十六进制私钥）")
	}
	priv, err := crypto.HexToECDSA(strings.TrimPrefix(raw, "0x"))
	if err != nil {
		return nil, common.Address{}, fmt.Errorf("解析私钥失败: %w", err)
	}
	return priv, crypto.PubkeyToAddress(priv.PublicKey), nil
}

// PublicKey 解析 PRIVATE_KEY 并返回对应的公钥（十六进制字符串，便于打印）。
func PublicKey() (string, error) {
	priv, _, err := PrivateKey()
	if err != nil {
		return "", err
	}
	return crypto.PubkeyToAddress(priv.PublicKey).Hex(), nil
}
