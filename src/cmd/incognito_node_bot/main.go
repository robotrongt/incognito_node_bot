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
	}

	err = env.db.CreateTablesIfNotExists()
	if err != nil {
		log.Fatal(err)
	}

	//	ChatData, _ := env.db.GetUserByChatID(1)
	//	fmt.Printf("ChatData: %T", ChatData)
	log.Println("SendMessageUrl: " + env.GetSendMessageUrl())

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
		fmt.Println("could not decode request body", err)
		return
	}
	ChatData, _ := env.db.GetUserByChatID(body.Message.Chat.ID)
	bbsd := models.BBSD{}
	fmt.Println("Ricevuto:", body.Message.Text)
	switch {
	case ChatData.NameAsked:
		ChatData.Name = body.Message.Text
		ChatData.NameAsked = false
		if err := env.db.UpdateUser(ChatData); err != nil {
			fmt.Println("error updating name:", err)
		}
		if err := env.sayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" ora mi ricordo di te!"); err != nil {
			fmt.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "/start"):
		if err := env.sayText(body.Message.Chat.ID, "Ciao "+ChatData.Name+" come ti chiami?"); err != nil {
			fmt.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "fiona") || strings.Contains(strings.ToLower(body.Message.Text), "olindo"):
		if err := env.sayText(body.Message.Chat.ID, env.db.GetFionaText()); err != nil {
			fmt.Println("error in sending reply:", err)
			return
		}
	case strings.Contains(strings.ToLower(body.Message.Text), "altezza"):
		if err := models.GetBeaconBestStateDetail("http://95.217.164.210:9334", ChatData, &bbsd); err != nil {
			fmt.Println("error getBeaconBestStateDetail:", err)
			return
		}
		messaggio := fmt.Sprintf("Ecco %s, al mio nodo risulta altezza: %d, epoca: %d", ChatData.Name, bbsd.Result.BeaconHeight, bbsd.Result.Epoch)
		if err := env.sayText(body.Message.Chat.ID, messaggio); err != nil {
			fmt.Println("error in sending reply:", err)
			return
		}
	default:
		if err := env.sayText(body.Message.Chat.ID, "prova a dire \"altezza\""); err != nil {
			fmt.Println("error in sending reply:", err)
			return
		}

	}

	// log a confirmation message if the message is sent successfully
	fmt.Printf("reply sent, chat id: %d\n", body.Message.Chat.ID)
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
