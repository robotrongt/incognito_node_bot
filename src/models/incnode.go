package models

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type TMiningPubKey struct {
	bls string
	dsa string
}
type TPubKey struct {
	IncPubKey    string
	MiningPubKey TMiningPubKey
}
type TPubKeyAuto struct {
	IncPubKey    string
	MiningPubKey TMiningPubKey
	IsAutoStake  bool
}

type TBeaconStateResult struct {
	BestBlockHash                          string
	PreviousBestBlockHash                  string
	BestShardHash                          map[string]string
	BestShardHeight                        map[string]int
	Epoch                                  int
	BeaconHeight                           int
	BeaconProposerIndex                    int
	BeaconCommittee                        []TPubKey
	BeaconPendingValidator                 []TPubKey
	CandidateShardWaitingForCurrentRandom  []TPubKey
	CandidateBeaconWaitingForCurrentRandom []TPubKey
	CandidateShardWaitingForNextRandom     []TPubKey
	CandidateBeaconWaitingForNextRandom    []TPubKey
	RewardReceiver                         interface{}
	ShardCommittee                         map[string][]TPubKey
	ShardPendingValidator                  map[string][]TPubKey
	AutoStaking                            []TPubKeyAuto
	CurrentRandomNumber                    int
	CurrentRandomTimeStamp                 int
	IsGetRandomNumber                      bool
	MaxBeaconCommitteeSize                 int
	MinBeaconCommitteeSize                 int
	MaxShardCommitteeSize                  int
	MinShardCommitteeSize                  int
	ActiveShards                           int
	LastCrossShardState                    interface{}
	ShardHandle                            interface{}
}
type BBSD struct {
	Id      int
	Result  TBeaconStateResult
	Error   string
	Params  []string
	Method  string
	Jsonrpc string
}

func GetBeaconBestStateDetail(reqUrl string, user *ChatUser, bbsd *BBSD) error {
	myClient := &http.Client{Timeout: 10 * time.Second}
	reqBody := strings.NewReader(`
	  {
		"id": 1,
		"jsonrpc": "1.0",
		"method": "getbeaconbeststatedetail",
		"params": []
	  }
	`)
	req, err := http.NewRequest(
		"GET",
		reqUrl,
		reqBody,
	)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	getJson(myClient, req, &bbsd)
	fmt.Printf("Result.BeaconHeight: %d\n", bbsd.Result.BeaconHeight)
	fmt.Printf("Result.Epoch: %d\n", bbsd.Result.Epoch)
	return err
}

func getJson(myClient *http.Client, req *http.Request, target interface{}) error {
	res, errGet := myClient.Do(req)
	if errGet != nil {
		log.Fatal("Error: ", errGet)
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}
