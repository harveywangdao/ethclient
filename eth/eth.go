package eth

import (
	"context"
	"crypto/ecdsa"
	//"encoding/hex"
	"errors"
	"ethclient/logger"
	"ethclient/store"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	//"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/ethclient"
	"io/ioutil"
	"math"
	"math/big"
	"os"
)

type EthClient struct {
	cli         *ethclient.Client
	password    string
	keyStoreDir string
}

func (e *EthClient) DeployContract(privateKeyHex string) error {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		logger.Error(err)
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("error casting public key to ECDSA")
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := e.cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	gasPrice, err := e.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	input := "1.0"
	address, tx, instance, err := store.DeployStore(auth, e.cli, input)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(address.Hex())
	logger.Info(tx.Hash().Hex())

	_ = instance
	return nil
}

func (e *EthClient) LoadContract(contractAddrHex string) error {
	address := common.HexToAddress(contractAddrHex)
	instance, err := store.NewStore(address, e.cli)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("contract is loaded")

	version, err := instance.Version(nil)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(version) // "1.0"

	return nil
}

func (e *EthClient) InvokeContract(richPrivKeyHex, contractAddrHex string) error {
	privateKey, err := crypto.HexToECDSA(richPrivKeyHex)
	if err != nil {
		logger.Error(err)
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("error casting public key to ECDSA")
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := e.cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	gasPrice, err := e.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	auth := bind.NewKeyedTransactor(privateKey)
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)     // in wei
	auth.GasLimit = uint64(300000) // in units
	auth.GasPrice = gasPrice

	address := common.HexToAddress(contractAddrHex)
	instance, err := store.NewStore(address, e.cli)
	if err != nil {
		logger.Error(err)
		return err
	}

	key := [32]byte{}
	value := [32]byte{}
	copy(key[:], []byte("foo"))
	copy(value[:], []byte("bar"))

	tx, err := instance.SetItem(auth, key, value)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("tx sent:", tx.Hash().Hex())

	result, err := instance.Items(nil, key)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(string(result[:]))

	return nil
}

func (e *EthClient) ToEthValue(num string) string {
	bigNum := new(big.Float).SetPrec(200)
	bigNum.SetString(num)
	ethValue := new(big.Float).Quo(bigNum, big.NewFloat(math.Pow10(18)))
	return ethValue.Text('g', len(num))
}

func (e *EthClient) GetBalance(addr string) (string, error) {
	account := common.HexToAddress(addr)
	logger.Debug("account :", account)

	balance, err := e.cli.BalanceAt(context.Background(), account, nil)
	if err != nil {
		logger.Error(err)
		return "", err
	}

	logger.Info(addr, "balance :", balance)
	//logger.Debug(e.ToEthValue(balance.String()))

	return balance.String(), nil
}

func (e *EthClient) GetNewWallet() error {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		logger.Error(err)
		return nil
	}

	privateKeyBytes := crypto.FromECDSA(privateKey)
	logger.Info(hexutil.Encode(privateKeyBytes))

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("error casting public key to ECDSA")
		return errors.New("error casting public key to ECDSA")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	logger.Debug(hexutil.Encode(publicKeyBytes))

	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	logger.Info(address)

	//hash := sha3.NewKeccak256()
	//hash.Write(publicKeyBytes[1:])
	//logger.Info(hexutil.Encode(hash.Sum(nil)[12:]))

	return nil
}

func (e *EthClient) CreateKeyStore() error {
	ks := keystore.NewKeyStore(e.keyStoreDir, keystore.StandardScryptN, keystore.StandardScryptP)

	account, err := ks.NewAccount(e.password)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(account.Address.Hex())
	return nil
}

