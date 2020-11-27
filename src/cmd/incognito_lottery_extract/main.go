package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/robotrongt/incognito_node_bot/src/pkg/btc"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
)

type Ticket struct {
	models.LotteryTicket
}

func (t Ticket) String() string {
	return fmt.Sprintf("{%s, %s}", t.PubKey, models.GetTSString(t.Timestamp))
}

type Tickets []Ticket

func (t *Tickets) Extract() Ticket {
	var retval Ticket
	switch {
	case len(*t) > 0:
		ex := rand.Intn(len(*t))
		retval = (*t)[ex] //index is from 0..len-1
		// fmt.Println ("Len:", len(*t), " ex:", ex)
		// Remove the element at index ex from *t.
		copy((*t)[ex:], (*t)[ex+1:]) // Shift t[ex+1:] left one index.
		(*t)[len(*t)-1] = Ticket{}   // Erase last element (write zero value).
		(*t) = (*t)[:len(*t)-1]      // Truncate slice.
	default:
		retval = Ticket{}
	}
	return retval
}

//we search the first btc block after the tmExtract Time and return
//the nonce and the btc blockheight and btctimestamp or error
func getNonce(tmExtract time.Time) (int64, int64, int64, error) {
	//return 0, 0, 0, errors.New("Skipping btc search")
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
	return nonce, blockHeight, timestamp, nil
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
	tmTickets := tmExtract.AddDate(0, 0, -1)
	tsExtract := models.MakeTSFromTime(tmExtract)
	btcblock := BtcBlock{}
	useDbNonce := false
	tmTicketsStr := fmt.Sprintf("%s-%s", strconv.Itoa(tmTickets.Year()), strconv.Itoa(int(tmTickets.Month())))
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
		extract, err := env.Db.GetLotteryExtract(lottery.LOId, tmTickets)
		if err != nil {
			log.Println("error GetLotteryExtract:", err)
			break
		}
		log.Printf("Trying extraction n.%d for Lottery %d", extract, lottery.LOId)
		tickets, err := env.Db.GetLotteryTickets(lottery.LOId, tmTickets, -1) //prendiamo tutti i tickets e ripetiamo le estrazioni fino a extract -1
		if err != nil {
			log.Println("error GetLotteryTickets:", err)
			break
		}
		tt := Tickets{}
		for _, ticket := range tickets {
			tt = append(tt, Ticket{ticket})
		}
		rand.Seed(btcblock.Nonce)      //initialize with btc NONCE
		for i := 1; i < extract; i++ { //throw away previous extractions
			log.Printf("Re-Extraction %d: %s", i, tt.Extract())
		}
		winner := tt.Extract()
		winner.Extracted = int64(extract)
		tk := models.LotteryTicket{LOId: winner.LOId, PubKey: winner.PubKey, Timestamp: winner.Timestamp, Extracted: int64(extract)}
		log.Printf("Extraction %d: %s", extract, winner)
		//updating the ticket for winner
		err = env.Db.UpdateLotteryTicketWinner(tk)
		if err != nil {
			log.Println("error UpdateLotteryTicketWinner:", err)
			break
		}
		//get the default alias for this pubkey
		lotterykey := env.Db.GetLotteryKeyByKey(winner.LOId, winner.PubKey)
		defaultalias := lotterykey.DefaultAlias
		flag := ""
		if winner.Extracted > 0 {
			flag = fmt.Sprintf("(%d)", winner.Extracted)
		}
		if winner.Extracted == 1 {
			flag = "🥇"
		}
		if winner.Extracted == 2 {
			flag = "🥈"
		}
		if winner.Extracted == 3 {
			flag = "🥉"
		}
		//ready to loop the chats of this lottery
		lotterychats, err := env.Db.GetLotteryChatIDS(winner.LOId)
		if err != nil {
			log.Println("error GetLotteryChatIDS:", err)
			break
		}
		for _, lotterychat := range lotterychats {
			//vediamo se la chat vuole essere notificata
			chatuser, err := env.Db.GetUserByChatID(lotterychat.ChatID)
			if !chatuser.Notify {
				log.Println("Skipping notify ChatUser:", chatuser.Name)
				break
			}
			// vediamo se la chat ha un alias specifico
			thealias := defaultalias
			chatkey, err := env.Db.GetChatKeyFromPub(lotterychat.ChatID, winner.PubKey)
			if err == nil { // se nessun errore prendiamo alias specifico
				thealias = chatkey.KeyAlias
			}
			msg := fmt.Sprintf("Hello %s, in lottery %s the winner of %s is...", chatuser.Name, lottery.LotteryName, tmTicketsStr)
			msg = fmt.Sprintf("%s\n%s", msg, " 🥳🎊🎉 🥳🎊🎉")
			msg = fmt.Sprintf("%s\n%s %s %s", msg, thealias, models.GetTSString(winner.Timestamp), flag)
			msg = fmt.Sprintf("%s\n%s", msg, " 🥳🎊🎉 🥳🎊🎉")
			msg = fmt.Sprintf("%s\nThe seed of the extraction is taken from blockchain block height %d (%s)", msg, btcblock.Height, models.GetTSString(btcblock.Timestamp))
			msg = fmt.Sprintf("%s\nNonce: %d", msg, btcblock.Nonce)
			msg = fmt.Sprintf("%s\nYou can verify it here: https://www.blockchain.com/btc/block/%d", msg, btcblock.Height)
			msg = fmt.Sprintf("%s\nAnd this is a sample code to test https://play.golang.org/p/WDF3-Eoh_l7", msg)
			if err := env.SayText(lotterychat.ChatID, msg); err != nil {
				log.Println("Error sending msg:", msg)
			}
		}
	}
}
