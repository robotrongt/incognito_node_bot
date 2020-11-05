package models

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

type DBnode struct {
	DB *sql.DB
}

func NewDB(driverName, dataSourceName string) (*DBnode, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	return &DBnode{db}, nil
}

type ChatUser struct {
	ChatID    int64
	Name      string
	NameAsked bool
}

//Recupera un record utente o lo crea vuoto se non esiste
func (db *DBnode) GetUserByChatID(chatID int64) (*ChatUser, error) {
	retVal := &ChatUser{chatID, "", true}

	stmt, err := db.DB.Prepare("select Name, NameAsked from chatdata where ChatID = ?")
	if err != nil {
		fmt.Println("GetUserByChatID error:", err)
		return nil, err
		//		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	var nameasked bool
	err = stmt.QueryRow(chatID).Scan(&name, &nameasked)
	if err != nil {
		retVal, err = db.CreatetUserByChatID(chatID)
	} else {
		retVal.Name = name
		retVal.NameAsked = nameasked
	}
	fmt.Println(retVal.Name, retVal.NameAsked)

	return retVal, err
}

//Crea un utente vuoto
func (db *DBnode) CreatetUserByChatID(chatID int64) (*ChatUser, error) {
	retVal := ChatUser{
		ChatID:    chatID,
		Name:      "Sconosciuto",
		NameAsked: true,
	}

	stmt, err := db.DB.Prepare("insert into chatdata(ChatID, Name, NameAsked) values (?, ?, ?)")
	if err != nil {
		fmt.Println("CreatetUserByChatID error:", err)
		return nil, err
		//		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(retVal.ChatID, retVal.Name, retVal.NameAsked)
	if err != nil {
		fmt.Println("CreatetUserByChatID error:", err)
	}

	return &retVal, err
}

func (db *DBnode) UpdateUser(user *ChatUser) error {
	stmt, err := db.DB.Prepare("UPDATE chatdata SET Name = ?, NameAsked = ? WHERE ChatID = ?")
	if err != nil {
		fmt.Println("UpdateUser error:", err)
		return err
		//		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Name, user.NameAsked, user.ChatID)
	if err != nil {
		fmt.Println("UpdateUser error:", err)
	}

	return err
}

func (db *DBnode) GetUsersList(limit, offset int) (*[]ChatUser, error) {
	stmt, err := db.DB.Prepare("SELECT ChatID, Name, NameAsked FROM chatdata LIMIT ? OFFSET ?")
	if err != nil {
		fmt.Println("GetUsersList error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(limit, offset)
	if err != nil {
		fmt.Println("GetUsersList error:", err)
		return nil, err
	}
	defer rows.Close()
	var chatusers []ChatUser
	for rows.Next() {
		var chatid int64
		var name string
		var nameasked bool
		err = rows.Scan(&chatid, &name, &nameasked)
		if err != nil {
			fmt.Println("GetUsersList error:", err)
			return nil, err
		}

		fmt.Println(chatid, name, nameasked)
		chatusers = append(chatusers, ChatUser{ChatID: chatid, Name: name, NameAsked: nameasked})
	}
	if err := rows.Err(); err != nil {
		fmt.Println("GetUsersList error:", err)
		return nil, err
	}

	return &chatusers, err
}

func (db *DBnode) GetFionaText() string {
	m := map[string]string{
		"2020-11-01": "-53 (-36)",
		"2020-11-02": "-52 (-36)",
		"2020-11-03": "-51 (-35)",
		"2020-11-04": "-50 (-34)",
		"2020-11-05": "-49 (-33)",
		"2020-11-06": "-48 (-32)",
		"2020-11-07": "-47 (-31)",
		"2020-11-08": "-46 (-31)",
		"2020-11-09": "-45 (-31)",
		"2020-11-10": "-44 (-30)",
		"2020-11-11": "-43 (-29)",
		"2020-11-12": "-42 (-28)",
		"2020-11-13": "-41 (-27)",
		"2020-11-14": "-40 (-26)",
		"2020-11-15": "-39 (-26)",
		"2020-11-16": "-38 (-26)",
		"2020-11-17": "-37 (-25)",
		"2020-11-18": "-36 (-24)",
		"2020-11-19": "-35 (-23)",
		"2020-11-20": "-34 (-22)",
		"2020-11-21": "-33 (-21)",
		"2020-11-22": "-32 (-21)",
		"2020-11-23": "-31 (-21)",
		"2020-11-24": "-30 (-20)",
		"2020-11-25": "-29 (-19)",
		"2020-11-26": "-28 (-18)",
		"2020-11-27": "-27 (-17)",
		"2020-11-28": "-26 (-16)",
		"2020-11-29": "-25 (-16)",
		"2020-11-30": "-24 (-16)",
		"2020-12-01": "-23 (-15)",
		"2020-12-02": "-22 (-14)",
		"2020-12-03": "-21 (-13)",
		"2020-12-04": "-20 (-12)",
		"2020-12-05": "-19 (-11)",
		"2020-12-06": "-18 (-11)",
		"2020-12-07": "-17 non facevano ponte?",
		"2020-12-08": "-16 (-11)",
		"2020-12-09": "-15 (-11)",
		"2020-12-10": "-14 (-10)",
		"2020-12-11": "-13 (-9)",
		"2020-12-12": "-12 (-8)",
		"2020-12-13": "-11  (-8)",
		"2020-12-14": "-10 (-8)",
		"2020-12-15": "-9 (-7)",
		"2020-12-16": "-8 (-6)",
		"2020-12-17": "-7 (-5)",
		"2020-12-18": "-6 (-4)",
		"2020-12-19": "-5 (-4)",
		"2020-12-20": "-4 (-4)",
		"2020-12-21": "-3 (-3)",
		"2020-12-22": "-2 (-2)",
		"2020-12-23": "-1 (-1)",
		"2020-12-24": "Boom!",
		"2020-12-25": "sono ancora qui?",
		"2020-12-26": "ma non erano in ferie?",
		"2020-12-27": "sono ancora qui?",
		"2020-12-28": "ma non erano in ferie?",
		"2020-12-29": "sono ancora qui?",
		"2020-12-30": "ma non erano in ferie?",
		"2020-12-31": "Buon anno!",
	}
	if val, ok := m[time.Now().Format("2006-01-02")]; ok {
		return val
	} else {
		return "Vi mancano tanto??"
	}
}

func (db *DBnode) CreateTablesIfNotExists() error {
	/*
	   CREATE TABLE IF NOT EXISTS "chatdata" ( "ChatID" integer NOT NULL, "Name" text, "NameAsked" INTEGER DEFAULT 1, PRIMARY KEY("ChatID") )
	*/
	_, err := db.DB.Exec(`CREATE TABLE IF NOT EXISTS "chatdata" ( "ChatID" integer NOT NULL, "Name" text, "NameAsked" INTEGER DEFAULT 1, PRIMARY KEY("ChatID") )`)
	return err
}
