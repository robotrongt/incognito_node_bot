package models

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
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

type UrlNode struct {
	UNId     int64
	ChatID   int64
	NodeName string
	NodeURL  string
}

type ChatKey struct {
	ChatID   int64
	KeyAlias string
	PubKey   string
}

type MiningKey struct {
	PubKey      string
	LastStatus  string
	LastPRV     int64
	IsAutoStake bool
	Bls         string
	Dsa         string
}

type StatusChangeNotifierFunc func(miningkey *MiningKey, oldstatus string, oldprv int64) error

//Recupera un record utente o lo crea vuoto se non esiste
func (db *DBnode) GetUserByChatID(chatID int64) (*ChatUser, error) {
	retVal := &ChatUser{chatID, "", true}

	stmt, err := db.DB.Prepare("select Name, NameAsked from chatdata where ChatID = ?")
	if err != nil {
		log.Println("GetUserByChatID error:", err)
		return nil, err
		//		log.Fatal(err)
	}
	defer stmt.Close()
	var name string
	var nameasked bool
	err = stmt.QueryRow(chatID).Scan(&name, &nameasked)
	if err != nil {
		retVal, err = db.CreateUserByChatID(chatID)
	} else {
		retVal.Name = name
		retVal.NameAsked = nameasked
	}
	log.Println(retVal.Name, retVal.NameAsked)

	return retVal, err
}

