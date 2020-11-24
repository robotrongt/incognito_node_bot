package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/robotrongt/incognito_node_bot/src/models"
)

type Ticket struct {
	Nodo      string
	Timestamp string
}

func (t Ticket) String() string {
	return fmt.Sprintf("{%s, %s}", t.Nodo, t.Timestamp)
}

type Tickets []Ticket

func main() {
	env := models.NewEnv()
	defer env.Db.DB.Close()
	defer log.Println("Exiting...")
	defer log.Printf("%T %T\n", env.Db, env.Db.DB)

	rand.Seed(time.Now().UnixNano())

	var tmLocalParse = "2006-01-02 15:04:05"
	loc := time.Now().Location()
	var tmExtract, err = time.ParseInLocation(tmLocalParse, "2020-11-01", loc)
	if err != nil {
		log.Println("error parsing time:", err)
		return
	}
	tsExtract := models.MakeTSFromTime(tmExtract)
	log.Println("tmExtract:", tmExtract)
	log.Println("tsExtract:", tsExtract)

}
