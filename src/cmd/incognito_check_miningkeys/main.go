package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
	"log"
	"net/http"
	"os"
)

type Env struct {
	DBFILE           string
	db               *models.DBnode
	TOKEN            string
	API              string
	BOT_NAME         string
	BOT_CMDS         []Cmd
	DEFAULT_NODE_URL string
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
			Cmd{cmd: "/altezza", descr: "`/altezza [nodo]` interroga il [nodo] per informazioni blockchain"},
			Cmd{cmd: "/addnode", descr: "`/addnode [nodo] [urlnodo]` salva o aggiorna url del tuo nodo"},
			Cmd{cmd: "/delnode", descr: "`/delnode [nodo]` elimina il tuo nodo"},
			Cmd{cmd: "/listnodes", descr: "`/listnodes` elenca i tuoi nodi"},
			Cmd{cmd: "/addkey", descr: "`/addkey [alias] [pubkey]` salva o aggiorna public key del tuo miner"},
			Cmd{cmd: "/delkey", descr: "`/delkey [alias]` elimina la public key"},
			Cmd{cmd: "/listkeys", descr: "`/listkeys` elenca le tua public keys"},
			Cmd{cmd: "/status", descr: "`/status [nodo]` elenca lo stato delle tue key di mining"},
		},
		DEFAULT_NODE_URL: "http://95.217.164.210:9334",
	}

	log.Println("DBFILE: " + env.DBFILE)
	db, err := models.NewDB("sqlite3", env.DBFILE)
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", db, db.DB)
	env.db = db

	log.Println("SendMessageUrl: " + env.GetSendMessageUrl())

	var miningkeys *[]models.MiningKey
	miningkeys, err = env.db.GetMiningKeys(100, 0)
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
		status := models.GetPubKeyStatus(&bbsd, miningkey.PubKey)
		mk := &models.MiningKey{
			PubKey:     miningkey.PubKey,
			LastStatus: status,
		}
		env.db.UpdateMiningKey(mk, models.StatusChangeNotifierFunc(env.StatusChanged))
	}
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

//The below code deals with the process of sending a response message
// to the user

// Create a struct to conform to the JSON body
// of the send message request
// https://core.telegram.org/bots/api#sendmessage
type sendMessageReqBody struct {
	ChatID int64  `json:"chat_id"`
	Text   string `json:"text"`
}

func (env *Env) StatusChanged(pubkey, oldstat, newstat string) error {
	log.Printf("Status Changed: %s %s %s", pubkey, oldstat, newstat)
	chatkeys, err := env.db.GetChatKeysByPubKey(pubkey, 100, 0)
	if err != nil {
		return err
	}
	for _, chatkey := range *chatkeys {
		messaggio := fmt.Sprintf("\"%s\" %s -> %s", chatkey.KeyAlias, oldstat, newstat)
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
