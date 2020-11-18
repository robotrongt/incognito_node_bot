package models

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

const PRV_ID = "0000000000000000000000000000000000000000000000000000000000000004"

type COIN struct {
	Name string
	ID   string
	Dec  float64
}

type COINS []COIN

var BIG_COINS = COINS{
	COIN{Name: "PRV", ID: PRV_ID, Dec: 1e-09},
	COIN{Name: "ETH", ID: "ffd8d42dc40a8d166ea4848baf8b5f6e912ad79875f4373070b59392b1756c8f", Dec: 1e-09},
	COIN{Name: "USDT", ID: "716fd1009e2a1669caacc36891e707bfdf02590f96ebd897548e8963c95ebac0", Dec: 1e-06},
	COIN{Name: "USDC", ID: "1ff2da446abfebea3ba30385e2ca99b0f0bbeda5c6371f4c23c939672b429a42", Dec: 1e-06},
	COIN{Name: "BTC", ID: "b832e5d3b1f01a4f0623f7fe91d6673461e1f5d37d91fe78c5c2e6183ff39696", Dec: 1e-09},
	COIN{Name: "BUSD", ID: "9e1142557e63fd20dee7f3c9524ffe0aa41198c494aa8d36447d12e85f0ddce7", Dec: 1e-09},
	COIN{Name: "BNB", ID: "b2655152784e8639fa19521a7035f331eea1f1e911b2f3200a507ebb4554387b", Dec: 1e-09},
	COIN{Name: "DAI", ID: "3f89c75324b46f13c7b036871060e641d996a24c09b3065835cb1d38b799d6c1", Dec: 1e-09},
	COIN{Name: "SAI", ID: "d240c61c6066fed0535df9302f1be9f5c9728ef6d01ce88d525c4f6ff9d65a56", Dec: 1e-09},
	COIN{Name: "TUSD", ID: "8c3a61e77061265aaefa1e7160abfe343c2189278dd224bb7da6e7edc6a1d4db", Dec: 1e-09},
	COIN{Name: "TOMO", ID: "a0a22d131bbfdc892938542f0dbe1a7f2f48e16bc46bf1c5404319335dc1f0df", Dec: 1e-09},
	COIN{Name: "LINK", ID: "e0926da2436adc42e65ca174e590c7b17040cd0b7bdf35982f0dd7fc067f6bcf", Dec: 1e-09},
	COIN{Name: "BAT", ID: "1fe75e9afa01b85126370a1583c7af9f1a5731625ef076ece396fcc6584c2b44", Dec: 1e-09},
	COIN{Name: "BAND", ID: "2dda855fb4660225882d11136a64ad80effbddfa18a168f78924629b8664a6b3", Dec: 1e-09},
	COIN{Name: "ZRX", ID: "de395b1914718702687b477703bdd36e52119033a9037bb28f6b33a3d0c2f867", Dec: 1e-09},
	COIN{Name: "FTM", ID: "d09ad0af0a34ea3e13b772ef9918b71793a18c79b2b75aec42c53b69537029fe", Dec: 1e-09},
	COIN{Name: "ZIL", ID: "880ea0787f6c1555e59e3958a595086b7802fc7a38276bcd80d4525606557fbc", Dec: 1e-09},
	COIN{Name: "MCO", ID: "caaf286e889a8e0cee122f434d3770385a0fd92d27fcee737405b73c45b4f05f", Dec: 1e-09},
	COIN{Name: "GUSD", ID: "465b0f709844be95d97e1f5c484e79c6c1ac51d28de2a68020e7313d34f644fe", Dec: 1e-09},
	COIN{Name: "PAX", ID: "4a790f603aa2e7afe8b354e63758bb187a4724293d6057a46859c81b7bd0e9fb", Dec: 1e-09},
	COIN{Name: "KCS", ID: "513467653e06af73cd2b2874dd4af948f11f1c6f2689e994c055fd6934349e05", Dec: 1e-09},
	COIN{Name: "OMG", ID: "249ca174b4dce58ea6e1f8eda6e6f74ab6a3de4e4913c4f50c15101001bb467b", Dec: 1e-09},
}

func (coins *COINS) GetNameByID(id string) string {
	for _, coin := range *coins {
		if coin.ID == id {
			return coin.Name
		}
	}
	return ""
}

func (coins *COINS) GetCoinByName(name string) COIN {
	for _, coin := range *coins {
		if coin.Name == name {
			return coin
		}
	}
	return COIN{"", "", 0}
}

func (coins *COINS) GetFloat64Val(name string, valint int64) float64 {
	val := float64(valint) * coins.GetCoinByName(name).Dec
	return val
}

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

type TMinerReward map[string]int64

func (mr *TMinerReward) GetValueIDs() []string {
	retval := []string{}
	for id := range *mr {
		retval = append(retval, id)
	}
	return retval
}

func (mr *TMinerReward) GetValueNames() []string {
	retval := []string{}
	for id := range *mr {
		n := BIG_COINS.GetNameByID(id)
		if n == "" {
			retval = append(retval, id)
		} else {
			retval = append(retval, n)
		}
	}
	return retval
}

func (mr *TMinerReward) GetNameValuePair(nameid string) (string, int64) {
	nm := ""
	val := int64(0)
	//	val = float64((*mr)[nameid]) / float64(1000000000)
	val = (*mr)[nameid]
	nm = BIG_COINS.GetNameByID(nameid)
	if nm == "" {
		nm = nameid
	}
	return nm, val
}
func (mr *TMinerReward) GetPRV() int64 {
	_, val := mr.GetNameValuePair("PRV")
	return val
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
	//log.Printf("Result.PRV: %f\n", float64(mrmfk.Result.PRV)/float64(1000000000))
	log.Printf("Result.PRV: %.9f\n", BIG_COINS.GetFloat64Val("PRV", mrmfk.Result.GetPRV()))
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
