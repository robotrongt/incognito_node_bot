package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

type Env struct {
	DBFILE               string
	db                   *models.DBnode
	TOKEN                string
	API                  string
	BOT_NAME             string
	BOT_CMDS             []Cmd
	DEFAULT_NODE_URL     string
	DEFAULT_FULLNODE_URL string
}

type Cmd struct {
	cmd   string
	descr string
}

func (env *Env) GetSendMessageUrl() string {
	return env.API + env.TOKEN + "/sendMessage"
}

func main() {
	env := &Env{
		DBFILE:   os.Getenv("DBFILE"),
		db:       nil,
		TOKEN:    os.Getenv("TOKEN"),
		API:      "https://api.telegram.org/bot",
		BOT_NAME: "@incognito_node_bot",
		BOT_CMDS: []Cmd{
			Cmd{cmd: "/start", descr: "inizializza il bot"},
			Cmd{cmd: "/help", descr: "elenco comandi bot"},
			Cmd{cmd: "/height", descr: "[nodo]: interroga il [nodo] per informazioni blockchain"},
			Cmd{cmd: "/addnode", descr: "[nodo] [urlnodo]: salva o aggiorna url del tuo nodo"},
			Cmd{cmd: "/delnode", descr: "[nodo]: elimina il tuo nodo"},
			Cmd{cmd: "/listnodes", descr: "elenca i tuoi nodi"},
			Cmd{cmd: "/addkey", descr: "[alias] [pubkey]: salva o aggiorna public key del tuo miner"},
			Cmd{cmd: "/delkey", descr: "[alias]: elimina la public key"},
			Cmd{cmd: "/listkeys", descr: "elenca le tue public keys"},
			Cmd{cmd: "/status", descr: "[nodo]: elenca lo stato delle tue key di mining"},
			Cmd{cmd: "/balance", descr: "[alias_chiave]: reward accurato della chiave di mining"},
		},
		DEFAULT_NODE_URL:     os.Getenv("DEFAULT_NODE_URL"),
		DEFAULT_FULLNODE_URL: os.Getenv("DEFAULT_FULLNODE_URL"),
	}
	db, err := models.NewDB("sqlite3", env.DBFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", db, db.DB)
	env.db = db

	err = env.db.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("SendMessageUrl: " + env.GetSendMessageUrl())

	rand.Seed(time.Now().UnixNano())

	for _, cmd := range env.BOT_CMDS {
		log.Printf("%s - %s\n", cmd.cmd, cmd.descr)
	}

	http.ListenAndServeTLS(":8443", "cert.pem", "key.pem", http.HandlerFunc(env.Handler))
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

// This handler is called everytime telegram sends us a webhook event
func (env *Env) Handler(res http.ResponseWriter, req *http.Request) {
	// First, decode the JSON response body
	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		log.Println("could not decode request body", err)
		return
	}
	ChatData, _ := env.db.GetUserByChatID(body.Message.Chat.ID)
	bbsd := models.BBSD{}
	bci := models.BCI{}
	log.Println("Ricevuto:", body.Message.Text)
	switch {
	case env.strCmd(body.Message.Text) == "/start":
		ChatData.NameAsked = true
		if err := env.db.UpdateUser(ChatData); err != nil {
			log.Println("error updating name:", err)
		}
		if err := env.sayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" come ti chiami?"); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case ChatData.NameAsked:
		ChatData.Name = body.Message.Text
		ChatData.NameAsked = false
		if err := env.db.UpdateUser(ChatData); err != nil {
			log.Println("error updating name:", err)
		}
		if err := env.sayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" ora mi ricordo di te!"); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "fiona") || strings.Contains(strings.ToLower(body.Message.Text), "olindo"):
		if err := env.sayText(body.Message.Chat.ID, env.db.GetFionaText()); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/height":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		if np > 0 {
			nodo = params[0]
		}
		log.Println("/height", nodo, np, params)
		theUrl := env.DEFAULT_NODE_URL
		if urlNode, err := env.db.GetUrlNode(ChatData.ChatID, nodo); err == nil {
			theUrl = urlNode.NodeURL
		} else {
			if np > 0 {
				messaggio := fmt.Sprintf("Non trovo tuo nodo \"%s\" uso mio nodo", nodo)
				if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				nodo = ""
			}
		}
		if err := models.GetBeaconBestStateDetail(theUrl, &bbsd); err != nil {
			log.Println("error getBeaconBestStateDetail:", err)
			env.sayErr(body.Message.Chat.ID, err)
			return
		}
		if err := models.GetBlockChainInfo(theUrl, &bci); err != nil {
			log.Println("error GetBlockChainInfo:", err)
			env.sayErr(body.Message.Chat.ID, err)
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
		if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/addnode":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		urlnodo := ""
		if np < 2 {
			messaggio := fmt.Sprint("Problema sui parametri di addnode, servono [nome] [url] ", len(params), " ", params)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		nodo = params[0]
		urlnodo = params[1]
		log.Println("/addnode", nodo, urlnodo, np, params)
		err := env.db.UpdateUrlNode(&models.UrlNode{UNId: 0, ChatID: body.Message.Chat.ID, NodeName: nodo, NodeURL: urlnodo})
		if err != nil {
			messaggio := fmt.Sprint("Problema aggiornamento nodo: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprint("Nodo aggiornato: \"", nodo, "\" ", urlnodo)
		if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/listnodes":
		listaNodi, err := env.db.GetUrlNodes(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando i nodi: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
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
		if err = env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/delnode":
		listaNodi, err := env.db.GetUrlNodes(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando i nodi: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		if len(*listaNodi) == 0 {
			messaggio := fmt.Sprintf("Mi spiace, non hai nodi.")
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		if len(*listaNodi) > 1 && np < 1 {
			messaggio := fmt.Sprintf("Problema sui parametri di delnode, serve [nome] perché hai %d nodi. ", len(*listaNodi))
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
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
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		err = env.db.DelNode(unid)
		if err != nil {
			messaggio := fmt.Sprint("Problema cancellando il nodo: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprintf("Nodo %s (%d) eliminato.", nodo, unid)
		if err = env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/addkey":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		alias := ""
		pubkey := ""
		if np < 2 {
			messaggio := fmt.Sprint("Problema sui parametri di addkey, servono [alias] [pubkey] ", len(params), " ", params)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		alias = params[0]
		pubkey = params[1]
		log.Println("/addkey", alias, pubkey, np, params)
		err := env.db.UpdateChatKey(&models.ChatKey{ChatID: body.Message.Chat.ID, KeyAlias: alias, PubKey: pubkey})
		if err != nil {
			messaggio := fmt.Sprint("Problema aggiornamento chiave: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprint("Chiave aggiornata: \"", alias, "\" ", pubkey)
		if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/listkeys":
		listaChiavi, err := env.db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
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
		if err = env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/delkey":
		listaChiavi, err := env.db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		if len(*listaChiavi) == 0 {
			messaggio := fmt.Sprintf("Mi spiace, non hai chiavi.")
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		if len(*listaChiavi) > 1 && np < 1 {
			messaggio := fmt.Sprintf("Problema sui parametri di delkey, serve [alias] perché hai %d chiavi. ", len(*listaChiavi))
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
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
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		err = env.db.DelChatKey(body.Message.Chat.ID, alias)
		if err != nil {
			messaggio := fmt.Sprint("Problema cancellando alias chiave: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
			return
		}
		messaggio := fmt.Sprintf("Chiave %s eliminata.", alias)
		if err = env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}
	case env.strCmd(body.Message.Text) == "/status":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		if np > 0 {
			nodo = params[0]
		}
		log.Println("/status", nodo, np, params)
		theUrl := env.DEFAULT_NODE_URL
		if urlNode, err := env.db.GetUrlNode(ChatData.ChatID, nodo); err == nil {
			theUrl = urlNode.NodeURL
		} else {
			if np > 0 {
				messaggio := fmt.Sprintf("Non trovo tuo nodo \"%s\" uso mio nodo", nodo)
				if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
			}
		}
		if err := models.GetBeaconBestStateDetail(theUrl, &bbsd); err != nil {
			log.Println("error getBeaconBestStateDetail:", err)
			env.sayErr(body.Message.Chat.ID, err)
			return
		}
		listaChiavi, err := env.db.GetChatKeys(body.Message.Chat.ID, 100, 0)
		if err != nil {
			messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
			if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
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
			messaggio = fmt.Sprintf("%s\n%s %s %fPRV", messaggio, pubkey.KeyAlias, status, models.BIG_COINS.GetFloat64Val("PRV", mk.LastPRV))

			env.db.UpdateMiningKey(mk, models.StatusChangeNotifierFunc(env.StatusChanged))
		}
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err = env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}

	case env.strCmd(body.Message.Text) == "/balance":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		key := ""
		if np > 0 {
			key = params[0]
		}
		log.Println("/balance", key, np, params)
		listaChiavi := &[]models.ChatKey{}
		if key == "" { //chiave non specificata, prendiamo tutte
			var err error
			listaChiavi, err = env.db.GetChatKeys(body.Message.Chat.ID, 100, 0)
			if err != nil {
				messaggio := fmt.Sprint("Problema recuperando le chiavi: ", err)
				if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
		} else { //chiave selezionata, usiamo quella
			chiave, err := env.db.GetChatKey(body.Message.Chat.ID, key)
			if err != nil {
				messaggio := fmt.Sprint("Problema recuperando la chiave: alias=", key, " err=", err)
				if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
					log.Println("error in sending reply:", err)
				}
				return
			}
			*listaChiavi = append(*listaChiavi, *chiave)
		}
		messaggio := ""
		for _, pubkey := range *listaChiavi {
			mk, errmk := env.db.GetMiningKey(pubkey.PubKey)
			if errmk == nil { //abbiamo info della chiave
				mrfmk := models.MRFMK{}
				err := models.GetMinerRewardFromMiningKey(env.DEFAULT_FULLNODE_URL, "bls:"+mk.Bls, &mrfmk)
				if err == nil { //no err, abbiamo anche i Saldi
					messaggio = fmt.Sprintf("%s\n%s:\n", messaggio, pubkey.KeyAlias)
					for _, id := range mrfmk.Result.GetValueIDs() {
						coin, val := mrfmk.Result.GetNameValuePair(id)
						mk.LastPRV = mrfmk.Result.GetPRV()
						messaggio = fmt.Sprintf("%s\t%f%s\n", messaggio, val, coin)
					}
				}
			}
		}
		if messaggio == "" {
			messaggio = "Non trovo nulla!"
		}
		if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
			return
		}

	default:
		if err := env.sayText(body.Message.Chat.ID, env.printBOT_CMDS()); err != nil {
			log.Println("error in sending reply:", err)
			return
		}

	}

	// log a confirmation message if the message is sent successfully
	log.Printf("reply sent, chat id: %d\n", body.Message.Chat.ID)
}

//ritorna il comando,se presente, e senza @nomebot tutto minuscolo. Altrimenti stringa vuota
func (env *Env) strCmd(text string) string {
	t := strings.ToLower(strings.TrimLeft(text, " "))
	for _, cmd := range env.BOT_CMDS {
		cmdbot := cmd.cmd + env.BOT_NAME
		if strings.HasPrefix(t, cmdbot) || strings.HasPrefix(t, cmd.cmd) {
			return cmd.cmd
		}
	}
	return ""
}

//ritorna la stringa dopo aver rimosso il comando o comando@nomebot
func (env *Env) removeCmd(text string) string {
	txt := strings.TrimLeft(text, " ")
	t := strings.ToLower(txt)
	for _, cmd := range env.BOT_CMDS {
		cmdbot := cmd.cmd + env.BOT_NAME
		switch {
		case strings.HasPrefix(t, cmdbot):
			{
				return strings.TrimLeft(txt[len(cmdbot):], " ")
			}
		case strings.HasPrefix(t, cmd.cmd):
			{
				return strings.TrimLeft(txt[len(cmd.cmd):], " ")
			}
		}
	}
	return txt
}

//The below code deals with the process of sending a response message
// to the user

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func (env *Env) StatusChanged(miningkey *models.MiningKey, oldstat string, oldprv int64) error {
	pubkey := miningkey.PubKey
	newstat := miningkey.LastStatus
	newprv := miningkey.LastPRV
	log.Printf("Status Changed: %s %s %s %fPRV %fPRV", pubkey, oldstat, newstat, float64(newprv)/float64(1000000000), float64(oldprv)/float64(1000000000))
	icons := []string{"🥳", "👍", "😇", "🤑", "🙌", "💰", "💶", "💵", "💸"}
	i := rand.Intn(len(icons))
	chatkeys, err := env.db.GetChatKeysByPubKey(pubkey, 100, 0)
	if err != nil {
		return err
	}
	for _, chatkey := range *chatkeys {
		messaggio := fmt.Sprintf("\"%s\" %s -> %s%s %fPRV", chatkey.KeyAlias, oldstat, newstat, icons[i], float64(newprv)/float64(1000000000))
		log.Printf("Notify chat: %d %s", chatkey.ChatID, messaggio)
		if err = env.sayText(chatkey.ChatID, messaggio); err != nil {
			log.Println("error in sending reply:", err)
		}

	}
	return err
}

func (env *Env) sayText(chatID int64, text string) error {
	// Create the request body struct
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
	}
	log.Printf("sayText: %s\n", text)
	// Create the JSON body from the struct
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	// Send a post request with your token
	res, err := http.Post(env.GetSendMessageUrl(), "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}

	return nil
}

func (env *Env) sayErr(chatID int64, err error) error {
	text := fmt.Sprintf("%s", err)
	return env.sayText(chatID, text)
}

func (env *Env) printBOT_CMDS() string {
	text := "Prova questi comandi:"
	for _, cmd := range env.BOT_CMDS {
		text = fmt.Sprintf("%s\n%s\t%s", text, cmd.cmd, cmd.descr)
	}

	return text
}
