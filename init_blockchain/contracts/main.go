package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/init_blockchain/counter"
)

const (
	// RPC 节点地址
	rpcURL = "https://sepolia.infura.io/v3/3866392dfd05409abfc3e5da3960669d"

	// 钱包私钥（不要包含 0x 前缀）
	privateKeyHex = "f8dc09bb7e17795fbbe03ce9b39f7a69c742f1e2fccc0db374e2c03c3ec7b18f"
)

func main() {
	// 1. 连接到 Sepolia 测试网络
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		log.Fatalf("连接到以太坊节点失败: %v", err)
	}
	defer client.Close()
	fmt.Println("成功连接到 Sepolia 测试网络")

	// 2. 加载私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("解析私钥失败: %v", err)
	}
	fmt.Println("私钥加载成功")

	// 3. 获取公钥和地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("获取公钥失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("钱包地址: %s\n", fromAddress.Hex())

	// 4. 获取链 ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("获取链 ID 失败: %v", err)
	}
	fmt.Printf("链 ID: %s\n", chainID.String())

	// 5. 创建交易签名器
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatalf("创建交易签名器失败: %v", err)
	}

	// 6. 部署合约
	fmt.Println("\n========== 部署合约 ==========")
	initialCount := big.NewInt(100) // 初始计数器值设为 100
	contractAddress, tx, counterContract, err := counter.DeployCounter(auth, client, initialCount)
	if err != nil {
		log.Fatalf("部署合约失败: %v", err)
	}

	fmt.Printf("合约部署交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("合约地址: %s\n", contractAddress.Hex())
	fmt.Printf("等待交易确认...\n")

	// 7. 等待交易被确认
	fmt.Println("正在等待合约部署确认（可能需要 15-30 秒）...")
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatalf("等待交易确认失败: %v", err)
	}
	fmt.Printf("合约部署成功！区块号: %d\n", receipt.BlockNumber.Uint64())

	// 等待几秒确保状态同步
	time.Sleep(3 * time.Second)

	// 8. 调用合约方法 - 读取当前计数器值
	fmt.Println("\n========== 读取计数器值 ==========")
	currentCount, err := counterContract.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("读取计数器失败: %v", err)
	}
	fmt.Printf("当前计数器值: %s\n", currentCount.String())

	// 9. 调用合约方法 - 增加计数器
	fmt.Println("\n========== 增加计数器 ==========")
	auth, _ = bind.NewKeyedTransactorWithChainID(privateKey, chainID) // 重新创建 auth
	incrementTx, err := counterContract.Increment(auth)
	if err != nil {
		log.Fatalf("调用 increment 失败: %v", err)
	}
	fmt.Printf("increment 交易哈希: %s\n", incrementTx.Hash().Hex())

	// 等待交易确认
	fmt.Println("等待交易确认...")
	receipt, err = bind.WaitMined(context.Background(), client, incrementTx)
	if err != nil {
		log.Fatalf("等待交易确认失败: %v", err)
	}
	fmt.Printf("increment 交易成功！区块号: %d\n", receipt.BlockNumber.Uint64())

	// 10. 再次读取计数器值
	time.Sleep(2 * time.Second)
	newCount, err := counterContract.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("读取计数器失败: %v", err)
	}
	fmt.Printf("增加后的计数器值: %s\n", newCount.String())

	// 11. 设置计数器值
	fmt.Println("\n========== 设置计数器值 ==========")
	auth, _ = bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	setCountTx, err := counterContract.SetCount(auth, big.NewInt(999))
	if err != nil {
		log.Fatalf("调用 setCount 失败: %v", err)
	}
	fmt.Printf("setCount 交易哈希: %s\n", setCountTx.Hash().Hex())

	// 等待交易确认
	fmt.Println("等待交易确认...")
	receipt, err = bind.WaitMined(context.Background(), client, setCountTx)
	if err != nil {
		log.Fatalf("等待交易确认失败: %v", err)
	}
	fmt.Printf("setCount 交易成功！区块号: %d\n", receipt.BlockNumber.Uint64())

	// 12. 最终读取计数器值
	time.Sleep(2 * time.Second)
	finalCount, err := counterContract.GetCount(&bind.CallOpts{})
	if err != nil {
		log.Fatalf("读取计数器失败: %v", err)
	}
	fmt.Printf("设置后的计数器值: %s\n", finalCount.String())
}
