// @author - deltartificial
package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"

	token "mevcopytrader/contracts_erc20"
)

var (
	WssRpcNode                  = ""
	SandwichBotEOAAddress       = common.HexToAddress("0xae2fc483527b8ef99eb5d9b44875f005ba1fae13")
	SandwitchBotContractAddress = common.HexToAddress("0x6b75d8AF000000e20B7a7DDf000Ba900b4009A80")
	transferFnSignature         = crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
	callOpts                    = &bind.CallOpts{}
	tokenCache                  = make(map[common.Address]string)
	mu                          sync.Mutex
	cyan                        = color.New(color.FgCyan, color.Bold).SprintFunc()
	green                       = color.New(color.FgGreen, color.Bold).SprintFunc()
	yellow                      = color.New(color.FgYellow, color.Bold).SprintFunc()
)

const workerCount = 10

func main() {
	client, err := ethclient.Dial(WssRpcNode)
	if err != nil {
		log.Fatal(err)
	}

	headers := make(chan *types.Header, 10)
	sub, err := client.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Fatal(err)
	}

	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	blocks := make(chan *types.Block, 100)
	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go worker(client, chainID, SandwichBotEOAAddress, blocks, &wg)
	}

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case header := <-headers:
			block, err := client.BlockByHash(context.Background(), header.Hash())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("%s Block Time: %d, Block Number: %d, Block Hash: %s\n", cyan("[INFO]"), block.Time(), block.Number().Uint64(), block.Hash().Hex())

			blocks <- block
		}
	}
}

func worker(client *ethclient.Client, chainID *big.Int, SandwichBotEOAAddress common.Address, blocks <-chan *types.Block, wg *sync.WaitGroup) {
	defer wg.Done()
	for block := range blocks {
		for _, tx := range block.Transactions() {
			processTransaction(client, chainID, SandwichBotEOAAddress, tx, block.Number())
		}
	}
}

func processTransaction(client *ethclient.Client, chainID *big.Int, SandwichBotEOAAddress common.Address, tx *types.Transaction, blockNumber *big.Int) {
	signer := types.NewLondonSigner(chainID)
	sender, err := types.Sender(signer, tx)
	if err != nil {
		log.Fatal("3", err)
	}
	if sender == SandwichBotEOAAddress {
		receipt, err := client.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			log.Fatal(err)
		}

		for _, vLog := range receipt.Logs {
			if vLog.Topics[0] == transferFnSignature {
				eventAddress := vLog.Address

				tokenNameAddress, err := getTokenName(client, eventAddress)
				if err != nil {
					fmt.Println(err)
				}

				fmt.Printf("%s (ERC20) - Address: %s (%v)\n", yellow("[>]"), eventAddress, yellow(tokenNameAddress))

				blockDetermined := new(big.Int).Sub(blockNumber, big.NewInt(3))

				balanceAtBlock, err := getTokenBalanceAtBlock(client, eventAddress, SandwitchBotContractAddress, blockDetermined)
				if err != nil {
					log.Fatal(err)
				}

				if balanceAtBlock.Cmp(big.NewInt(0)) == 0 {
					fmt.Printf("%s NEW ERC-20 BUY at block %v : %s (%v)\n", green("[FOUND]"), blockDetermined, eventAddress, tokenNameAddress)
				}
			}
		}
	}
}

func getTokenName(client *ethclient.Client, tokenAddress common.Address) (string, error) {
	mu.Lock()
	if name, ok := tokenCache[tokenAddress]; ok {
		mu.Unlock()
		return name, nil
	}
	mu.Unlock()

	instance, err := token.NewToken(tokenAddress, client)
	if err != nil {
		return "", err
	}

	name, err := instance.Name(callOpts)
	if err != nil {
		return "", err
	}

	mu.Lock()
	tokenCache[tokenAddress] = name
	mu.Unlock()

	return name, nil
}

func getTokenBalanceAtBlock(client *ethclient.Client, tokenAddress common.Address, addressToCheck common.Address, blockNumber *big.Int) (*big.Int, error) {
	instance, err := token.NewToken(tokenAddress, client)
	if err != nil {
		return nil, err
	}

	callOpts.BlockNumber = blockNumber

	balance, err := instance.BalanceOf(callOpts, addressToCheck)
	if err != nil {
		return nil, err
	}

	return balance, nil
}
