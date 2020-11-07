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
	"strings"
)

type Env struct {
	db       *models.DBnode
	TOKEN    string
	API      string
	BOT_NAME string
	BOT_CMDS map[string]string
}

func (env *Env) GetSendMessageUrl() string {
	return env.API + env.TOKEN + "/sendMessage"
}

func main() {
	//	var err error
	db, err := models.NewDB("sqlite3", "./incbot.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", db, db.DB)

	env := &Env{
		db:       db,
		TOKEN:    os.Getenv("TOKEN"),
		API:      "https://api.telegram.org/bot",
		BOT_NAME: "@incognito_node_bot",
		BOT_CMDS: map[string]string{
			"/start":     "inizializza il bot",
			"/help":      "elenco comandi bot",
			"/altezza":   "`/altezza [nodo]` interroga il [nodo] per informazioni blockchain",
			"/addnode":   "`/addnode [nodo] [urlnodo]` salva o aggiorna url del tuo nodo",
			"/delnode":   "`/delnode [nodo]` elimina il tuo nodo",
			"/listnodes": "`/listnodes` elenca i tuoi nodi",
		},
	}

	err = env.db.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}
	/*
		u, e := env.db.GetUrlNode(1, "pippo")
		log.Println("env.db.GetUrlNode err:", e)
		log.Println("env.db.GetUrlNode u:", u)
		if e != nil {
			u = &models.UrlNode{}
		}
		u.ChatID = 2
		u.NodeName = "pippo2"
		u.NodeURL = "pippurlx2"
		e = env.db.UpdateUrlNode(u)
		log.Println("env.db.UpdateUrlNode err:", e)
	*/
	//	fmt.Printf("ChatData: %T", ChatData)
	log.Println("SendMessageUrl: " + env.GetSendMessageUrl())
	for cmd, descr := range env.BOT_CMDS {
		log.Printf("cmd=\"%s\" descr=\"%s\"\n", cmd, descr)
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
	case env.strCmd(body.Message.Text) == "/altezza":
		params := strings.Fields(env.removeCmd(body.Message.Text))
		np := len(params)
		nodo := ""
		if np > 0 {
			nodo = params[0]
		}
		log.Println("/altezza", nodo, np, params)
		theUrl := "http://127.0.0.1:9334"
		if urlNode, err := env.db.GetUrlNode(ChatData.ChatID, nodo); err == nil {
			theUrl = urlNode.NodeURL
		}
		if err := models.GetBeaconBestStateDetail(theUrl, ChatData, &bbsd); err != nil {
			log.Println("error getBeaconBestStateDetail:", err)
			return
		}
		nodestring := "mio nodo"
		if len(nodo) > 0 {
			nodestring = fmt.Sprintf("nodo \"%s\"", nodo)
		}
		messaggio := fmt.Sprintf("Ecco %s, al %s risulta altezza: %d, epoca: %d/%d", ChatData.Name, nodestring, bbsd.Result.BeaconHeight, bbsd.Result.Epoch, 350-(bbsd.Result.BeaconHeight%350))
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
	default:
		if err := env.sayText(body.Message.Chat.ID, "prova a dire \"/altezza\""); err != nil {
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
	for cmd := range env.BOT_CMDS {
		cmdbot := cmd + env.BOT_NAME
		if strings.HasPrefix(t, cmdbot) || strings.HasPrefix(t, cmd) {
			return cmd
		}
	}
	return ""
}

//ritorna la stringa dopo aver rimosso il comando o comando@nomebot
func (env *Env) removeCmd(text string) string {
	txt := strings.TrimLeft(text, " ")
	t := strings.ToLower(txt)
	for cmd := range env.BOT_CMDS {
		cmdbot := cmd + env.BOT_NAME
		switch {
		case strings.HasPrefix(t, cmdbot):
			{
				return strings.TrimLeft(txt[len(cmdbot):], " ")
			}
		case strings.HasPrefix(t, cmd):
			{
				return strings.TrimLeft(txt[len(cmd):], " ")
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