func (e *EthClient) ImportKeyStore(file string) error {
	ks := keystore.NewKeyStore(e.keyStoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	jsonBytes, err := ioutil.ReadFile(file)
	if err != nil {
		logger.Error(err)
		return err
	}

	account, err := ks.Import(jsonBytes, e.password, e.password)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(account.Address.Hex()) // 0x20F8D42FB0F667F2E53930fed426f225752453b3

	if err := os.Remove(file); err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (e *EthClient) CheckAddress(addr string) error {
	/*	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")

		logger.Info("is valid:", re.MatchString("0x323b5d4c32345ced77393b3530b1eed0f346429d")) // is valid: true
		logger.Info("is valid:", re.MatchString("0xZYXb5d4c32345ced77393b3530b1eed0f346429d")) // is valid: false
	*/
	address := common.HexToAddress(addr)
	bytecode, err := e.cli.CodeAt(context.Background(), address, nil) // nil is latest block
	if err != nil {
		logger.Error(err)
		return err
	}

	isContract := len(bytecode) > 0
	logger.Info("is contract:", isContract) // is contract: true

	return nil
}

func (e *EthClient) Header() error {
	header, err := e.cli.HeaderByNumber(context.Background(), nil)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(header.Number.String())

	return nil
}

func (e *EthClient) BlockByNumber(num int64) error {
	blockNumber := big.NewInt(num)
	block, err := e.cli.BlockByNumber(context.Background(), blockNumber)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(block.Number().Uint64())     // 5671744
	logger.Info(block.Time().Uint64())       // 1527211625
	logger.Info(block.Difficulty().Uint64()) // 3217000136609065
	logger.Info(block.Hash().Hex())          // 0x9e8751ebb5069389b855bba72d94902cc385042661498a415979b7b6ee9ba4b9
	logger.Info(len(block.Transactions()))   // 144

	count, err := e.TransactionCount(block.Hash())
	if err != nil {
		logger.Error(err)
		return err
	}

	for _, tx := range block.Transactions() {
		logger.Info(tx.Hash().Hex())        // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2
		logger.Info(tx.Value().String())    // 10000000000000000
		logger.Info(tx.Gas())               // 105000
		logger.Info(tx.GasPrice().Uint64()) // 102000000000
		logger.Info(tx.Nonce())             // 110644
		logger.Info(tx.Data())              // []
		logger.Info(tx.To().Hex())          // 0x55fE59D8Ad77035154dDd0AD0388D09Dd4047A8e

		chainID, err := e.cli.NetworkID(context.Background())
		if err != nil {
			logger.Error(err)
			return err
		}

		if msg, err := tx.AsMessage(types.NewEIP155Signer(chainID)); err == nil {
			logger.Info(msg.From().Hex()) // 0x0fD081e3Bb178dc45c0cb23202069ddA57064258
		}

		receipt, err := e.cli.TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			logger.Error(err)
			return err
		}

		logger.Info(receipt.Status) // 1
	}

	for idx := uint(0); idx < count; idx++ {
		tx, err := e.cli.TransactionInBlock(context.Background(), block.Hash(), idx)
		if err != nil {
			logger.Error(err)
			return err
		}

		logger.Info(tx.Hash().Hex()) // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2

		tx2, isPending, err := e.cli.TransactionByHash(context.Background(), tx.Hash())
		if err != nil {
			logger.Error(err)
			return err
		}

		logger.Info(tx2.Hash().Hex()) // 0x5d49fcaa394c97ec8a9c3e7bd9e8388d420fb050a52083ca52ff24b3b65bc9c2
		logger.Info(isPending)        // false
	}

	return nil
}

func (e *EthClient) TransactionCount(hash common.Hash) (uint, error) {
	count, err := e.cli.TransactionCount(context.Background(), hash)
	if err != nil {
		logger.Error(err)
		return 0, err
	}

	logger.Info(count)

	return count, nil
}

func (e *EthClient) QueryTransaction(txHex string) error {
	txHash := common.HexToHash(txHex)
	tx, isPending, err := e.cli.TransactionByHash(context.Background(), txHash)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info(tx.Hash().Hex())
	logger.Info(tx.Value().String())
	logger.Info(tx.Gas())
	logger.Info(tx.GasPrice().Uint64())
	logger.Info(tx.Nonce())
	logger.Info(tx.Data())
	logger.Info(tx.To().Hex())
	logger.Info(isPending)

	return nil
}

func (e *EthClient) Transfer(from, to, num string) error {
	fromAddress := common.HexToAddress(from)
	nonce, err := e.cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	value := new(big.Int)
	value.SetString(num, 10)
	gasLimit := uint64(21000) // in units
	gasPrice, err := e.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := e.cli.NetworkID(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	ks := keystore.NewKeyStore(e.keyStoreDir, keystore.StandardScryptN, keystore.StandardScryptP)
	/*	signedTx, err := ks.SignTx(accounts.Account{Address: fromAddress}, tx, chainID)
		if err != nil {
			logger.Error(err)
			return err
		}*/
	signedTx, err := ks.SignTxWithPassphrase(accounts.Account{Address: fromAddress}, e.password, tx, chainID)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = e.cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("tx sent:", signedTx.Hash().Hex())
	return nil
}

func (e *EthClient) Transfer2(from, to, num, fromPriv string) error {
	privateKey, err := crypto.HexToECDSA(fromPriv)
	if err != nil {
		logger.Error(err)
		return err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		logger.Error("error casting public key to ECDSA")
		return errors.New("error casting public key to ECDSA")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	nonce, err := e.cli.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	value := new(big.Int)
	value.SetString(num, 10)
	gasLimit := uint64(21000) // in units
	gasPrice, err := e.cli.SuggestGasPrice(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	toAddress := common.HexToAddress(to)
	var data []byte
	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	chainID, err := e.cli.NetworkID(context.Background())
	if err != nil {
		logger.Error(err)
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		logger.Error(err)
		return err
	}

	/*	ts := types.Transactions{signedTx}
		rawTxBytes := ts.GetRlp(0)
		rawTxHex := hex.EncodeToString(rawTxBytes)

		logger.Info(rawTxHex) // f86...772

		rawTxBytes2, err := hex.DecodeString(rawTxHex)
		if err != nil {
			logger.Error(err)
			return err
		}

		tx2 := new(types.Transaction)
		rlp.DecodeBytes(rawTxBytes2, &tx2)
	*/

	err = e.cli.SendTransaction(context.Background(), signedTx)
	if err != nil {
		logger.Error(err)
		return err
	}

	logger.Info("tx sent:", signedTx.Hash().Hex())

	return nil
}

func NewEthClient(ipport string) (*EthClient, error) {
	e := new(EthClient)

	var err error
	e.cli, err = ethclient.Dial("http://" + ipport)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	e.password = "13717064390"
	e.keyStoreDir = "/home/thomas/eth/data/keystore"

	return e, nil
}
