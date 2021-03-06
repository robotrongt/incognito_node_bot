package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
)

type MyEnv struct {
	*models.Env
}

func main() {
	env := &MyEnv{models.NewEnv()}
	defer env.Db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", env.Db, env.Db.DB)
	err := env.Db.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}

	rand.Seed(time.Now().UnixNano())

	for _, cmd := range env.BOT_CMDS {
		log.Printf("%s - %s\n", cmd.Cmd, cmd.Descr)
	}
	http.HandleFunc("/", env.RootHandler)
	http.HandleFunc("/telegram"+env.TGTOKEN+"/", env.TelegramHandler)
	log.Fatal(http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", nil))
}

// Create a struct that mimics the webhook response body
// https://core.telegram.org/bots/api#update
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

// This handler is called everytime someone requests any other page and writes on console
func (env MyEnv) RootHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hi! Nice request %s!", r.URL.Path)
	log.Printf("IP:%s (for: %s) (ua: %s) requested %s", r.RemoteAddr, r.Header.Get("X-Forwarded-For"), r.UserAgent(), r.URL.Path)
}

// This handler is called everytime telegram sends us a webhook event
func (env MyEnv) TelegramHandler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		log.Println("could not decode request body", err)
		return
	}
	ChatData, _ := env.Db.GetUserByChatID(body.Message.Chat.ID)
	bbsd := models.BBSD{}
	bci := models.BCI{}
	log.Println("Ricevuto:", body.Message.Text)
	switch {
	case env.StrCmd(body.Message.Text) == "/start":
		ChatData.NameAsked = true
		if err := env.Db.UpdateUser(ChatData); err != nil {
			log.Println("error updating name:", err)
		}
		if err := env.SayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" come ti chiami?"); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case ChatData.NameAsked:
		ChatData.Name = body.Message.Text
		ChatData.NameAsked = false
		if err := env.Db.UpdateUser(ChatData); err != nil {
			log.Println("error updating name:", err)
		}
		if err := env.SayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" ora mi ricordo di te!"); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "fiona") || strings.Contains(strings.ToLower(body.Message.Text), "olindo"):
		if err := env.SayText(body.Message.Chat.ID, env.Db.GetFionaText()); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "ringraziamento"):
		if err := env.SayText(body.Message.Chat.ID, env.Db.GetRingraziamentoText()); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/height":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		if np > 0 {
			nodo = params[0]
		}
		log.Println("/height", nodo, np, params)
		theUrl := env.DEFAULT_NODE_URL
		if urlNode, err := env.Db.GetUrlNode(ChatData.ChatID, nodo); err == nil {
			theUrl = urlNode.NodeURL
		} else {
			if np > 0 {
				messaggio := fmt.Sprintf("Non trovo tuo nodo \"%s\" uso mio nodo", nodo)
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				nodo = ""
			}
		}
		if err := models.GetBeaconBestStateDetail(theUrl, &bbsd); err != nil {
			log.Println("error getBeaconBestStateDetail:", err)
			env.SayErr(body.Message.Chat.ID, err)
			return
		}
		if err := models.GetBlockChainInfo(theUrl, &bci); err != nil {
			log.Println("error GetBlockChainInfo:", err)
			env.SayErr(body.Message.Chat.ID, err)
			return
		}
		nodestring := "mio nodo"
		if len(nodo) > 0 {
			nodestring = fmt.Sprintf("nodo \"%s\"", nodo)
		}
		messaggio := fmt.Sprintf("Ecco %s, al %s risulta altezza: %d, epoca: %d/%d (%d)", ChatData.Name, nodestring, bbsd.Result.BeaconHeight, bbsd.Result.Epoch, 350-(bbsd.Result.BeaconHeight%350), bci.Result.BestBlocks["-1"].RemainingBlockEpoch)
		shards := make([]string, 0, len(bbsd.Result.BestShardHeight))
		for shard := range bbsd.Result.BestShardHeight {
			shards = append(shards, shard)
		}
		sort.Strings(shards)
		for _, shard := range shards {
			height := bbsd.Result.BestShardHeight[shard]
			nodeheight := bci.Result.BestBlocks[shard].Height
			messaggio = fmt.Sprintf("%s\nshard %s heigth %d node height %d", messaggio, shard, height, nodeheight)
		}
		if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/addnode":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		urlnodo := ""
		if np < 2 {
			messaggio := fmt.Sprint("Problema sui parametri di addnode, servono [nome] [url] ", len(params), " ", params)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		nodo = params[0]
		urlnodo = params[1]
		log.Println("/addnode", nodo, urlnodo, np, params)
		err := env.Db.UpdateUrlNode(&models.UrlNode{UNId: 0, ChatID: body.Message.Chat.ID, NodeName: nodo, NodeURL: urlnodo})
		if err != nil {
			messaggio := fmt.Sprint("Problema aggiornamento nodo: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprint("Nodo aggiornato: \"", nodo, "\" ", urlnodo)
		if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/listnodes":
		listaNodi, err := env.Db.GetUrlNodes(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando i nodi: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := ""
		for i, urlnodo := range *listaNodi {
			messaggio = fmt.Sprintf("%s\n%d)\t\"%s\"\t%s", messaggio, i+1, urlnodo.NodeName, urlnodo.NodeURL)
		}
		log.Printf("/listnodes invio %d nodi.", len(*listaNodi))
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err = env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/delnode":
		listaNodi, err := env.Db.GetUrlNodes(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando i nodi: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		if len(*listaNodi) == 0 {
			messaggio := fmt.Sprintf("Mi spiace, non hai nodi.")
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		if len(*listaNodi) > 1 && np < 1 {
			messaggio := fmt.Sprintf("Problema sui parametri di delnode, serve [nome] perché hai %d nodi. ", len(*listaNodi))
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		var unid int64
		nodo := "not found"
		if np > 0 {
			for _, urlnodo := range *listaNodi {
				if urlnodo.NodeName == params[0] {
					unid = urlnodo.UNId
					nodo = urlnodo.NodeName
				}
			}
		} else {
			unid = (*listaNodi)[0].UNId
			nodo = (*listaNodi)[0].NodeName
		}
		log.Println("/delnode UNId=", unid, " Nome=", nodo)
		if nodo == "not found" {
			messaggio := fmt.Sprint("Problema cancellando il nodo: ", nodo)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		err = env.Db.DelNode(unid)
		if err != nil {
			messaggio := fmt.Sprint("Problema cancellando il nodo: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprintf("Nodo %s (%d) eliminato.", nodo, unid)
		if err = env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/addkey":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		alias := ""
		pubkey := ""
		if np < 2 {
			messaggio := fmt.Sprint("Problema sui parametri di addkey, servono [alias] [pubkey] ", len(params), " ", params)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		alias = params[0]
		pubkey = params[1]
		log.Println("/addkey", alias, pubkey, np, params)
		err := env.Db.UpdateChatKey(&models.ChatKey{ChatID: body.Message.Chat.ID, KeyAlias: alias, PubKey: pubkey})
		if err != nil {
			messaggio := fmt.Sprint("Problema aggiornamento chiave: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprint("Chiave aggiornata: \"", alias, "\" ", pubkey)
		if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/listkeys":
		listaChiavi, err := env.Db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := ""
		for i, pubkey := range *listaChiavi {
			messaggio = fmt.Sprintf("%s\n%d)\t\"%s\"\t%s", messaggio, i+1, pubkey.KeyAlias, pubkey.PubKey)
		}
		log.Printf("/listkeys invio %d chiavi.", len(*listaChiavi))
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err = env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/delkey":
		listaChiavi, err := env.Db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		if len(*listaChiavi) == 0 {
			messaggio := fmt.Sprintf("Mi spiace, non hai chiavi.")
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		if len(*listaChiavi) > 1 && np < 1 {
			messaggio := fmt.Sprintf("Problema sui parametri di delkey, serve [alias] perché hai %d chiavi. ", len(*listaChiavi))
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		alias := "not found"
		if np > 0 {
			for _, pubkey := range *listaChiavi {
				if pubkey.KeyAlias == params[0] {
					alias = pubkey.KeyAlias
				}
			}
		} else {
			alias = (*listaChiavi)[0].KeyAlias
		}
		log.Println("/delkey ChatId=", body.Message.Chat.ID, " Alias=", alias)
		if alias == "not found" {
			messaggio := fmt.Sprint("Problema cancellando alias chiave: ", alias)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		err = env.Db.DelChatKey(body.Message.Chat.ID, alias)
		if err != nil {
			messaggio := fmt.Sprint("Problema cancellando alias chiave: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprintf("Chiave %s eliminata.", alias)
		if err = env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/status":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		if np > 0 {
			nodo = params[0]
		}
		log.Println("/status", nodo, np, params)
		theUrl := env.DEFAULT_NODE_URL
		if urlNode, err := env.Db.GetUrlNode(ChatData.ChatID, nodo); err == nil {
			theUrl = urlNode.NodeURL
		} else {
			if np > 0 {
				messaggio := fmt.Sprintf("Non trovo tuo nodo \"%s\" uso mio nodo", nodo)
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
			}
		}
		if err := models.GetBeaconBestStateDetail(theUrl, &bbsd); err != nil {
			log.Println("error getBeaconBestStateDetail:", err)
			env.SayErr(body.Message.Chat.ID, err)
			return
		}
		listaChiavi, err := env.Db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := ""
		for _, pubkey := range *listaChiavi {
			status, pki := models.GetPubKeyStatus(&bbsd, pubkey.PubKey)
			mk := &models.MiningKey{
				PubKey:     pubkey.PubKey,
				LastStatus: status,
			}
			if pki != nil { //abbiamo info della chiave
				mk.LastPRV = pki.PRV
				mk.IsAutoStake = pki.IsAutoStake
				mk.Bls = pki.MiningPubKey.Bls
				mk.Dsa = pki.MiningPubKey.Dsa
				mrfmk := models.MRFMK{}
				err := models.GetMinerRewardFromMiningKey(env.DEFAULT_FULLNODE_URL, "bls:"+mk.Bls, &mrfmk)
				if err == nil { //no err, abbiamo anche i PRV
					mk.LastPRV = mrfmk.Result.GetPRV()
				} else { //non abbiamo i PRV
					mk.LastPRV = -1 //segnaliamo che non è da aggiornare
				}
			}
			//messaggio = fmt.Sprintf("%s\n%s %s %fPRV", messaggio, pubkey.KeyAlias, status, float64(mk.LastPRV)/float64(1000000000))
			messaggio = fmt.Sprintf("%s\n%s %s %.9fPRV", messaggio, pubkey.KeyAlias, status, models.BIG_COINS.GetFloat64Val("PRV", mk.LastPRV))

			env.Db.UpdateMiningKey(mk, models.StatusChangeNotifierFunc(env.StatusChanged))
		}
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err = env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}

	case env.StrCmd(body.Message.Text) == "/balance":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		key := ""
		if np > 0 {
			key = params[0]
		}
		log.Println("/balance", key, np, params)
		listaChiavi := &[]models.ChatKey{}
		if key == "" { //chiave non specificata, prendiamo tutte
			var err error
			listaChiavi, err = env.Db.GetChatKeys(body.Message.Chat.ID, 100, 0)
			if err != nil {
				messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
		} else { //chiave selezionata, usiamo quella
			chiave, err := env.Db.GetChatKey(body.Message.Chat.ID, key)
			if err != nil {
				messaggio := fmt.Sprint("Problema recuperando la chiave: alias=", key, " err=", err)
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
			*listaChiavi = append(*listaChiavi, *chiave)
		}
		messaggio := ""
		for _, pubkey := range *listaChiavi {
			mk, errmk := env.Db.GetMiningKey(pubkey.PubKey)
			if errmk == nil { //abbiamo info della chiave
				mrfmk := models.MRFMK{}
				err := models.GetMinerRewardFromMiningKey(env.DEFAULT_FULLNODE_URL, "bls:"+mk.Bls, &mrfmk)
				if err == nil { //no err, abbiamo anche i Saldi
					messaggio = fmt.Sprintf("%s\n%s:\n", messaggio, pubkey.KeyAlias)
					for _, id := range mrfmk.Result.GetValueIDs() {
						coin, val := mrfmk.Result.GetNameValuePair(id)
						messaggio = fmt.Sprintf("%s\t%.9f%s\n", messaggio, models.BIG_COINS.GetFloat64Val(coin, val), coin)
					}
				}
			}
		}
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/notify":
		newNotify := env.Db.ChangeNotify(body.Message.Chat.ID)
		messaggio := fmt.Sprintf("Notify is now %t.", newNotify)
		if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.StrCmd(body.Message.Text) == "/lstickets":
		params := strings.Fields(env.RemoveCmd(body.Message.Text))
		np := len(params)
		var errParse error = nil
		y, m, _ := time.Now().Date()
		var starttm = time.Date(y, m, 1, 0, 0, 0, 0, time.Now().Location())
		if np == 1 {
			starttm, errParse = time.ParseInLocation("2006-01-02 15:04:05", params[0]+"-01 00:00:00", time.Now().Location())
			if errParse != nil {
				log.Println("errParse:", errParse)
				messaggio := fmt.Sprintf("Problems with /lsnotify command params, need aaaa-mm but found '%s'.", env.RemoveCmd(body.Message.Text))
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
		} else if np > 1 {
			messaggio := fmt.Sprintf("Problems with /lsnotify command params, need aaaa-mm but found '%s'.", env.RemoveCmd(body.Message.Text))
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
				return
			}
		}
		period := fmt.Sprintf("%s-%s", strconv.Itoa(starttm.Year()), strconv.Itoa(int(starttm.Month())))

		log.Println("/lstickets", period)

		lotterychats, err := env.Db.GetLotteryIDS(body.Message.Chat.ID)
		if err != nil {
			log.Println("/lsnotify err:", err)
			messaggio := fmt.Sprintf("Problems with /lsnotify GetLotteryIDS '%v'.", err)
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		for _, lotterychat := range lotterychats {
			lottery := env.Db.GetLotteryByKey(lotterychat.LOId)
			messaggio := fmt.Sprintf("Lottery %s.", lottery.LotteryName)
			messaggio = fmt.Sprintf("%s\n*Listing 🎫 of %s.", messaggio, period)
			lotterytickets, err := env.Db.GetLotteryTickets(lotterychat.LOId, starttm, -1)
			if err != nil {
				log.Println("/lsnotify err:", err)
				messaggio := fmt.Sprintf("Problems with /lsnotify GetLotteryTickets '%v'.", err)
				if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
			for _, lotteryticket := range lotterytickets {
				chatkey, err := env.Db.GetChatKeyFromPub(lotterychat.ChatID, lotteryticket.PubKey)
				if err != nil { // we get default description for chatkey
					lotterykey := env.Db.GetLotteryKeyByKey(lotteryticket.LOId, lotteryticket.PubKey)
					chatkey = &models.ChatKey{lotterychat.ChatID, lotterykey.DefaultAlias, lotterykey.PubKey}
				}
				flag := ""
				if lotteryticket.Extracted > 0 {
					flag = fmt.Sprintf("(%d)", lotteryticket.Extracted)
				}
				if lotteryticket.Extracted == 1 {
					flag = "🥇"
				}
				if lotteryticket.Extracted == 2 {
					flag = "🥈"
				}
				if lotteryticket.Extracted == 3 {
					flag = "🥉"
				}
				messaggio = fmt.Sprintf("%s\n  %s %s %s", messaggio, chatkey.KeyAlias, models.GetTSString(lotteryticket.Timestamp), flag)
			}
			if err := env.SayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
				return
			}
		}
	default:
		if err := env.SayText(body.Message.Chat.ID, env.PrintBOT_CMDS()); err != nil {
			log.Println("error in sending reply:", err)
			return
		}

	}

	// log a confirmation message if the message is sent successfully
	log.Printf("reply sent, chat id: %d\n", body.Message.Chat.ID)
}
