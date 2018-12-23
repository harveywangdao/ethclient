package test

import (
	"ethclient/eth"
	"ethclient/logger"
	"sync"
	"time"
)

type EthClientTest struct {
	ethIpPort string
	coinbase  string
	cli       *eth.EthClient
	password  string
	keyStore  string
}

func (e *EthClientTest) testApi() error {
	var err error

	//Transfer
	_, err = e.cli.GetBalance(e.coinbase)
	if err != nil {
		logger.Error(err)
		return err
	}

	addr1, err := e.cli.CreateKeyStore(e.password, e.keyStore)
	if err != nil {
		logger.Error(err)
		return err
	}
	e.cli.GetBalance(addr1)

	txid, err := e.cli.Transfer(e.coinbase, addr1, "2000000000000000000", e.password, e.keyStore)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = e.cli.QueryTransaction(txid)
	if err != nil {
		logger.Error(err)
		return err
	}

	oldBalance, _ := e.cli.GetBalance(addr1)
	for {
		time.Sleep(2 * time.Second)
		newBalance, _ := e.cli.GetBalance(addr1)
		if newBalance != oldBalance {
			break
		}
	}

	//Contract
	contractAddress, err := e.cli.DeployContract(addr1, e.password, e.keyStore)
	if err != nil {
		logger.Error(err)
		return err
	}

	oldBalance, _ = e.cli.GetBalance(addr1)
	for {
		time.Sleep(2 * time.Second)
		newBalance, _ := e.cli.GetBalance(addr1)
		if newBalance != oldBalance {
			break
		}
	}

	err = e.cli.LoadContract(contractAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	err = e.cli.InvokeContract(addr1, e.password, e.keyStore, contractAddress)
	if err != nil {
		logger.Error(err)
		return err
	}

	return nil
}

func (e *EthClientTest) testing(wg *sync.WaitGroup) {
	defer wg.Done()
	var err error

	err = e.testApi()
	if err != nil {
		logger.Error(err)
		return
	}

	/*
		err = cli.GetNewWallet()
		if err != nil {
			logger.Error(err)
			return
		}

		err = cli.Header()
		if err != nil {
			logger.Error(err)
			return
		}

		err = cli.BlockByNumber(9848)
		if err != nil {
			logger.Error(err)
			return
		}

		err = cli.CheckAddress(e.coinbase)
		if err != nil {
			logger.Error(err)
			return
		}

		err = cli.ImportKeyStore("./keystore/UTC--2018-10-13T16-51-24.438238692Z--abe480fd76b8a1515115ef947aa140f85b602ae4")
		if err != nil {
			logger.Error(err)
			return
		}
	*/
}

func NewEthClientTest(ipPort, coinbase, password, keyStore string, wg *sync.WaitGroup) (*EthClientTest, error) {
	test := new(EthClientTest)
	test.ethIpPort = ipPort
	test.coinbase = coinbase
	test.password = password
	test.keyStore = keyStore

	cli, err := eth.NewEthClient(test.ethIpPort)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	test.cli = cli

	go test.testing(wg)

	return test, nil
}
