package mailcrawler

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"example.com/config"
	"example.com/dbhandler"

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

	defer c.Logout()

	if err := c.Login(emailConf.Auth.User, emailConf.Auth.Password); err != nil {
		log.Fatal(err)
	}
	done := make(chan error, 1)

	mbox, err := c.Select(emailConf.Paths.Source, false)
	if err != nil {
		log.Fatal(err)
	}
	if mbox.Messages == 0 {
		return
	}
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
		validMails := new(imap.SeqSet)
		invalidMails := new(imap.SeqSet)
		hasPdfAttachment := false
		var emailLog dbhandler.Email
		var attachmentsArray []string
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
		emailLog.Uuid = uuid.NewString()
		emailLog.HasPdfAttachment = 0
		emailLog.Completed = 1
		emailLog.Sender = mr.Header.Get("From")
		emailLog.Date = mr.Header.Get("Date")
		emailLog.Subject = mr.Header.Get("Subject")
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				log.Println("Attachment found")
				var attachmentLog dbhandler.Attachment
				filename, _ := h.Filename()
				if pdf.MatchString(filename) {
					hasPdfAttachment = true
					log.Println(filename)
					genUuid := "a" + uuid.NewString()
					newFilename := genUuid + ".pdf"
					file, err := os.Create(filepath.Join(emailConf.DownloadFolder, newFilename))
					if err != nil {
						log.Fatal(err)
					}
					size, err := io.Copy(file, p.Body)
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("Saved %v bytes into %v\n", size, filename)
					attachmentLog.Uuid = genUuid
					attachmentLog.FromEmail = emailLog.Uuid
					attachmentLog.DateOfDownload = strconv.FormatInt(time.Now().Unix(), 10)
					attachmentLog.OgFilename = filename
					attachmentLog.NewFilename = newFilename
					attachmentLog.FileProcessed = 0
					attachmentsArray = append(attachmentsArray, genUuid)
					emailLog.HasPdfAttachment = 1
					emailLog.Completed = 0
					dbhandler.LogAttachment(attachmentLog)
					file.Close()
					file.Sync()
				}
			}
		}
		for i, item := range attachmentsArray {
			if i != 0 {
				emailLog.AttachmentsUuidArrray += ","
			}
			emailLog.AttachmentsUuidArrray += item
		}
		dbhandler.LogEmail(emailLog)
		if hasPdfAttachment {
			validMails.AddNum(msg.Uid)
			//c.UidMove(validMails, emailConf.Paths.ProcessedMails)
			// log.Println("Moving email from inbox")
			//c.UidMove(validMails, "MARTINOK")
		} else {
			invalidMails.AddNum(msg.Uid)
			//c.UidMove(invalidMails, emailConf.Paths.InvalidMails)
			//c.UidMove(invalidMails, "MARTINNOK")
			// log.Println("Emails moved")
		}

	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

}
