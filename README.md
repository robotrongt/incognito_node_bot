# Preparazione

## Creare chiave e certificato autofirmato con openssl per il demone

Esempio:

```bash
cd src/cmd/incognito_node_bot
openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3560 -subj '/C=IT/O=Organizzazione/CN=your.host.name' -addext 'subjectAltName=IP:PUB.LIC.IP.ADDR,DNS:your.host.name' -nodes
```
## Impostare variabile di ambiente con TOKEN di autorizzazione ricevuto dal `@BotFather`

Esempio:

```bash
TOKEN=1234567890:ABcdEFghIL_M0noPQrsTUvwXYaBcDeFgHiL
```

## Upload ed attivazione del `Webhook` verso il nostro bot presso telegram 

Esempio:

```bash
curl -F "url=https://your.host.name:8443/" -F "certificate=@cert.pem" https://api.telegram.org/bot${TOKEN}/setWebhook
```

## Qualche link di documentazione

https://www.sohamkamani.com/golang/telegram-bot/
https://core.telegram.org/bots/webhooks#how-do-i-set-a-webhook-for-either-type
https://core.telegram.org/bots/api#update
https://core.telegram.org/bots/api#sendmessage

