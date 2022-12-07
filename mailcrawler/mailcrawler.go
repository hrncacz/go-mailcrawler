package mailcrawler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"example.com/config"
	"example.com/dbhandler"
	sapcomm "example.com/sap"
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
	"github.com/google/uuid"
	"golang.org/x/net/html/charset"
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
	var sapEmailLog []sapcomm.SapEmailItem
	var sapAttachmentLog []sapcomm.SapAttachmentItem
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
		emailLog.Sender = getSender(decodeString(mr.Header.Get("From")))

		mailServerTimeTemplate := "Mon, 2 Jan 2006 15:04:05 -0700"
		mailserverTime, _ := time.Parse(mailServerTimeTemplate, decodeString(mr.Header.Get("Date")))
		sapTimeTemplate := "20060102"
		emailLog.Date = mailserverTime.Format(sapTimeTemplate)
		emailLog.Subject = decodeString(mr.Header.Get("Subject"))

		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}
			switch h := p.Header.(type) {
			case *mail.AttachmentHeader:
				var attachmentLog dbhandler.Attachment
				filename, _ := h.Filename()
				if pdf.MatchString(filename) {
					hasPdfAttachment = true
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
					attachmentLog.DateOfDownload = time.Now().Format("20060102")
					attachmentLog.OgFilename = decodeString(filename)
					attachmentLog.NewFilename = newFilename
					attachmentLog.FileProcessed = 0
					attachmentsArray = append(attachmentsArray, genUuid)
					emailLog.HasPdfAttachment = 1
					emailLog.Completed = 0
					dbhandler.LogAttachment(attachmentLog)
					sapAttachmentLog = append(sapAttachmentLog, prepareAttachmentForSapLog(attachmentLog))
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
		sapEmailLog = append(sapEmailLog, prepareEmailForSapLog(emailLog))
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
	sapcomm.Sapcomm(sapEmailLog, sapAttachmentLog)

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	log.Println("Done!")

}

func prepareEmailForSapLog(emailLog dbhandler.Email) sapcomm.SapEmailItem {
	var sapEmailItem sapcomm.SapEmailItem
	sapEmailItem.EMLUUID = emailLog.Uuid
	sapEmailItem.CLIENT = 100
	sapEmailItem.RECVDATE = emailLog.Date
	sapEmailItem.SUBJECT = emailLog.Subject
	sapEmailItem.SENDER = emailLog.Sender
	sapEmailItem.HASPDF = emailLog.HasPdfAttachment
	sapEmailItem.ATTACHMENTS = emailLog.AttachmentsUuidArrray
	sapEmailItem.COMPLETED = emailLog.Completed

	return sapEmailItem
}

func prepareAttachmentForSapLog(attachmentLog dbhandler.Attachment) sapcomm.SapAttachmentItem {
	var sapAttachmentItem sapcomm.SapAttachmentItem
	sapAttachmentItem.UUID = attachmentLog.Uuid
	sapAttachmentItem.FROMEMAIL = attachmentLog.FromEmail
	sapAttachmentItem.DOWNDATE = attachmentLog.DateOfDownload
	sapAttachmentItem.OGFILENAME = attachmentLog.OgFilename
	sapAttachmentItem.NEWFILENAME = attachmentLog.NewFilename
	sapAttachmentItem.CLIENT = 100
	sapAttachmentItem.FILEPROCESSED = attachmentLog.FileProcessed

	return sapAttachmentItem
}

func getSender(fromHeader string) string {
	re := regexp.MustCompile(`<.*@.*\..*>`)
	result := re.FindString(fromHeader)
	result = strings.ReplaceAll(result, "<", "")
	result = strings.ReplaceAll(result, ">", "")
	return result
}

func decodeString(textToDecode string) string {

	dec := new(mime.WordDecoder)
	dec.CharsetReader = func(chs string, input io.Reader) (io.Reader, error) {
		log.Println(chs)
		switch chs {
		case "iso-8859-2":
			r, err := charset.NewReaderLabel(chs, input)
			if err != nil {
				return nil, err
			}
			content, err := ioutil.ReadAll(r)
			if err != nil {
				return nil, err
			}
			return bytes.NewReader(bytes.ToUpper(content)), nil
		default:
			return nil, fmt.Errorf("unhandled charset %q", chs)
		}
	}
	text, err := dec.DecodeHeader(textToDecode)
	if err != nil {
		return textToDecode
	}
	return text
}
