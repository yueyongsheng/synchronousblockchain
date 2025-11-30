package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	// RPC 节点地址 - 可选择以下任一个：
	// 1. Infura: https://sepolia.infura.io/v3/3866392dfd05409abfc3e5da3960669d
	// 2. Alchemy: https://eth-sepolia.g.alchemy.com/v2/0EbsdN3a3nq5h1tArVk_8
	// 3. 公共节点（免费，但可能不稳定）: https://rpc.sepolia.org
	// 4. 其他公共节点: https://ethereum-sepolia-rpc.publicnode.com
	rpcURL = "https://sepolia.infura.io/v3/3866392dfd05409abfc3e5da3960669d"

	// 钱包私钥
	privateKeyHex = "f8dc09bb7e17795fbbe03ce9b39f7a69c742f1e2fccc0db374e2c03c3ec7b18f"

	// 接收方地址
	toAddress = "0x489C6e2f86d21F84B5207520D070B12573F739F5"
)

func main() {
	// 连接到 Sepolia 测试网络
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接到以太坊节点失败: %v", err)
	}
	defer client.Close()
	fmt.Println("✓ 成功连接到 Sepolia 测试网络")

	// 1. 查询区块信息
	fmt.Println("\n========== 查询区块信息 ==========")
	queryBlockInfo(client, 9735711)

	// 2. 发送交易
	fmt.Println("\n========== 发送交易 ==========")
	sendTransaction(client)
}

// queryBlockInfo 查询区块信息
// 参数 targetBlockNumber: 传入 0 查询最新区块，传入具体数字查询指定区块
func queryBlockInfo(client *ethclient.Client, targetBlockNumber int64) {
	ctx := context.Background()

	// 获取最新区块号
	latestBlockNumber, err := client.BlockNumber(ctx)
	if err != nil {
		log.Fatalf("获取最新区块号失败: %v", err)
	}
	fmt.Printf("最新区块号: %d\n", latestBlockNumber)

	// 确定要查询的区块号
	var queryBlockNum *big.Int
	if targetBlockNumber <= 0 {
		// 查询最新区块
		queryBlockNum = nil // nil 表示最新区块
		fmt.Println("正在查询: 最新区块")
	} else {
		// 查询指定区块
		queryBlockNum = big.NewInt(targetBlockNumber)
		fmt.Printf("正在查询: 指定区块 #%d\n", targetBlockNumber)
	}

	// 查询最新区块的详细信息
	// block, err := client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	block, err := client.BlockByNumber(ctx, queryBlockNum)
	if err != nil {
		log.Fatalf("获取区块信息失败: %v", err)
	}

	// 输出区块信息
	fmt.Println("\n--- 区块详细信息 ---")
	fmt.Printf("区块号:       %d\n", block.Number().Uint64())
	fmt.Printf("区块哈希:     %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希:   %s\n", block.ParentHash().Hex())
	fmt.Printf("时间戳:       %d\n", block.Time())
	fmt.Printf("交易数量:     %d\n", len(block.Transactions()))
	fmt.Printf("Gas 使用量:   %d\n", block.GasUsed())
	fmt.Printf("Gas 限制:     %d\n", block.GasLimit())
	fmt.Printf("矿工地址:     %s\n", block.Coinbase().Hex())

	// 如果区块中有交易，显示第一笔交易的信息
	if len(block.Transactions()) > 0 {
		fmt.Println("\n--- 区块中第一笔交易 ---")
		tx := block.Transactions()[0]
		fmt.Printf("交易哈希:     %s\n", tx.Hash().Hex())
		fmt.Printf("Gas 价格:     %s\n", tx.GasPrice().String())
		fmt.Printf("Gas 限制:     %d\n", tx.Gas())
		fmt.Printf("转账金额:     %s Wei\n", tx.Value().String())
	}
}

// sendTransaction 发送以太币转账交易
func sendTransaction(client *ethclient.Client) {
	ctx := context.Background()

	// 1. 加载私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败: %v\n请确保私钥格式正确（不要包含 0x 前缀）", err)
	}
	fmt.Println("✓ 私钥加载成功")

	// 2. 从私钥获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("获取公钥失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("发送方地址: %s\n", fromAddress.Hex())

	// 3. 获取账户余额
	balance, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		log.Fatalf("获取余额失败: %v", err)
	}
	fmt.Printf("账户余额:   %s Wei (约 %s ETH)\n", balance.String(), weiToEth(balance))

	// 检查余额是否足够
	if balance.Cmp(big.NewInt(0)) == 0 {
		log.Fatal("账户余额为 0")
	}

	// 4. 获取 nonce（交易序号）
	nonce, err := client.PendingNonceAt(ctx, fromAddress)
	if err != nil {
		log.Fatalf("获取 nonce 失败: %v", err)
	}
	fmt.Printf("当前 Nonce: %d\n", nonce)

	// 5. 设置交易参数
	value := big.NewInt(1000000000000000) // 0.001 ETH = 10^15 Wei
	gasLimit := uint64(21000)             // 标准转账 gas 限制

	// 获取建议的 gas 价格
	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		log.Fatalf("获取 gas 价格失败: %v", err)
	}
	fmt.Printf("Gas 价格:   %s Wei\n", gasPrice.String())

	// 6. 创建交易
	to := common.HexToAddress(toAddress)
	tx := types.NewTransaction(nonce, to, value, gasLimit, gasPrice, nil)

	// 7. 获取链 ID
	chainID, err := client.NetworkID(ctx)
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}
	fmt.Printf("链 ID:      %s\n", chainID.String())

	// 8. 签名交易
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		log.Fatalf("签名交易失败: %v", err)
	}
	fmt.Println("✓ 交易签名成功")

	// 9. 发送交易
	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	// 10. 输出交易哈希
	fmt.Println("\n========== 交易发送成功 ==========")
	fmt.Printf("交易哈希: %s\n", signedTx.Hash().Hex())
	fmt.Printf("转账金额: %s ETH\n", weiToEth(value))
	fmt.Printf("接收地址: %s\n", toAddress)
}

// weiToEth 将 Wei 转换为 ETH 字符串
func weiToEth(wei *big.Int) string {
	eth := new(big.Float).SetInt(wei)
	eth.Quo(eth, big.NewFloat(1e18))
	return eth.Text('f', 6)
}
