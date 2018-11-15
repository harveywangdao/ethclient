package main

import (
	"ethclient/logger"
	"ethclient/test"
	"gopkg.in/ini.v1"
	"log"
	"sync"
)

const (
	EthClientConfFilePath = "conf/my.ini"
)

func initLogger() error {
	//fileHandler := logger.NewFileHandler("test.log")
	//logger.SetHandlers(logger.Console, fileHandler)
	logger.SetHandlers(logger.Console)
	//defer logger.Close()
	logger.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	logger.SetLevel(logger.INFO)

	return nil
}

func main() {
	err := initLogger()
	if err != nil {
		log.Fatalln(err)
	}

	cfg, err := ini.Load(EthClientConfFilePath)
	if err != nil {
		logger.Error(err)
		return
	}

	ipport := cfg.Section("").Key("EthServerIpPort").String()

	var wg sync.WaitGroup
	wg.Add(1)

	_, err = test.NewEthClientTest(ipport, &wg)
	if err != nil {
		logger.Error(err)
		return
	}

	wg.Wait()
	logger.Debug("ethereum client exit")
}
