package main

import (
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
)

func main() {
	env := models.NewEnv()
	defer env.Db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", env.Db, env.Db.DB)

	rand.Seed(time.Now().UnixNano())

	var miningkeys *[]models.MiningKey
	var err error
	miningkeys, err = env.Db.GetMiningKeys(100, 0)
	if err != nil {
		log.Fatal(err)
	}
	theUrl := env.DEFAULT_NODE_URL
	bbsd := models.BBSD{}
	if err := models.GetBeaconBestStateDetail(theUrl, &bbsd); err != nil {
		log.Println("error getBeaconBestStateDetail:", err)
		return
	}
	for _, miningkey := range *miningkeys {
		status, pki := models.GetPubKeyStatus(&bbsd, miningkey.PubKey)
		mk := &models.MiningKey{
			PubKey:     miningkey.PubKey,
			LastStatus: status,
		}
		if pki != nil { //abbiamo info della chiave
			mk.LastPRV = pki.PRV
			mk.IsAutoStake = pki.IsAutoStake
			mk.Bls = pki.MiningPubKey.Bls
			mk.Dsa = pki.MiningPubKey.Dsa
			mrfmk := models.MRFMK{}
			err := models.GetMinerRewardFromMiningKey(env.DEFAULT_FULLNODE_URL, "bls:"+mk.Bls, &mrfmk)
			if err == nil { //no err, abbiamo anche i Saldi
				mk.LastPRV = mrfmk.Result.GetPRV()
			} else { //non abbiamo i PRV
				mk.LastPRV = -1 //segnaliamo che non è da aggiornare
			}
		}
		env.Db.UpdateMiningKey(mk, models.StatusChangeNotifierFunc(env.StatusChanged))
	}
}
