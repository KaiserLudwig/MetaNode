package ethcli

// 合约交互：用 go-ethereum 的 abi 包封装对 ERC-20 / ERC-721 / Uniswap 合约的调用。
//
// 核心思路：
//   1) 用 ABI（应用二进制接口）描述合约的函数签名与返回类型；
//   2) 把「函数名 + 参数」编码成 calldata（十六进制字节）；
//   3) 调用 eth_call 只读方法（不消耗 gas，不改变状态），拿回并解码返回值；
//   4) 对需要写状态的调用（mint/swap），则构造交易、签名并广播。
//
// 这里提供标准 ERC-20、ERC-721 与 Uniswap V2 池的最常用接口（只读部分）。

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// 常用 ERC-20 接口（name/symbol/decimals/balanceOf/totalSupply）。
var erc20ABI = mustABI(`[
	{"type":"function","name":"name","stateMutability":"view","inputs":[],"outputs":[{"type":"string"}]},
	{"type":"function","name":"symbol","stateMutability":"view","inputs":[],"outputs":[{"type":"string"}]},
	{"type":"function","name":"decimals","stateMutability":"view","inputs":[],"outputs":[{"type":"uint8"}]},
	{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"type":"address"}],"outputs":[{"type":"uint256"}]},
	{"type":"function","name":"totalSupply","stateMutability":"view","inputs":[],"outputs":[{"type":"uint256"}]}
]`)

// 常用 ERC-721 接口（name/symbol/totalSupply/ownerOf/balanceOf/tokenURI）。
var erc721ABI = mustABI(`[
	{"type":"function","name":"name","stateMutability":"view","inputs":[],"outputs":[{"type":"string"}]},
	{"type":"function","name":"symbol","stateMutability":"view","inputs":[],"outputs":[{"type":"string"}]},
	{"type":"function","name":"totalSupply","stateMutability":"view","inputs":[],"outputs":[{"type":"uint256"}]},
	{"type":"function","name":"ownerOf","stateMutability":"view","inputs":[{"type":"uint256"}],"outputs":[{"type":"address"}]},
	{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"type":"address"}],"outputs":[{"type":"uint256"}]},
	{"type":"function","name":"tokenURI","stateMutability":"view","inputs":[{"type":"uint256"}],"outputs":[{"type":"string"}]}
]`)

// Uniswap V2 池接口（token0/token1/decimals?/getReserves/balanceOf）。
var uniswapPairABI = mustABI(`[
	{"type":"function","name":"token0","stateMutability":"view","inputs":[],"outputs":[{"type":"address"}]},
	{"type":"function","name":"token1","stateMutability":"view","inputs":[],"outputs":[{"type":"address"}]},
	{"type":"function","name":"balanceOf","stateMutability":"view","inputs":[{"type":"address"}],"outputs":[{"type":"uint256"}]},
	{"type":"function","name":"totalSupply","stateMutability":"view","inputs":[],"outputs":[{"type":"uint256"}]},
	{"type":"function","name":"getReserves","stateMutability":"view","inputs":[],"outputs":[
		{"type":"uint112"},{"type":"uint112"},{"type":"uint32"}]}
]`)

func mustABI(js string) abi.ABI {
	a, err := abi.JSON(strings.NewReader(js))
	if err != nil {
		panic(err)
	}
	return a
}

// call 执行一次只读 eth_call，返回解码后的返回值列表（无输出/单输出/多输出通用）。
//
// go-ethereum v1.15.0 的 abi.Unpack(name, data) 直接返回 []interface{}，
// 调用方需按 ABI 声明顺序对 out[0]、out[1]... 做类型断言。
func (c *Client) call(to common.Address, abiObj abi.ABI, method string, args ...interface{}) ([]interface{}, error) {
	data, err := abiObj.Pack(method, args...)
	if err != nil {
		return nil, fmt.Errorf("编码调用失败: %w", err)
	}
	msg := ethereum.CallMsg{To: &to, Data: data}
	res, err := c.rpc.CallContract(c.ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("eth_call %s 失败: %w", method, err)
	}
	return abiObj.Unpack(method, res)
}

// ---------- ERC-20 ----------