//Crea un utente vuoto
func (db *DBnode) CreateUserByChatID(chatID int64) (*ChatUser, error) {
	retVal := ChatUser{
		ChatID:    chatID,
		Name:      "Sconosciuto",
		NameAsked: true,
	}

	log.Println("CreateUserByChatID:", retVal.ChatID, retVal.Name, retVal.NameAsked)
	stmt, err := db.DB.Prepare("insert into chatdata(ChatID, Name, NameAsked) values (?, ?, ?)")
	if err != nil {
		log.Println("CreateUserByChatID error:", err)
		return nil, err
		//		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(retVal.ChatID, retVal.Name, retVal.NameAsked)
	if err != nil {
		log.Println("CreateUserByChatID error:", err)
	}

	return &retVal, err
}

//Aggiorna utente
func (db *DBnode) UpdateUser(user *ChatUser) error {
	log.Println("UpdateUser:", user.ChatID, user.Name, user.NameAsked)
	stmt, err := db.DB.Prepare("UPDATE chatdata SET Name = ?, NameAsked = ? WHERE ChatID = ?")
	if err != nil {
		log.Println("UpdateUser error:", err)
		return err
		//		log.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(user.Name, user.NameAsked, user.ChatID)
	if err != nil {
		log.Println("UpdateUser error:", err)
	}

	return err
}

//Recupera lista utenti
func (db *DBnode) GetUsersList(limit, offset int) (*[]ChatUser, error) {
	stmt, err := db.DB.Prepare("SELECT ChatID, Name, NameAsked FROM chatdata LIMIT ? OFFSET ?")
	if err != nil {
		log.Println("GetUsersList error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(limit, offset)
	if err != nil {
		log.Println("GetUsersList error:", err)
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
			log.Println("GetUsersList error:", err)
			return nil, err
		}

		log.Println(chatid, name, nameasked)
		chatusers = append(chatusers, ChatUser{ChatID: chatid, Name: name, NameAsked: nameasked})
	}
	if err := rows.Err(); err != nil {
		log.Println("GetUsersList error:", err)
		return nil, err
	}

	return &chatusers, err
}

//Recupera un Nodo della Chat
func (db *DBnode) GetUrlNode(chatID int64, NodeName string) (*UrlNode, error) {
	log.Println("GetUrlNode:", chatID, NodeName)
	retVal := &UrlNode{}

	stmt, err := db.DB.Prepare("SELECT `UNId`, `ChatID`,`NodeName`,`NodeURL` FROM `urlnodes` where ChatID = ? AND NodeName = ?")
	if err != nil {
		log.Println("GetUrlNode error:", err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(chatID, NodeName).Scan(&retVal.UNId, &retVal.ChatID, &retVal.NodeName, &retVal.NodeURL)
	if err != nil {
		log.Println("GetUrlNode error:", err)
		return nil, err
	} else {
	}
	log.Println("GetUrlNode: ", retVal.UNId, retVal.ChatID, retVal.NodeName, retVal.NodeURL)

	return retVal, err
}

//Aggiorna/crea UrlNode con chiave `ChatID`+`NodeName`
func (db *DBnode) UpdateUrlNode(urlnode *UrlNode) error {
	log.Println("UpdateUrlNode:", urlnode.UNId, urlnode.ChatID, urlnode.NodeName, urlnode.NodeURL)
	u, e := db.GetUrlNode(urlnode.ChatID, urlnode.NodeName)
	if e != nil {
		u = &UrlNode{}
	}
	u.ChatID = urlnode.ChatID
	u.NodeName = urlnode.NodeName
	u.NodeURL = urlnode.NodeURL
	if e != nil { //il record non c'era, lo inseriamo
		stmt, err := db.DB.Prepare("INSERT INTO `urlnodes`(`ChatID`,`NodeName`,`NodeURL`) VALUES (?,?,?)")
		if err != nil {
			log.Println("UpdateUrlNode error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(u.ChatID, u.NodeName, u.NodeURL)
		if err != nil {
			log.Println("UpdateUrlNode error:", err)
		}
	} else { //il record era presente, lo aggiorniamo, la chiave non serve aggiornarla
		stmt, err := db.DB.Prepare("UPDATE urlnodes SET NodeURL = ? WHERE UNId = ?")
		if err != nil {
			log.Println("UpdateUrlNode error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(u.NodeURL, u.UNId)
		if err != nil {
			log.Println("UpdateUrlNode error:", err)
		}
	}

	return nil
}

//Recupera lista nodi per ChatID
func (db *DBnode) GetUrlNodes(chatID int64, limit, offset int) (*[]UrlNode, error) {
	stmt, err := db.DB.Prepare("SELECT `UNId`, `ChatID`,`NodeName`,`NodeURL` FROM `urlnodes` WHERE ChatID = ? LIMIT ? OFFSET ?")
	if err != nil {
		log.Println("GetUrlNodes error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(chatID, limit, offset)
	if err != nil {
		log.Println("GetUrlNodes error:", err)
		return nil, err
	}
	defer rows.Close()
	var urlnodes []UrlNode
	for rows.Next() {
		var unid int64
		var chatid int64
		var nodename string
		var nodeurl string
		err = rows.Scan(&unid, &chatid, &nodename, &nodeurl)
		if err != nil {
			log.Println("GetUrlNodes error:", err)
			return nil, err
		}

		log.Println(unid, chatid, nodename, nodeurl)
		urlnodes = append(urlnodes, UrlNode{UNId: unid, ChatID: chatid, NodeName: nodename, NodeURL: nodeurl})
	}
	if err := rows.Err(); err != nil {
		log.Println("GetUrlNodes error:", err)
		return nil, err
	}

	return &urlnodes, err
}

//Elimina UrlNode con chiave `UNId`
func (db *DBnode) DelNode(unid int64) error {
	log.Println("DelNode:", unid)
	stmt, err := db.DB.Prepare("DELETE FROM `urlnodes` WHERE `UNId` = ?")
	if err != nil {
		log.Println("DelNode error:", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(unid)
	if err != nil {
		log.Println("DelNode error:", err)
	}

	return nil
}

//Recupera una Chiave della Chat
func (db *DBnode) GetChatKey(chatID int64, keyAlias string) (*ChatKey, error) {
	log.Println("GetChatKey:", chatID, keyAlias)
	retVal := &ChatKey{}

	stmt, err := db.DB.Prepare("SELECT `ChatID`,`KeyAlias`,`PubKey` FROM `chatkeys` where ChatID = ? AND KeyAlias = ?")
	if err != nil {
		log.Println("GetChatKey error:", err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(chatID, keyAlias).Scan(&retVal.ChatID, &retVal.KeyAlias, &retVal.PubKey)
	if err != nil {
		log.Println("GetChatKey error:", err)
		return nil, err
	} else {
	}
	log.Println("GetChatKey: ", retVal.ChatID, retVal.KeyAlias, retVal.PubKey)

	return retVal, err
}

//Aggiorna/crea ChatKey con chiave `ChatID`+`KeyAlias`
func (db *DBnode) UpdateChatKey(chatKey *ChatKey) error {
	log.Println("UpdateChatKey:", chatKey.ChatID, chatKey.KeyAlias, chatKey.PubKey)
	ck, e := db.GetChatKey(chatKey.ChatID, chatKey.KeyAlias)
	if e != nil {
		ck = &ChatKey{}
	}
	ck.ChatID = chatKey.ChatID
	ck.KeyAlias = chatKey.KeyAlias
	ck.PubKey = chatKey.PubKey
	if e != nil { //il record non c'era, lo inseriamo
		stmt, err := db.DB.Prepare("INSERT INTO `chatkeys`(`ChatID`,`KeyAlias`,`PubKey`) VALUES (?,?,?)")
		if err != nil {
			log.Println("UpdateChatKey error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(ck.ChatID, ck.KeyAlias, ck.PubKey)
		if err != nil {
			log.Println("UpdateChatKey error:", err)
		}
	} else { //il record era presente, lo aggiorniamo, la chiave non serve aggiornarla
		stmt, err := db.DB.Prepare("UPDATE chatkeys SET PubKey = ? WHERE ChatID = ? AND KeyAlias = ?")
		if err != nil {
			log.Println("UpdateChatKey error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(ck.PubKey, ck.ChatID, ck.KeyAlias)
		if err != nil {
			log.Println("UpdateChatKey error:", err)
		}
	}

	return nil
}

//Recupera lista chiavi per ChatID
func (db *DBnode) GetChatKeys(chatID int64, limit, offset int) (*[]ChatKey, error) {
	stmt, err := db.DB.Prepare("SELECT `ChatID`,`KeyAlias`,`PubKey` FROM `chatkeys` WHERE ChatID = ? LIMIT ? OFFSET ?")
	if err != nil {
		log.Println("GetChatKeys error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(chatID, limit, offset)
	if err != nil {
		log.Println("GetChatKeys error:", err)
		return nil, err
	}
	defer rows.Close()
	var chatkeys []ChatKey
	for rows.Next() {
		var chatid int64
		var keyalias string
		var pubkey string
		err = rows.Scan(&chatid, &keyalias, &pubkey)
		if err != nil {
			log.Println("GetChatKeys error:", err)
			return nil, err
		}

		log.Println(chatid, keyalias, pubkey)
		chatkeys = append(chatkeys, ChatKey{ChatID: chatid, KeyAlias: keyalias, PubKey: pubkey})
	}
	if err := rows.Err(); err != nil {
		log.Println("GetChatKeys error:", err)
		return nil, err
	}

	return &chatkeys, err
}

//Recupera lista chiavi Chat per PubKey
func (db *DBnode) GetChatKeysByPubKey(pubkey string, limit, offset int) (*[]ChatKey, error) {
	stmt, err := db.DB.Prepare("SELECT `ChatID`,`KeyAlias`,`PubKey` FROM `chatkeys` WHERE PubKey = ? LIMIT ? OFFSET ?")
	if err != nil {
		log.Println("GetChatKeysByPubKey error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(pubkey, limit, offset)
	if err != nil {
		log.Println("GetChatKeysByPubKey error:", err)
		return nil, err
	}
	defer rows.Close()
	var chatkeys []ChatKey
	for rows.Next() {
		var chatid int64
		var keyalias string
		var pubkey string
		err = rows.Scan(&chatid, &keyalias, &pubkey)
		if err != nil {
			log.Println("GetChatKeys error:", err)
			return nil, err
		}

		log.Println(chatid, keyalias, pubkey)
		chatkeys = append(chatkeys, ChatKey{ChatID: chatid, KeyAlias: keyalias, PubKey: pubkey})
	}
	if err := rows.Err(); err != nil {
		log.Println("GetChatKeysByPubKey error:", err)
		return nil, err
	}

	return &chatkeys, err
}

//Elimina ChatKey con chiave `ChatID`+`KeyAlias`
func (db *DBnode) DelChatKey(chatID int64, keyAlias string) error {
	log.Println("DelChatKey:", chatID, keyAlias)
	stmt, err := db.DB.Prepare("DELETE FROM `chatkeys` WHERE ChatID = ? AND KeyAlias = ?")
	if err != nil {
		log.Println("DelChatKey error:", err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(chatID, keyAlias)
	if err != nil {
		log.Println("DelChatKey error:", err)
	}

	return nil
}

//Recupera una MiningKey
func (db *DBnode) GetMiningKey(pubkey string) (*MiningKey, error) {
	log.Println("GetMiningKey:", pubkey)
	retVal := &MiningKey{}

	stmt, err := db.DB.Prepare("SELECT `PubKey`,`LastStatus`,`LastPRV`,`IsAutoStake`,`Bls`,`Dsa` FROM `miningkeys` where PubKey = ?")
	if err != nil {
		log.Println("GetMiningKey error:", err)
		return nil, err
	}
	defer stmt.Close()
	err = stmt.QueryRow(pubkey).Scan(&retVal.PubKey, &retVal.LastStatus, &retVal.LastPRV, &retVal.IsAutoStake, &retVal.Bls, &retVal.Dsa)
	if err != nil {
		log.Println("GetMiningKey error:", err)
		return nil, err
	} else {
	}
	log.Println("GetMiningKey: ", retVal.PubKey, retVal.LastStatus)

	return retVal, err
}

//Aggiorna/crea MiningKey con chiave `PubKey`
func (db *DBnode) UpdateMiningKey(miningkey *MiningKey, callback StatusChangeNotifierFunc) error {
	log.Printf("UpdateMiningKey: %+v\n", miningkey)
	mk, e := db.GetMiningKey(miningkey.PubKey) //prendiamo la MiningKey prima di aggiornarla
	var precLastStatus = "missing"
	var precPRV int64 = 0
	if e == nil { //se c'era ci salviamo lo stato precedente e lo aggiorniamo (esclusa la chiave)
		precLastStatus = mk.LastStatus //salviamo il vecchio LastStatus prima di aggiornare
		precPRV = mk.LastPRV           //salviamo il vecchio PRV prima di aggiornare
		stmt, err := db.DB.Prepare("UPDATE miningkeys SET LastStatus = ?, LastPRV = ?, IsAutoStake = ?, Bls = ?, Dsa = ? WHERE PubKey = ?")
		if err != nil {
			log.Println("UpdateMiningKey error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(miningkey.LastStatus, miningkey.LastPRV, miningkey.IsAutoStake, miningkey.Bls, miningkey.Dsa, miningkey.PubKey)
		if err != nil {
			log.Println("UpdateMiningKey error:", err)
		}
	} else { //il record non c'era, lo inseriamo
		precLastStatus = "missing" //non abbiano un LastStatus precedente
		stmt, err := db.DB.Prepare("INSERT INTO `miningkeys`(`PubKey`,`LastStatus`,`LastPRV`,`IsAutoStake`,`Bls`,`Dsa`) VALUES (?,?,?,?,?,?)")
		if err != nil {
			log.Println("UpdateMiningKey error:", err)
			return err
		}
		defer stmt.Close()

		_, err = stmt.Exec(miningkey.PubKey, miningkey.LastStatus, miningkey.LastPRV, miningkey.IsAutoStake, miningkey.Bls, miningkey.Dsa)
		if err != nil {
			log.Println("UpdateMiningKey error:", err)
		}
	}

	log.Printf("UpdateMiningKey STATUS: (%s)=(%s) (%d)=(%d)\n", precLastStatus, miningkey.LastStatus, precPRV, miningkey.LastPRV)
	if precLastStatus != miningkey.LastStatus { //status changed, must notify
		log.Printf("UpdateMiningKey found status change for key %s: from \"%s\" to\" %s\".", miningkey.PubKey, precLastStatus, miningkey.LastStatus)
		err := callback(miningkey, precLastStatus, precPRV)
		if err != nil {
			log.Println("UpdateMiningKey Err in callback: ", err)
		}

	}

	return nil
}

//Recupera lista chiavi mining
func (db *DBnode) GetMiningKeys(limit, offset int) (*[]MiningKey, error) {
	stmt, err := db.DB.Prepare("SELECT `PubKey`,`LastStatus`,`LastPRV`,`IsAutoStake`,`Bls`,`Dsa` FROM `miningkeys` LIMIT ? OFFSET ?")
	if err != nil {
		log.Println("GetMiningKeys error:", err)
		return nil, err
	}
	defer stmt.Close()

	rows := &sql.Rows{}
	rows, err = stmt.Query(limit, offset)
	if err != nil {
		log.Println("GetMiningKeys error:", err)
		return nil, err
	}
	defer rows.Close()
	var miningkeys []MiningKey
	for rows.Next() {
		var pubkey string
		var laststatus string
		var lastprv int64
		var isautostake bool
		var bls string
		var dsa string
		err = rows.Scan(&pubkey, &laststatus, &lastprv, &isautostake, &bls, &dsa)
		if err != nil {
			log.Println("GetMiningKeys error:", err)
			return nil, err
		}

		log.Println(pubkey, laststatus)
		miningkeys = append(miningkeys, MiningKey{PubKey: pubkey, LastStatus: laststatus, LastPRV: lastprv, IsAutoStake: isautostake, Bls: bls, Dsa: dsa})
	}
	if err := rows.Err(); err != nil {
		log.Println("GetMiningKeys error:", err)
		return nil, err
	}

	return &miningkeys, err
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
	var create_statements = [...]string{
		`CREATE TABLE IF NOT EXISTS "chatdata" ( "ChatID" integer NOT NULL, "Name" text, "NameAsked" INTEGER DEFAULT 1, PRIMARY KEY("ChatID") )`,
		`CREATE TABLE IF NOT EXISTS "urlnodes" ( "UNId" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT UNIQUE, "ChatID" INTEGER, "NodeName" TEXT, "NodeURL" TEXT )`,
		`CREATE TABLE IF NOT EXISTS "chatkeys" ( "ChatID" INTEGER, "KeyAlias" TEXT, "PubKey" TEXT, PRIMARY KEY("ChatID","KeyAlias") )`,
		`CREATE TABLE IF NOT EXISTS "miningkeys" ( "PubKey" TEXT NOT NULL UNIQUE, "LastStatus" TEXT, "LastPRV" INTEGER, "IsAutoStake" INTEGER, "Bls" TEXT, "Dsa" TEXT, PRIMARY KEY("PubKey") )`,
	}
	var err error = nil
	for _, statement := range create_statements {
		log.Println(statement)
		_, err = db.DB.Exec(statement)
		if err != nil {
			log.Println("CreateTablesIfNotExists error:", err)
			return err
		}
	}
	return err
}
