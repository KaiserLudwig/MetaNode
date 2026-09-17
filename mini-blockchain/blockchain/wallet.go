package blockchain

// ECDSA 钱包：公私钥生成、地址派生、签名与验签。
//
// 现实中的账户地址是这样派生的（以比特币为例）：
//   私钥(32字节) -> SECP256k1 椭圆曲线 -> 公钥(65字节)
//   -> SHA-256(公钥) -> RIPEMD-160 -> 公钥哈希(20字节)
//   -> Base58Check 编码 -> 钱包地址(以 1 开头)
//
// 以太坊则是：
//   私钥(32字节) -> keccak256(公钥去掉前缀) 取后 20 字节 -> 地址(0x + 20字节)
//
// 本教学实现使用 Go 标准库内置的 P256 曲线 + SHA-256，并直接把「公钥哈希」
// 作为地址打印，保留最关键的三个概念：私钥签名、公钥验签、地址 != 私钥。

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// Wallet 保存一个用户的公私钥对。
type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  *ecdsa.PublicKey
}

// NewWallet 生成一个新的公私钥对（SECP256R1，即 P-256 曲线）。
func NewWallet() *Wallet {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	pub := &priv.PublicKey
	return &Wallet{PrivateKey: priv, PublicKey: pub}
}

// PublicKeyBytes 把公钥序列化为未压缩格式 0x04 || x || y。
func (w *Wallet) PublicKeyBytes() []byte {
	return elliptic.Marshal(w.PublicKey.Curve, w.PublicKey.X, w.PublicKey.Y)
}

// Address 派生一个可读地址（本实现用公钥哈希的十六进制表示）。
func (w *Wallet) Address() string {
	return hex.EncodeToString(hashPubKey(w.PublicKeyBytes()))
}

// Sign 使用私钥对数据哈希签名（ECDSA）。返回 (r, s) 的紧凑形式。
func signECDSA(priv *ecdsa.PrivateKey, hash []byte) []byte {
	r, s, err := ecdsa.Sign(rand.Reader, priv, hash)
	if err != nil {
		panic(err)
	}
	sig := make([]byte, 64)
	r.FillBytes(sig[:32])
	s.FillBytes(sig[32:])
	return sig
}

// verifyECDSA 使用公钥验证签名。
func verifyECDSA(pub *ecdsa.PublicKey, hash, sig []byte) bool {
	if len(sig) != 64 {
		return false
	}
	r := new(big.Int).SetBytes(sig[:32])
	s := new(big.Int).SetBytes(sig[32:])
	return ecdsa.Verify(pub, hash, r, s)
}

// GetPubKeyFromBytes 把序列化公钥还原为 *ecdsa.PublicKey。
func GetPubKeyFromBytes(pubBytes []byte) (*ecdsa.PublicKey, error) {
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)
	if x == nil {
		return nil, fmt.Errorf("无效的公钥字节")
	}
	return &ecdsa.PublicKey{Curve: elliptic.P256(), X: x, Y: y}, nil
}

// hashPubKey 对公钥做 SHA-256（教学简化，真实系统还会做 RIPEMD-160）。
func hashPubKey(pub []byte) []byte {
	sum := sha256.Sum256(pub)
	return sum[:]
}

// sha256Sum 返回一段数据的 SHA-256 摘要（供对外签名/校验使用）。
func sha256Sum(b []byte) []byte {
	sum := sha256.Sum256(b)
	return sum[:]
}
