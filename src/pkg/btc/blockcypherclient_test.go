//+build !test

package btc

import (
	"testing"
	"time"
)

/*
func TestGetChainTimeStampAndNonceBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	_, err := btcClient.GetCurrentChainTimeStamp()
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
	}
}
func TestGetTimestampAndNonceByBlockHeightBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	timestamp, nonce, err := btcClient.GetTimeStampAndNonceByBlockHeight(2)
	t.Log(timestamp, nonce)
	if err != nil {
		t.Error("Fail to get timestamp and nonce")
	}
	if timestamp != 1231469744 {
		t.Error("Wrong Timestamp")
	}
	if nonce != 1639830024 {
		t.Error("Wrong Nonce")
	}
}
*/
func TestGetNonceByTimeStampBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	var tm, errParse = time.ParseInLocation("02-01-2006 15:04:05", "01-11-2020 00:00:00", time.Now().Location())
	var ts = makeTimestamp(tm)
	var td time.Duration = time.Duration(120) * time.Second
	t.Log(tm.Local().Format("02-01-2006 15:04:05 MST"), errParse)
	blockHeight, timestamp, nonce, err := btcClient.GetNonceByTimestamp(time.Now(), td, ts)
	t.Log("H:", blockHeight, "TS:", timestamp, "NN:", nonce, "err:", err)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce", err)
		t.Fatal("Exiting")
	}
	if blockHeight != 654922 {
		t.Error("Wrong Block")
	}
	if timestamp != int64(1604185338) {
		t.Error("Wrong Timestamp")
	}
	if nonce != int64(1311888545) {
		t.Error("Wrong Nonce")
	}
	//check Block at https://api.blockcypher.com/v1/btc/main/blocks/654922?start=1&count=1
	// https://www.blockchain.com/btc/block/654922
}

/*
func TestVerifyNonceByTimeStampBlockCypher(t *testing.T) {
	var btcClient = BlockCypherClient{}
	isOk, err := btcClient.VerifyNonceWithTimestamp(1373297940, 3029573794)
	if err != nil {
		t.Error("Fail to get chain timestamp and nonce")
		t.FailNow()
	}
	if !isOk {
		t.Error("Fail to verify nonce by timestamp")
	}
}
*/
