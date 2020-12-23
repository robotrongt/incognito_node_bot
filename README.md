# Incognito Node Bot

Questo telegram bot serve a monitorare nodi del blockchain incognito. Vedere il blockchain incognito su [incognito.org](https://incognito.org/)


# Preparazione

## Creare chiave e certificato autofirmato con openssl per il demone

Esempio:

```bash
cd src/cmd/incognito_node_bot
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3560 -subj '/C=IT/O=Organizzazione/CN=your.host.name' -addext 'subjectAltName=IP:PUB.LIC.IP.ADDR,DNS:your.host.name' -nodes
```
## Impostare variabili di ambiente
Impostare variabile di ambiente con TOKEN di autorizzazione ricevuto dal `@BotFather` ed impostare variabile DBFILE con il nome del file del db da creare/usare
Definire anche un token (TGTOKEN) arbitrario per riconoscere le chiamate che ci farà Telegram ed assegnargli un valore random da registrare poi presso Telegram.
Servono anche DEFAUL_NODE_URL da usare se il nodo fornito dall'utente è errato 
e DEFAULT_FULLNODE_URL da usare per richieste a fullNode

Esempio:

```bash
export TOKEN=1234567890:ABcdEFghIL_M0noPQrsTUvwXYaBcDeFgHiL
export TGTOKEN=eeRiev7Efu4Esah3cafooPoh4dou7thi
export DBFILE=/srv/incbot.db
export DEFAULT_NODE_URL=http://127.0.0.1:9334
export DEFAULT_FULLNODE_URL=https://mainnet.incognito.org/fullnode
```

## Upload ed attivazione del `Webhook` verso il nostro bot presso telegram 

Esempio:

```bash
curl -F "url=https://your.host.name:8443/telegram${TGTOKEN}/" -F "certificate=@cert.pem" https://api.telegram.org/bot${TOKEN}/setWebhook
```

## Compilare e lanciare il bot

```bash
cd src/cmd/incognito_node_bot
go build
./incognito_node_bot
```

## Comando da schedulare per controllo stato mining chiavi ed eventuale notifica

```bash
cd src/cmd/incognito_check_miningkeys
go build
while true; do ./incognito_check_miningkeys; sleep 1m; done
```


## Qualche link di documentazione

https://www.sohamkamani.com/golang/telegram-bot/

https://core.telegram.org/bots/webhooks#how-do-i-set-a-webhook-for-either-type

https://core.telegram.org/bots/api#update

https://core.telegram.org/bots/api#sendmessage


