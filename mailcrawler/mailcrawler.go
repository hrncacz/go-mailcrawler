package mailcrawler

import (
	"GOLANG/config"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/google/uuid"
)

func MailCrawler(emailConf config.Email) {
	log.Println("Connecting to server...")
	c, err := client.DialTLS("exchange.vollmann-group.com:993", nil)

	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	validMails := new(imap.SeqSet)
	invalidMails := new(imap.SeqSet)

	defer c.Logout()

	if err := c.Login(emailConf.Auth.User, emailConf.Auth.Password); err != nil {
		log.Fatal(err)
	}
	done := make(chan error, 1)

	mbox, err := c.Select("MARTIN", false)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(mbox.Messages)
	seqset := new(imap.SeqSet)
	seqset.AddRange(uint32(1), mbox.Messages)

	messages := make(chan *imap.Message, mbox.Messages)
	done = make(chan error, 2)
	pdf := regexp.MustCompile("(?i).pdf")
	var section imap.BodySectionName
	go func() {
		done <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem(), imap.FetchUid}, messages)
	}()
	for msg := range messages {
		hasPdfAttachment := false
		if msg == nil {
			log.Fatal("Server didn't return message!")
		}

		r := msg.GetBody(&section)
		if r == nil {
			log.Fatal("Server didn't returned message body")
		}
		mr, err := mail.CreateReader(r)
		if err != nil {
			log.Fatal(err)
		}

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				filename, _ := h.Filename()
				if pdf.MatchString(filename) {
					hasPdfAttachment = true
					log.Println(filename)
					file, err := os.Create(emailConf.DownloadFolder + uuid.NewString())
					if err != nil {
						log.Fatal(err)
					}
					size, err := io.Copy(file, p.Body)
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("Saved %v bytes into %v\n", size, filename)
				}
			}
		}

		if hasPdfAttachment {
			validMails.AddNum(msg.Uid)
		} else {
			invalidMails.AddNum(msg.Uid)
		}
	}
	// log.Println("Moving email from inbox")
	// c.UidMove(validMails, "MARTINOK")
	// c.UidMove(invalidMails, "MARTINNOK")
	// log.Println("Emails moved")

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

}