// ERC20Name 查询代币名称。
func (c *Client) ERC20Name(token common.Address) (string, error) {
	out, err := c.call(token, erc20ABI, "name")
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// ERC20Symbol 查询代币符号。
func (c *Client) ERC20Symbol(token common.Address) (string, error) {
	out, err := c.call(token, erc20ABI, "symbol")
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// ERC20Decimals 查询代币精度（decimals，如 USDC=6, ETH/WETH=18）。
func (c *Client) ERC20Decimals(token common.Address) (uint8, error) {
	out, err := c.call(token, erc20ABI, "decimals")
	if err != nil {
		return 0, err
	}
	return out[0].(uint8), nil
}

// ERC20BalanceOf 查询某账户持有某 ERC-20 代币的数量（wei 级别的最小单位）。
func (c *Client) ERC20BalanceOf(token, owner common.Address) (*big.Int, error) {
	out, err := c.call(token, erc20ABI, "balanceOf", owner)
	if err != nil {
		return nil, err
	}
	return out[0].(*big.Int), nil
}

// ERC20TotalSupply 查询代币总供应量。
func (c *Client) ERC20TotalSupply(token common.Address) (*big.Int, error) {
	out, err := c.call(token, erc20ABI, "totalSupply")
	if err != nil {
		return nil, err
	}
	return out[0].(*big.Int), nil
}

// ---------- ERC-721 (NFT) ----------

// NFTName 查询 NFT 项目名称。
func (c *Client) NFTName(contract common.Address) (string, error) {
	out, err := c.call(contract, erc721ABI, "name")
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// NFTSymbol 查询 NFT 项目符号。
func (c *Client) NFTSymbol(contract common.Address) (string, error) {
	out, err := c.call(contract, erc721ABI, "symbol")
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// NFTOwnerOf 查询某个 tokenId 的当前所有者地址。
func (c *Client) NFTOwnerOf(contract common.Address, tokenID *big.Int) (common.Address, error) {
	out, err := c.call(contract, erc721ABI, "ownerOf", tokenID)
	if err != nil {
		return common.Address{}, err
	}
	return out[0].(common.Address), nil
}

// NFTBalanceOf 查询某地址持有该 NFT 项目代币的数量。
func (c *Client) NFTBalanceOf(contract, owner common.Address) (*big.Int, error) {
	out, err := c.call(contract, erc721ABI, "balanceOf", owner)
	if err != nil {
		return nil, err
	}
	return out[0].(*big.Int), nil
}

// NFTTokenURI 查询某个 tokenId 的元数据 URI（通常指向 JSON 元数据）。
func (c *Client) NFTTokenURI(contract common.Address, tokenID *big.Int) (string, error) {
	out, err := c.call(contract, erc721ABI, "tokenURI", tokenID)
	if err != nil {
		return "", err
	}
	return out[0].(string), nil
}

// ---------- Uniswap V2 池 ----------

// PoolToken0 返回池子中的代币0地址。
func (c *Client) PoolToken0(pool common.Address) (common.Address, error) {
	out, err := c.call(pool, uniswapPairABI, "token0")
	if err != nil {
		return common.Address{}, err
	}
	return out[0].(common.Address), nil
}

// PoolToken1 返回池子中的代币1地址。
func (c *Client) PoolToken1(pool common.Address) (common.Address, error) {
	out, err := c.call(pool, uniswapPairABI, "token1")
	if err != nil {
		return common.Address{}, err
	}
	return out[0].(common.Address), nil
}

// PoolReserves 返回池子的流动性储备 (reserve0, reserve1)。这是 DEX 定价的基础。
func (c *Client) PoolReserves(pool common.Address) (r0, r1 *big.Int, err error) {
	out, err := c.call(pool, uniswapPairABI, "getReserves")
	if err != nil {
		return nil, nil, err
	}
	reserve0 := out[0].(*big.Int)
	reserve1 := out[1].(*big.Int)
	return reserve0, reserve1, nil
}

// PoolTotalSupply 返回流动性代币（LP token）总供应量。
func (c *Client) PoolTotalSupply(pool common.Address) (*big.Int, error) {
	out, err := c.call(pool, uniswapPairABI, "totalSupply")
	if err != nil {
		return nil, err
	}
	return out[0].(*big.Int), nil
}
