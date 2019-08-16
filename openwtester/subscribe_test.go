/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package openwtester

import (
	"path/filepath"
	"testing"

	"github.com/astaxie/beego/config"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openw"
	"github.com/blocktree/openwallet/openwallet"
)

////////////////////////// 测试单个扫描器 //////////////////////////

type subscriberSingle struct {
	manager *openw.WalletManager
}

//BlockScanNotify 新区块扫描完成通知
func (sub *subscriberSingle) BlockScanNotify(header *openwallet.BlockHeader) error {
	log.Notice("header:", header)
	return nil
}

//BlockTxExtractDataNotify 区块提取结果通知
func (sub *subscriberSingle) BlockExtractDataNotify(sourceKey string, data *openwallet.TxExtractData) error {
	log.Notice("account:", sourceKey)

	for i, input := range data.TxInputs {
		log.Std.Notice("data.TxInputs[%d]: %+v", i, input)
	}

	for i, output := range data.TxOutputs {
		log.Std.Notice("data.TxOutputs[%d]: %+v", i, output)
	}

	log.Std.Notice("data.Transaction: %+v", data.Transaction)

	walletID := "WKFkmvsSFz5mC1cAX3edJC2e6hH6ow3X9E"
	accountID := "HX4tUVg5eETb6SvZeGeAFwk4PQ1CWS6dQeyjj3CqfYyK"

	balance, err := sub.manager.GetAssetsAccountBalance(testApp, walletID, accountID)
	if err != nil {
		log.Error("GetAssetsAccountBalance failed, unexpected error:", err)
		return nil
	}
	log.Notice("account balance:", balance)

	return nil
}

func TestSubscribeAddress(t *testing.T) {

	var (
		endRunning = make(chan bool, 1)
		symbol     = "VLX"
		addrs      = map[string]string{
			"VLSWWh9SCcutqB9APLSxyUyfzeuvG8XXTB1": "sender",
			"VLa1hi77ZXD2BSWDD9wQe8vAhejXyS7vBM4": "receiver",
		}
	)

	//GetSourceKeyByAddress 获取地址对应的数据源标识
	scanAddressFunc := func(address string) (string, bool) {
		key, ok := addrs[address]
		if !ok {
			return "", false
		}
		return key, true
	}

	assetsMgr, err := openw.GetAssetsAdapter(symbol)
	if err != nil {
		log.Error(symbol, "is not support")
		return
	}

	//读取配置
	absFile := filepath.Join(configFilePath, symbol+".ini")

	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return
	}
	assetsMgr.LoadAssetsConfig(c)

	assetsLogger := assetsMgr.GetAssetsLogger()
	if assetsLogger != nil {
		assetsLogger.SetLogFuncCall(true)
	}

	//log.Debug("already got scanner:", assetsMgr)
	scanner := assetsMgr.GetBlockScanner()
	scanner.SetRescanBlockHeight(150503)

	if scanner == nil {
		log.Error(symbol, "is not support block scan")
		return
	}

	scanner.SetBlockScanAddressFunc(scanAddressFunc)

	sub := subscriberSingle{manager: testInitWalletManager()}
	scanner.AddObserver(&sub)

	scanner.Run()

	<-endRunning
}