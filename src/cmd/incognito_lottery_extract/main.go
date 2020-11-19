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
	"time"
)

type Env struct {
	DBFILE               string
	db                   *models.DBnode
	TOKEN                string
	TGTOKEN              string
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

type Ticket struct {
	Nodo      string
	Timestamp string
}

func (t Ticket) String() string {
	return fmt.Sprintf("{%s, %s}", t.Nodo, t.Timestamp)
}

type Tickets []Ticket

func main() {
	env := &Env{
		DBFILE:   os.Getenv("DBFILE"),
		db:       nil,
		TOKEN:    os.Getenv("TOKEN"),
		TGTOKEN:  os.Getenv("TGTOKEN"),
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
			Cmd{cmd: "/notify", descr: "turns notifications off or on"},
		},
		DEFAULT_NODE_URL:     os.Getenv("DEFAULT_NODE_URL"),
		DEFAULT_FULLNODE_URL: os.Getenv("DEFAULT_FULLNODE_URL"),
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

	rand.Seed(time.Now().UnixNano())

	var theTickets = Tickets{
		Ticket{Timestamp: "2020-11-05 08:44:00", Nodo: "12PQD7LdaLueRDcW7KkbEESQ2vccXqYDPbcPrqh6vd76Fs2aUPS"},
		Ticket{Timestamp: "2020-11-05 12:44:00", Nodo: "12PQD7LdaLueRDcW7KkbEESQ2vccXqYDPbcPrqh6vd76Fs2aUPS"},
		Ticket{Timestamp: "2020-11-05 16:44:00", Nodo: "12PQD7LdaLueRDcW7KkbEESQ2vccXqYDPbcPrqh6vd76Fs2aUPS"},
		Ticket{Timestamp: "2020-11-11 00:21:00", Nodo: "12Vfyz2sR655unUsW3HychxCkx8xhcmwPRZNz9FbGW9zZvTojpW"},
		Ticket{Timestamp: "2020-11-11 04:15:00", Nodo: "12Vfyz2sR655unUsW3HychxCkx8xhcmwPRZNz9FbGW9zZvTojpW"},
		Ticket{Timestamp: "2020-11-11 08:09:00", Nodo: "12Vfyz2sR655unUsW3HychxCkx8xhcmwPRZNz9FbGW9zZvTojpW"},
		Ticket{Timestamp: "2020-11-13 10:43:00", Nodo: "12m5Lh5G8TfiVzsU61cGgwHbRbQrXCik4m1ZRrHTQbLEhaCDiCE"},
		Ticket{Timestamp: "2020-11-13 14:39:00", Nodo: "12m5Lh5G8TfiVzsU61cGgwHbRbQrXCik4m1ZRrHTQbLEhaCDiCE"},
		Ticket{Timestamp: "2020-11-16 20:28:00", Nodo: "12PQD7LdaLueRDcW7KkbEESQ2vccXqYDPbcPrqh6vd76Fs2aUPS"},
		Ticket{Timestamp: "2020-11-17 00:21:00", Nodo: "12PQD7LdaLueRDcW7KkbEESQ2vccXqYDPbcPrqh6vd76Fs2aUPS"},
		Ticket{Timestamp: "2020-11-17 08:08:00", Nodo: "12m5Lh5G8TfiVzsU61cGgwHbRbQrXCik4m1ZRrHTQbLEhaCDiCE"},
		Ticket{Timestamp: "2020-11-17 12:02:00", Nodo: "12m5Lh5G8TfiVzsU61cGgwHbRbQrXCik4m1ZRrHTQbLEhaCDiCE"},
		Ticket{Timestamp: "2020-11-18 03:34:00", Nodo: "1TLQUkQP4ER3zkU2xttjmyit9L5mQT9evCxmEFWnF9pTTPWqKt"},
		Ticket{Timestamp: "2020-11-18 07:28:00", Nodo: "1TLQUkQP4ER3zkU2xttjmyit9L5mQT9evCxmEFWnF9pTTPWqKt"},
	}
	var tmLocalParse = "2006-01-02 15:04:05"
	loc := time.Now().Location()
	for _, ticket := range theTickets {
		var tm, errParse = time.ParseInLocation(tmLocalParse, ticket.Timestamp, loc)
		if errParse != nil {
			log.Println("error parsing time:", err)
			return
		}
		ts := models.MakeTSFromTime(tm)
		lotterykeys, err := db.AddLotteryTickets(ts, ticket.Nodo)
		if err != nil {
			log.Println("error AddLotteryTickets:", err)
			return
		}
		err = db.NotifyAllLotteryUsersTicket(ts, lotterykeys, env.NotifyTicket)
		if err != nil {
			log.Println("error NotifyAllLotteryUsersTicket:", err)
			return
		}

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

func (env *Env) NotifyTicket(ts int64, chatuser *models.ChatUser, chatkey *models.ChatKey) error {
	tmstring := models.GetTSString(ts)
	log.Printf("%d \"%s\" %t->New Ticket for %s %s\n", chatuser.ChatID, chatuser.Name, chatuser.Notify, chatkey.KeyAlias, tmstring)
	return nil
}
func (env *Env) StatusChanged(miningkey *models.MiningKey, oldstat string, oldprv int64) error {
	pubkey := miningkey.PubKey
	newstat := miningkey.LastStatus
	newprv := miningkey.LastPRV
	log.Printf("Status Changed: %s %s %s %fPRV %fPRV", pubkey, oldstat, newstat, models.BIG_COINS.GetFloat64Val("PRV", newprv), models.BIG_COINS.GetFloat64Val("PRV", oldprv))
	icons := []string{"🥳", "👍", "😇", "🤑", "🙌", "💰", "💶", "💵", "💸"}
	i := rand.Intn(len(icons))
	chatkeys, err := env.db.GetChatKeysByPubKey(pubkey, 100, 0)
	if err != nil {
		return err
	}
	for _, chatkey := range *chatkeys {
		messaggio := fmt.Sprintf("\"%s\" %s -> %s%s %fPRV", chatkey.KeyAlias, oldstat, newstat, icons[i], models.BIG_COINS.GetFloat64Val("PRV", newprv))
		if env.db.GetNotify(chatkey.ChatID) {
			log.Printf("Notify chat: %d %s", chatkey.ChatID, messaggio)
			if err = env.sayText(chatkey.ChatID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
		} else {
			log.Printf("Notify off for chat: %d (%s)", chatkey.ChatID, messaggio)

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
