package blockchain

import (
	"testing"
)

// TestGenesisValidity 创世块应满足难度且链自洽。
func TestGenesisValidity(t *testing.T) {
	bc := NewBlockchain(2, 50)
	if !bc.IsValid() {
		t.Fatal("创世链应自洽")
	}
	if len(bc.GetBlocks()) != 1 {
		t.Fatalf("应只有创世块，实为 %d", len(bc.GetBlocks()))
	}
}

// TestMiningAndChainGrowth 挖矿后区块数应增长，难度与哈希都应满足。
func TestMiningAndChainGrowth(t *testing.T) {
	bc := NewBlockchain(2, 50)
	before := len(bc.GetBlocks())

	reward := NewCoinbaseTx("miner", bc.GetCoinbaseReward())
	block, err := bc.MineBlock([]*Transaction{reward})
	if err != nil {
		t.Fatalf("挖矿失败: %v", err)
	}
	if len(bc.GetBlocks()) != before+1 {
		t.Fatalf("区块数应为 %d，实为 %d", before+1, len(bc.GetBlocks()))
	}
	if !HashMatchesDifficulty(block.Hash, bc.GetDifficulty()) {
		t.Fatalf("区块哈希 %s 不满足难度", block.Hash)
	}
	if !bc.IsValid() {
		t.Fatal("增加区块后链应仍自洽")
	}
}

// TestTransferAndBalances 转账后余额应正确，且支持找零。
func TestTransferAndBalances(t *testing.T) {
	bc := NewBlockchain(2, 50)
	alice := NewWallet()
	bob := NewWallet()

	// 先给 Alice 发奖励，使其有钱。
	if _, err := bc.MineBlock([]*Transaction{NewCoinbaseTx(alice.Address(), 50)}); err != nil {
		t.Fatal(err)
	}
	if got := bc.BalanceOf(alice.Address()); got != 50 {
		t.Fatalf("Alice 初始余额应为 50，实为 %d", got)
	}

	// Alice 转 30 给 Bob，输出应有找零 20。
	tx, err := bc.SignAndBuildTransfer(alice, bob.Address(), 30)
	if err != nil {
		t.Fatalf("构建转账失败: %v", err)
	}
	if _, err := bc.MineBlock([]*Transaction{tx}); err != nil {
		t.Fatalf("打包转账失败: %v", err)
	}

	if got := bc.BalanceOf(alice.Address()); got != 20 {
		t.Fatalf("Alice 余额应为 20（找零），实为 %d", got)
	}
	if got := bc.BalanceOf(bob.Address()); got != 30 {
		t.Fatalf("Bob 余额应为 30，实为 %d", got)
	}
	if !bc.IsValid() {
		t.Fatal("转账后链应自洽")
	}
}

// TestDoubleSpendRejected 引用已经花费的输出（真正双花）应被校验拦截。
func TestDoubleSpendRejected(t *testing.T) {
	bc := NewBlockchain(2, 50)
	alice := NewWallet()
	bob := NewWallet()

	if _, err := bc.MineBlock([]*Transaction{NewCoinbaseTx(alice.Address(), 50)}); err != nil {
		t.Fatal(err)
	}
	// Alice 转 30 给 Bob，花掉了 coinbase 的 50 输出。
	tx1, err := bc.SignAndBuildTransfer(alice, bob.Address(), 30)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := bc.MineBlock([]*Transaction{tx1}); err != nil {
		t.Fatalf("第一次转账应成功: %v", err)
	}

	// 现在手工构造一笔「恶意交易」：其输入直接引用已被 tx1 花费的同一输出
	// （即 tx1 的输入引用的 coinbase 输出），从而构成真正的双花。
	spentOp := tx1.Inputs[0].Prev // 这是 coinbase 输出，已在 tx1 中花费
	mallTx := &Transaction{
		Inputs: []TXInput{{
			Prev:       spentOp,
			Signature:  []byte("forged"),
			PubKey:     alice.PublicKeyBytes(),
			PubKeyHash: hashPubKey([]byte(alice.Address())),
		}},
		Outputs: []TXOutput{{Value: 50, PubKeyHash: hashPubKey([]byte("evil"))}},
	}
	mallTx.ID = mallTx.CalculateID()

	if _, err := bc.MineBlock([]*Transaction{mallTx}); err == nil {
		t.Fatal("引用已花费输出的双花交易应被拦截")
	}
}

// TestTamperDetected 篡改已上链区块内容应导致校验失败。
func TestTamperDetected(t *testing.T) {
	bc := NewBlockchain(2, 50)
	if _, err := bc.MineBlock([]*Transaction{NewCoinbaseTx("m", 50)}); err != nil {
		t.Fatal(err)
	}
	blocks := bc.GetBlocks()
	last := blocks[len(blocks)-1]
	// 把区块里交易输出金额改大（模拟恶意篡改）。
	if len(last.Transactions) > 0 {
		last.Transactions[0].Outputs[0].Value = 999999
	}
	// 注意：这里用原始的 prevHash（来自链上前一块）重新校验，
	// 由于金额变了 -> CalculateID 变 -> 区块哈希变 -> 校验应失败。
	err := last.Validate(last.PrevHash)
	if err == nil {
		t.Fatal("篡改后的区块不应通过校验")
	}
}

// TestSignVerifyRoundtrip 签名/验签应往返一致，且错误消息不通过。
func TestSignVerifyRoundtrip(t *testing.T) {
	w := NewWallet()
	pubBytes := w.PublicKeyBytes()
	pub, err := GetPubKeyFromBytes(pubBytes)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256sumBytes([]byte("hello"))
	sig := signECDSA(w.PrivateKey, hash)
	if !verifyECDSA(pub, hash, sig) {
		t.Fatal("合法签名应通过验证")
	}
	// 篡改消息后应失败。
	if verifyECDSA(pub, sha256sumBytes([]byte("tampered")), sig) {
		t.Fatal("篡改后的消息不应通过验证")
	}
}

// 辅助：对字节做 SHA-256。
func sha256sumBytes(b []byte) []byte {
	s := sha256Sum(b)
	return s
}

// 避免与其它文件同名的辅助命名冲突：直接内联。
