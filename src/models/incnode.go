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
	Bls string
	Dsa string
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
type TPubKeyInfo struct {
	IncPubKey    string
	MiningPubKey TMiningPubKey
	IsAutoStake  bool
	PRV          int64
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

type TBestBlock struct {
	Height              int64
	Hash                string
	TotalTxs            int64
	BlockProducer       string
	ValidationData      interface{}
	Epoch               int64
	Time                int64
	RemainingBlockEpoch int
	EpochBlock          int
}
type TBlochChainInfo struct {
	ChainName    string
	BestBlocks   map[string]TBestBlock
	ActiveShards int
}
type BCI struct {
	Id      int
	Result  TBlochChainInfo
	Error   string
	Params  []string
	Method  string
	Jsonrpc string
}

type TMinerReward struct {
	PRV int64
}
type MRFMK struct {
	Id      int
	Result  TMinerReward
	Error   string
	Params  []string
	Method  string
	Jsonrpc string
}

func CheckIfPresent(pubkey string, arr *[]TPubKey) bool {
	retval := false
	for _, tpk := range *arr {
		if tpk.IncPubKey == pubkey {
			retval := true
			return retval
		}

	}
	return retval
}

//ritorna vero se trovato in AutoStake=true più la completa TPubKeyAuto
func CheckAutoStake(pubkey string, arr *[]TPubKeyAuto) (bool, *TPubKeyAuto) {
	for _, tpka := range *arr {
		if tpka.IncPubKey == pubkey {
			retval := tpka.IsAutoStake
			pka := tpka
			return retval, &pka
		}
	}
	return false, nil
}

//ritorna status più puntatore a TPubKeyInfo se trovata attiva
func GetPubKeyStatus(bbsd *BBSD, pubkey string) (string, *TPubKeyInfo) {
	pki := TPubKeyInfo{}
	pki.IncPubKey = pubkey
	result := "missing"
	up := "👆"
	down := "👇"
	autostake, tpka := CheckAutoStake(pubkey, &bbsd.Result.AutoStaking)
	if tpka != nil {
		pki.IncPubKey = tpka.IncPubKey
		pki.MiningPubKey = tpka.MiningPubKey
		pki.IsAutoStake = tpka.IsAutoStake
		pki.PRV = 0
	}

	as := down     //indice in basso
	if autostake { //se autostake allora
		as = up //indice in alto
	}
	if CheckIfPresent(pubkey, &bbsd.Result.CandidateShardWaitingForNextRandom) {
		result = fmt.Sprintf("%s%s", "Waiting", as)
		return result, &pki
	}
	if CheckIfPresent(pubkey, &bbsd.Result.CandidateShardWaitingForCurrentRandom) {
		result = fmt.Sprintf("%s%s", "Waiting", as)
		return result, &pki
	}
	for shard, arrpk := range bbsd.Result.ShardPendingValidator {
		if CheckIfPresent(pubkey, &arrpk) {
			result = fmt.Sprintf("%s shard %s%s", "Pending", shard, as)
			return result, &pki
		}
	}
	for shard, arrpk := range bbsd.Result.ShardCommittee {
		if CheckIfPresent(pubkey, &arrpk) {
			result = fmt.Sprintf("%s shard %s%s", "Committee", shard, as)
			return result, &pki
		}
	}
	if CheckIfPresent(pubkey, &bbsd.Result.CandidateBeaconWaitingForNextRandom) {
		result = fmt.Sprintf("%s%s", "BeaconWaiting", as)
		return result, &pki
	}
	if CheckIfPresent(pubkey, &bbsd.Result.CandidateBeaconWaitingForCurrentRandom) {
		result = fmt.Sprintf("%s%s", "BeaconWaiting", as)
		return result, &pki
	}
	if CheckIfPresent(pubkey, &bbsd.Result.BeaconPendingValidator) {
		result = fmt.Sprintf("%s%s", "BeaconPending", as)
		return result, &pki
	}
	if CheckIfPresent(pubkey, &bbsd.Result.BeaconCommittee) {
		result = fmt.Sprintf("%s%s", "BeaconCommittee", as)
		return result, &pki
	}
	return result, nil
}

func GetBeaconBestStateDetail(reqUrl string, bbsd *BBSD) error {
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
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	err = getJson(myClient, req, &bbsd)
	if err != nil {
		return err
	}
	log.Printf("Result.BeaconHeight: %d\n", bbsd.Result.BeaconHeight)
	log.Printf("Result.Epoch: %d\n", bbsd.Result.Epoch)
	return err
}

func GetBlockChainInfo(reqUrl string, bci *BCI) error {
	myClient := &http.Client{Timeout: 10 * time.Second}
	reqBody := strings.NewReader(`
	  {
		"id": 1,
		"jsonrpc": "1.0",
		"method": "getblockchaininfo",
		"params": []
	  }
	`)
	req, err := http.NewRequest(
		"GET",
		reqUrl,
		reqBody,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	err = getJson(myClient, req, &bci)
	if err != nil {
		return err
	}
	log.Printf("Result.ChainName: %s\n", bci.Result.ChainName)
	log.Printf("Result.ActiveShards: %d\n", bci.Result.ActiveShards)
	return err
}

func GetMinerRewardFromMiningKey(reqUrl string, key string, mrmfk *MRFMK) error {
	myClient := &http.Client{Timeout: 10 * time.Second}
	reqBody := strings.NewReader(`
	  {
		"id": 1,
		"jsonrpc": "1.0",
		"method": "getminerrewardfromminingkey",
		"params": ["` + key + `"]
	  }
	`)
	req, err := http.NewRequest(
		"GET",
		reqUrl,
		reqBody,
	)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")

	err = getJson(myClient, req, &mrmfk)
	if err != nil {
		return err
	}
	log.Printf("Result.PRV: %d\n", mrmfk.Result.PRV)
	return err
}

func getJson(myClient *http.Client, req *http.Request, target interface{}) error {
	res, errGet := myClient.Do(req)
	if errGet != nil {
		log.Printf("Error in myClient: %s", errGet)
		return errGet
	}
	defer res.Body.Close()

	return json.NewDecoder(res.Body).Decode(target)
}
