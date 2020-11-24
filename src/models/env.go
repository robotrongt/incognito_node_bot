package models

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Env struct {
	DBFILE               string
	Db                   *DBnode
	TOKEN                string
	TGTOKEN              string
	API                  string
	BOT_NAME             string
	BOT_CMDS             []Cmd
	DEFAULT_NODE_URL     string
	DEFAULT_FULLNODE_URL string
}

type Cmd struct {
	Cmd   string
	Descr string
}

func (env *Env) GetSendMessageUrl() string {
	return env.API + env.TOKEN + "/sendMessage"
}

func NewEnv() *Env {
	env := &Env{
		DBFILE:   os.Getenv("DBFILE"),
		Db:       nil,
		TOKEN:    os.Getenv("TOKEN"),
		TGTOKEN:  os.Getenv("TGTOKEN"),
		API:      "https://api.telegram.org/bot",
		BOT_NAME: "@incognito_node_bot",
		BOT_CMDS: []Cmd{
			Cmd{Cmd: "/start", Descr: "inizializza il bot"},
			Cmd{Cmd: "/help", Descr: "elenco comandi bot"},
			Cmd{Cmd: "/height", Descr: "[nodo]: interroga il [nodo] per informazioni blockchain"},
			Cmd{Cmd: "/addnode", Descr: "[nodo] [urlnodo]: salva o aggiorna url del tuo nodo"},
			Cmd{Cmd: "/delnode", Descr: "[nodo]: elimina il tuo nodo"},
			Cmd{Cmd: "/listnodes", Descr: "elenca i tuoi nodi"},
			Cmd{Cmd: "/addkey", Descr: "[alias] [pubkey]: salva o aggiorna public key del tuo miner"},
			Cmd{Cmd: "/delkey", Descr: "[alias]: elimina la public key"},
			Cmd{Cmd: "/listkeys", Descr: "elenca le tue public keys"},
			Cmd{Cmd: "/status", Descr: "[nodo]: elenca lo stato delle tue key di mining"},
			Cmd{Cmd: "/balance", Descr: "[alias_chiave]: reward accurato della chiave di mining"},
			Cmd{Cmd: "/notify", Descr: "turns notifications off or on"},
			Cmd{Cmd: "/lstickets", Descr: "[aaaa-mm] lists all lottery tickets"},
		},
		DEFAULT_NODE_URL:     os.Getenv("DEFAULT_NODE_URL"),
		DEFAULT_FULLNODE_URL: os.Getenv("DEFAULT_FULLNODE_URL"),
	}
	log.Println("DBFILE: " + env.DBFILE)
	db, err := NewDB("sqlite3", env.DBFILE)
	if err != nil {
		log.Fatal(err)
	}
	env.Db = db

	log.Println("SendMessageUrl: " + env.GetSendMessageUrl())
	return env
}

//ritorna il comando,se presente, e senza @nomebot tutto minuscolo. Altrimenti stringa vuota
func (env *Env) StrCmd(text string) string {
	t := strings.ToLower(strings.TrimLeft(text, " "))
	for _, cmd := range env.BOT_CMDS {
		cmdbot := cmd.Cmd + env.BOT_NAME
		if strings.HasPrefix(t, cmdbot) || strings.HasPrefix(t, cmd.Cmd) {
			return cmd.Cmd
		}
	}
	return ""
}

//ritorna la stringa dopo aver rimosso il comando o comando@nomebot
func (env *Env) RemoveCmd(text string) string {
	txt := strings.TrimLeft(text, " ")
	t := strings.ToLower(txt)
	for _, cmd := range env.BOT_CMDS {
		cmdbot := cmd.Cmd + env.BOT_NAME
		switch {
		case strings.HasPrefix(t, cmdbot):
			{
				return strings.TrimLeft(txt[len(cmdbot):], " ")
			}
		case strings.HasPrefix(t, cmd.Cmd):
			{
				return strings.TrimLeft(txt[len(cmd.Cmd):], " ")
			}
		}
	}
	return txt
}

func (env *Env) NotifyTicket(loid, ts int64, chatuser *ChatUser, chatkey *ChatKey) error {
	tmstring := GetTSString(ts)
	lottery := env.Db.GetLotteryByKey(loid)
	log.Printf("%s \"%s\" %t->New Ticket for %s %s\n", lottery.LotteryName, chatuser.Name, chatuser.Notify, chatkey.KeyAlias, tmstring)
	messaggio := fmt.Sprintf("Lottery %s\n*🎫 %s->%s", lottery.LotteryName, chatkey.KeyAlias, tmstring)
	if err := env.SayText(chatkey.ChatID, messaggio); err != nil {
		log.Println("error in sending reply:", err)
	}
	return nil
}

func (env *Env) StatusChanged(miningkey *MiningKey, oldstat string, oldprv int64) error {
	pubkey := miningkey.PubKey
	newstat := miningkey.LastStatus
	newprv := miningkey.LastPRV
	log.Printf("Status Changed: %s %s %s %.9fPRV %.9fPRV", pubkey, oldstat, newstat, BIG_COINS.GetFloat64Val("PRV", newprv), BIG_COINS.GetFloat64Val("PRV", oldprv))
	icons := []string{"🥳", "👍", "😇", "🤑", "🙌", "💰", "💶", "💵", "💸"}
	i := rand.Intn(len(icons))

	newst := strings.ToLower(strings.TrimLeft(newstat, " "))
	if strings.HasPrefix(newst, "committe") && newprv >= oldprv { // this is a new round
		var tm = time.Now()
		ts := MakeTSFromTime(tm)
		lotterykeys, err := env.Db.AddLotteryTickets(ts, pubkey)
		if err != nil {
			log.Println("Status Changed: error AddLotteryTickets:", err)
		}
		err = env.Db.NotifyAllLotteryUsersTicket(ts, lotterykeys, env.NotifyTicket)
		if err != nil {
			log.Println("Status Changed: error NotifyAllLotteryUsersTicket:", err)
		}
	}
	chatkeys, err := env.Db.GetChatKeysByPubKey(pubkey, 100, 0)
	if err != nil {
		return err
	}
	for _, chatkey := range *chatkeys {
		messaggio := fmt.Sprintf("\"%s\" %s -> %s%s %.9fPRV", chatkey.KeyAlias, oldstat, newstat, icons[i], BIG_COINS.GetFloat64Val("PRV", newprv))
		if env.Db.GetNotify(chatkey.ChatID) {
			log.Printf("Notify chat: %d %s", chatkey.ChatID, messaggio)
			if err = env.SayText(chatkey.ChatID, messaggio); err != nil {
				log.Println("error in sending reply:", err)
			}
		} else {
			log.Printf("Notify off for chat: %d (%s)", chatkey.ChatID, messaggio)

		}

	}
	return err
}

func (env *Env) PrintBOT_CMDS() string {
	text := "Prova questi comandi:"
	for _, cmd := range env.BOT_CMDS {
		text = fmt.Sprintf("%s\n%s\t%s", text, cmd.Cmd, cmd.Descr)
	}

	return text
}
