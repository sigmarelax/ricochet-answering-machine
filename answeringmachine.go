package main

import (
	"bufio"
	"bytes"
	"crypto/subtle"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/divan/num2words"
	"github.com/s-rah/go-ricochet/application"
	"github.com/s-rah/go-ricochet/utils"
)

// these _need_ to be edited
var adminID = "" // example "anomalyolpdslogr"
var adminPass = "My message is my passport. Verify me"

const adminName = "Nate"

func newprivatekey() {
	if _, err := ioutil.ReadFile("private_key"); err != nil {
		log.Printf("Error accessing the private key file: %v \nWill attempt to generate new private key...", err)

		generatedkey, err := utils.GeneratePrivateKey()
		if err != nil {
			log.Fatalf("Error generating a new private key: %v", err)
		}

		keystring := utils.PrivateKeyToString(generatedkey)

		err = ioutil.WriteFile("./private_key", []byte(keystring), 0644)
		if err != nil {
			log.Fatalf("Error writing a new private key: %v", err)
		}

		log.Printf("Successfully created new private key.")
	}
}

func readState() (storedmessages []string, err error) {
	var (
		file   *os.File
		piece  []byte
		prefix bool
	)
	if file, err = os.Open("bot_state"); err != nil {
		return
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := bytes.NewBuffer(make([]byte, 0))
	for {
		if piece, prefix, err = reader.ReadLine(); err != nil {
			break
		}
		buffer.Write(piece)
		if !prefix {
			storedmessages = append(storedmessages, buffer.String())
			buffer.Reset()
		}
	}
	if err == io.EOF {
		err = nil
	}
	return
}

func writestate(lines []string) (err error) {
	var file *os.File

	file, err = os.Create("bot_state")
	if err != nil {
		return
	}
	defer file.Close()

	for _, msgstring := range lines {
		_, err := file.WriteString(strings.TrimSpace(strings.Replace(msgstring, "\n", " ", -1)) + "\n")
		if err != nil {
			break
		}
	}
	return
}

func main() {
	voicemails := make([]string, 0)
	var deletedvoicemail string

	voicemails, err := readState()
	if err != nil {
		log.Printf("No previous Answering Machine state (%v)", err)
	}

	answeringmachine := new(application.RicochetApplication)

	newprivatekey()

	pk, err := utils.LoadPrivateKeyFromFile("./private_key")
	if err != nil {
		log.Fatalf("Error loading private key file: %v", err)
	}

	l, err := application.SetupOnion("127.0.0.1:9051", "tcp4", "", pk, 9878)
	if err != nil {
		log.Fatalf("Error setting up onion service: %v", err)
	}

	answeringmachine.Init(pk, new(application.AcceptAllContactManager))

	answeringmachine.OnChatMessage(func(rai *application.RicochetApplicationInstance, id uint32, timestamp time.Time, message string) {

		if rai.RemoteHostname != adminID {
			if subtle.ConstantTimeCompare([]byte(message), []byte(adminPass)) == 0 {
				//  format and insert message into slice
				messageandmeta := fmt.Sprint(rai.RemoteHostname, " [", time.Now().Format("15:04:05 Jan _2 2006"), "]: ", message)
				rai.SendChatMessage("Sorry," + adminName + " is not available. Your message has been stored in " + adminName + "'s answering machine.")
				voicemails = append(voicemails, messageandmeta)
				err = writestate(voicemails)
				if err != nil {
					log.Printf("Error saving Answering Machine state: %v", err)
				}
			} else {
				// establish new admin
				adminID = rai.RemoteHostname
				rai.SendChatMessage("You are now the Admin! \"/h\" for list of commands")
			}

		} else if len(message) < 2 {
			rai.SendChatMessage("Sorry, " + adminName + ".\nThis command is too short. Please type /h for a list of valid commands.")

		} else if rai.RemoteHostname == adminID {
			// interpret remote commands
			switch message[0:2] {
			case "/m":
				// playback the oldest stored message
				if len(voicemails) > 0 {
					rai.SendChatMessage(voicemails[0])
					deletedvoicemail = voicemails[0]
					voicemails = voicemails[1:]
					err = writestate(voicemails)
					if err != nil {
						log.Printf("Error saving Answering Machine state: %v", err)
						rai.SendChatMessage("Error deleting message.")
					}
				} else {
					rai.SendChatMessage("There are currently *" + strings.ToUpper(num2words.Convert(len(voicemails))) + "* messages.")
				}

			case "/k":
				if len(deletedvoicemail) == 0 {
					rai.SendChatMessage("No deleted message found.")
					break
				}

				voicemails = append(voicemails, deletedvoicemail)

				err = writestate(voicemails)
				if err != nil {
					log.Printf("Error saving Answering Machine state: %v", err)
					rai.SendChatMessage("Error saving last message.")
				} else {
					rai.SendChatMessage("Last played message has been saved.")
				}

			case "/n":
				rai.SendChatMessage("There are currently *" + strings.ToUpper(num2words.Convert(len(voicemails))) + "* messages.")

			case "/q":
				rai.SendChatMessage("Goodbye.")
				os.Exit(0)

			case "/p":
				rai.SendChatMessage("New password: " + message[3:])
				adminPass = message[3:]

			case "/h":
				// help
				rai.SendChatMessage("/m will play the oldest message.\n/k will undelete and keep the last played message.\n/n will list how many messages exist.\n/p will set a new password.\n/q will turn off the answering machine.")

			default:
				rai.SendChatMessage("Sorry, " + adminName + ".\nThis command is invalid. Please type /h for a list of valid commands.")
			}
		}
	})

	log.Printf("Ricochet Answering Machine is now available at ricochet:%s", l.Addr().String()[:16])

	answeringmachine.Run(l)
}
