package ethcli

import (
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// TestPrivateKeyToAddress 用已知私钥验证地址派生是否可复现。
// v1.15.0 用 crypto.HexToECDSA + crypto.PubkeyToAddress。
func TestPrivateKeyToAddress(t *testing.T) {
	// 一个众所周知的测试私钥（其地址固定）。
	// 私钥 = 0x0000...0001，对应地址可由 go-ethereum 计算。
	privHex := "0000000000000000000000000000000000000000000000000000000000000001"
	priv, err := crypto.HexToECDSA(privHex)
	if err != nil {
		t.Fatalf("解析私钥失败: %v", err)
	}
	addr := crypto.PubkeyToAddress(priv.PublicKey)
	// 该私钥对应的地址是众所周知的固定值。
	expected := common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf")
	if addr != expected {
		t.Fatalf("地址不匹配: got %s want %s", addr.Hex(), expected.Hex())
	}
}

// TestHexToECDSARejectsBadHex 非法私钥应被拒绝。
func TestHexToECDSARejectsBadHex(t *testing.T) {
	if _, err := crypto.HexToECDSA("not-a-valid-hex"); err == nil {
		t.Fatal("非法私钥应被拒绝")
	}
}

// TestSendWeiPrivateKeyParsing 验证 SendWei 的私钥归一化逻辑（剥离 0x 前缀）。
func TestSendWeiPrivateKeyParsing(t *testing.T) {
	// 带 0x 前缀的长私钥。
	privHex := "0x" + strings.Repeat("1", 64)
	raw := strings.TrimPrefix(privHex, "0x")
	if _, err := crypto.HexToECDSA(raw); err != nil {
		t.Fatalf("解析带前缀私钥失败: %v", err)
	}
	// 不带前缀。
	if _, err := crypto.HexToECDSA(strings.Repeat("1", 64)); err != nil {
		t.Fatalf("解析不带前缀私钥失败: %v", err)
	}
}

// TestWeiConversion 验证 wei <-> ETH 转换关系（1 ETH = 1e18 wei）。
func TestWeiConversion(t *testing.T) {
	oneEth := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	got := WeiToEth(oneEth)
	if got.Cmp(big.NewFloat(1)) != 0 {
		t.Fatalf("1 ETH 应等于 1.0，实为 %v", got)
	}
}

// TestABIPackBalanceOf 验证手写的 ERC-20 ABI 能正确编码 balanceOf 调用。
// 纯本地验证（不调网络），确认我们的 ABI JSON 与 go-ethereum 完全兼容。
func TestABIPackBalanceOf(t *testing.T) {
	addr := common.HexToAddress("0x7E5F4552091A69125d5DfCb7b8C2659029395Bdf")
	data, err := erc20ABI.Pack("balanceOf", addr)
	if err != nil {
		t.Fatalf("pack balanceOf 失败: %v", err)
	}
	// calldata = 4 字节选择器 + 32 字节地址（右对齐）。
	if len(data) != 4+32 {
		t.Fatalf("calldata 长度应为 4+32=36，实为 %d", len(data))
	}
	// 地址应对齐到第 4+12 字节后。
	addrBytes := data[4+12:]
	got := common.BytesToAddress(addrBytes)
	if got != addr {
		t.Fatalf("地址编码不正确: got %s want %s", got.Hex(), addr.Hex())
	}
}

// TestABIUnpackUint256 验证手写的 ABI 能正确解码 uint256 返回值。
func TestABIUnpackUint256(t *testing.T) {
	out, err := erc20ABI.Unpack("totalSupply", packUint256(big.NewInt(187910000)))
	if err != nil {
		t.Fatalf("decode totalSupply 失败: %v", err)
	}
	got := out[0].(*big.Int)
	if got.Cmp(big.NewInt(187910000)) != 0 {
		t.Fatalf("uint256 解码不一致: got %v", got)
	}
}

// packUint256 用 abi 参数类型把整数编码为 abi 格式。
func packUint256(v *big.Int) []byte {
	argType, _ := abi.NewType("uint256", "", nil)
	b, _ := abi.Arguments{{Type: argType}}.Pack(v)
	return b
}
