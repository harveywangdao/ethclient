package test

import (
	"ethclient/eth"
	"ethclient/logger"
	"sync"
)

type EthClientTest struct {
	EthIpPort string
	coinbase  string
	addrs     []string
}

func (e *EthClientTest) testing(wg *sync.WaitGroup) {
	defer wg.Done()

	cli, err := eth.NewEthClient(e.EthIpPort)
	if err != nil {
		logger.Error(err)
		return
	}

	richAddrHex := "0xf036eF6F352048b27C291295B6f6DCD237973a0d"
	richAddrPriv := "24e37b26f354d123eb4d2675bac002fd802e4db39d6df10fcd2faae8c111b586"
	contractAddr := "0x2b01981B95904CA40B0F390b5D196D4fe56e2f0E"

	err = cli.InvokeContract(richAddrPriv, contractAddr)
	if err != nil {
		logger.Error(err)
		return
	}
	return

	cli.GetBalance(richAddrHex)
	err = cli.DeployContract(richAddrPriv)
	if err != nil {
		logger.Error(err)
		return
	}

	cli.Transfer(e.coinbase, richAddrHex, "1000000000000000000")

	err = cli.QueryTransaction("0x98593fe3321925c8ef1fb2acdfbd93932a97bbe79654a1bc6d06a8746c7806f0")
	if err != nil {
		logger.Error(err)
		return
	}

	cli.GetBalance(e.addrs[0])
	cli.GetBalance(e.addrs[1])

	err = cli.GetNewWallet()
	if err != nil {
		logger.Error(err)
		return
	}

	err = cli.Transfer(e.addrs[0], e.addrs[1], "55555")
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

	err = cli.CreateKeyStore()
	if err != nil {
		logger.Error(err)
		return
	}

	err = cli.ImportKeyStore("./keystore/UTC--2018-10-13T16-51-24.438238692Z--abe480fd76b8a1515115ef947aa140f85b602ae4")
	if err != nil {
		logger.Error(err)
		return
	}
}

func (e *EthClientTest) setData() {
	e.coinbase = "0xe89d4872b78ab5c5c903583725fe5d485686d6ce"
	e.addrs = append(e.addrs, "0x044b8ab7c603f0938f53e72b7586ec38f3eff044")
	e.addrs = append(e.addrs, "0xced5d036328e9b6da0ed9a7ce1a7770b951fc636")
}

func NewEthClientTest(ipPort string, wg *sync.WaitGroup) (*EthClientTest, error) {
	test := new(EthClientTest)
	test.EthIpPort = ipPort

	test.setData()

	go test.testing(wg)

	return test, nil
}
