# Ricochet Answering Machine

This project runs a [Ricochet IM](https://ricochet.im) bot that will receive/store/retrieve missed messages, much like a telephone answering machine. This is built using [GoRicochet](https://github.com/s-rah/go-ricochet) (an experimental implementation of the Ricochet Protocol in Go).

## Features

* Generates a new Ricochet identity if a private_key is not provided
* Stores incoming messages
* Permits admin access to the machine from different Ricochet identities with the use of a passphrase

Future features:
* Send/queue outgoing messages for the next time that contact is available

## Warnings

I offer no guarantees that this is tested or will maintain your anonymity.

## How to setup

1. Install [Go](https://golang.org/doc/install) and [Tor](https://torproject.org/download)

2. Configure Tor (to run a hidden service on port 9878 and allow cookie control) by editing torrc

    For example, in ubuntu run these commands with root:
```
echo -e "ControlPort 9051\nCookieAuthentication 1" >> /etc/tor/torrc
echo -e "HiddenServiceDir /var/lib/tor/hidden_service/\nHiddenServicePort 9878 127.0.0.1:9878" >> /etc/tor/torrc
service tor restart
chmod 644 /var/run/tor/control.authcookie
```

3. go get github.com/sigmarelax/ricochet-answering-machine

4. Navigate to the source and edit settings for passphrase and admin at the top of main.go

5. go run main.go
