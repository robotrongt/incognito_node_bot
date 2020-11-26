package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
	"github.com/robotrongt/incognito_node_bot/src/pkg/btc"
)

type Ticket struct {
	Nodo      string
	Timestamp string
}

func (t Ticket) String() string {
	return fmt.Sprintf("{%s, %s}", t.Nodo, t.Timestamp)
}

type Tickets []Ticket

//we search the first btc block after the tmExtract Time and return
//the nonce and the btc blockheight and btctimestamp or error
func getNonce(tmExtract time.Time) (int64, int64, int64, error) {
	var btcClient = btc.BlockCypherClient{}
	var ts = models.MakeTSFromTime(tmExtract)
	var td time.Duration = time.Duration(120) * time.Second //define the timeout to get the correct block
	log.Println("Finding first BTC BLOCK after:", tmExtract)
	blockHeight, timestamp, nonce, err := btcClient.GetNonceByTimestamp(time.Now(), td, ts)
	if err != nil {
		log.Println("Fail to get chain timestamp and nonce", err)
		return 0, 0, 0, err
	}
	//check Block at https://api.blockcypher.com/v1/btc/main/blocks/654922?start=1&count=1
	// https://www.blockchain.com/btc/block/654922
	log.Printf("Found BTC BLOCK at height %d with nonce %d\n", blockHeight, nonce)
	log.Printf("Verify at:\n\thttps://api.blockcypher.com/v1/btc/main/blocks/%d?start=1&count=1\n\thttps://www.blockchain.com/btc/block/%d\n", blockHeight, blockHeight)
	return nonce, int64(blockHeight), timestamp, nil
}

type BtcBlock struct {
	Nonce     int64
	Height    int64
	Timestamp int64
}

func main() {
	env := models.NewEnv()
	defer env.Db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", env.Db, env.Db.DB)

	rand.Seed(time.Now().UnixNano())

	tmNow := time.Now() //prendiamo data attuale
	//per l'estrazione prendiamo il primo blocco BTC dopo mezzanotte ora locale
	//del primo del mese
	tmExtract := time.Date(tmNow.Year(), tmNow.Month(), 1, 0, 0, 0, 0, tmNow.Location())
	tsExtract := models.MakeTSFromTime(tmExtract)
	btcblock := BtcBlock{}
	useDbNonce := false
	if nonce, blockHeight, btcts, err := getNonce(tmExtract); err == nil {
		btcblock = BtcBlock{Nonce: nonce, Height: blockHeight, Timestamp: btcts}
		log.Println("tmExtract:", tmExtract)
		log.Println("tsExtract:", tsExtract)
		log.Println("blockHeight:", blockHeight)
		log.Println("btcts:", btcts, models.GetTSTime(btcts))
		log.Println("nonce:", nonce)
	} else {
		log.Println("error searching btc block and nonce:", err)
		useDbNonce = true
	}
	lotteries, err := env.Db.GetLotteries()
	if err != nil {
		log.Println("Err in GetLotteries:", err)
		return
	}
	for _, lottery := range lotteries {
		log.Println("Lottery:", lottery)
		var lotteryextraction models.LotteryExtraction
		var err error
		lotteryextraction, err = env.Db.GetLotteryExtraction(lottery.LOId, tsExtract)
		if err != nil {
			log.Println("error GetLotteryExtraction:", err)
			if useDbNonce { //se dobbiamo usare db usciamo perche non l'abbiamo trovato
				break
			} else { // non abbiamo db, salviamo dato blockchain
				lotteryextraction.LOId = lottery.LOId
				lotteryextraction.Nonce = btcblock.Nonce
				lotteryextraction.Timestamp = btcblock.Timestamp
				lotteryextraction.BTCBlock = btcblock.Height
				if err := env.Db.ReplaceLotteryExtraction(lotteryextraction); err != nil {
					log.Println("error ReplaceLotteryExtraction:", err)
					break
				}
			}
		}
		if useDbNonce { // usiamo il record dell'ultima estrazione del periodo se c'è
			btcblock = BtcBlock{Nonce: lotteryextraction.Nonce, Height: lotteryextraction.BTCBlock, Timestamp: lotteryextraction.Timestamp}
		} else { // controlliamo la validità dell'estrazione nel db e nel caso la sostituiamo
			if lotteryextraction.Nonce != btcblock.Nonce || lotteryextraction.Timestamp != btcblock.Timestamp || lotteryextraction.BTCBlock != btcblock.Height {
				lotteryextraction.Nonce = btcblock.Nonce
				lotteryextraction.Timestamp = btcblock.Timestamp
				lotteryextraction.BTCBlock = btcblock.Height
				if err := env.Db.ReplaceLotteryExtraction(lotteryextraction); err != nil {
					log.Println("error ReplaceLotteryExtraction:", err)
					break
				}
			}
		}
		//se siamo qui abbiamo il dato del DB dell'estrazione salvato o aggiornato ed
		// il btcblock eventualmente caricato col dato del db
		log.Println("btcblock.Nonce:", btcblock.Nonce)
		log.Println("btcblock.Timestamp:", btcblock.Timestamp)
		log.Println("btcblock.Height:", btcblock.Height)
	}
}
