package dbhandler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func dbPath() string {
	var path string
	if runtime.GOOS == "windows" {
		ex, err := os.Executable()
		if err != nil {
			panic(err)
		}
		exPath := filepath.Dir(ex)

		path = filepath.Join(exPath, "db", "main.db")
	} else {
		path = "db/main.db"
	}
	return path
}

func InitTables() {
	path := dbPath()
	database, _ := sql.Open("sqlite3", path)
	defer database.Close()
	database.Exec("CREATE TABLE IF NOT EXISTS emails(uuid TEXT, date TEXT, subject TEXT, sender TEXT, hasPdfAttachment INTEGER, attachments, completed)")
	database.Exec("CREATE TABLE IF NOT EXISTS attachments(uuid, fromEmail, dateOfDownload, ogFilename, newFilename, fileProcessed)")
}

func LogEmail(email Email) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	_, err = database.Exec("INSERT INTO emails(uuid, date, subject, sender, hasPdfAttachment, attachments, completed) VALUES (?, ?, ?, ?, ?, ?, ?)", email.Uuid, email.Date, email.Subject, email.Sender, email.HasPdfAttachment, email.AttachmentsUuidArrray, email.Completed)
	checkErr(err)
}

func LogAttachment(attachment Attachment) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	_, err = database.Exec("INSERT INTO attachments(uuid, fromEmail, dateOfDownload, ogFilename, newFilename, fileProcessed) VALUES (?, ?, ?, ?, ?, ?)", attachment.Uuid, attachment.FromEmail, attachment.DateOfDownload, attachment.OgFilename, attachment.NewFilename, attachment.FileProcessed)
	checkErr(err)
}

func EmailIsProcessed(emailUuid string) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	statement, err := database.Prepare(`UPDATE emails SET completed=1 WHERE uuid=$1`)
	checkErr(err)
	defer statement.Close()
	statement.Exec(emailUuid)
}

func AttachmentIsProcessed(attachmentUuid string) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	statement, err := database.Prepare(`UPDATE attachments SET fileProcessed=1 WHERE uuid=$1`)
	checkErr(err)
	defer statement.Close()
	statement.Exec(attachmentUuid)
}

func EmailGetStatus(emailUuid string) int8 {
	database, _ := sql.Open("sqlite3", dbPath())
	defer database.Close()
	var status int8
	row := database.QueryRow(`SELECT fileProcessed FROM attachments WHERE uuid=$1`, emailUuid)
	switch err := row.Scan(&status); err {
	case sql.ErrNoRows:
		status = 0
		return status
	case nil:
		return status
	default:
		return 0
	}
}

func AttachmentGetStatus(attachmentUuid string) int8 {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	var status int8
	row := database.QueryRow(`SELECT fileProcessed FROM attachments WHERE uuid=$1`, attachmentUuid)
	switch err := row.Scan(&status); err {
	case sql.ErrNoRows:
		status = 0
		return status
	case nil:
		return status
	default:
		return 0
	}
}

func GetEmail(attachmentUuid string) (emailUuid string, attachmentsUuidArrray string, err error) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	row := database.QueryRow(`SELECT fromEmail FROM attachments WHERE uuid=$1`, attachmentUuid)
	err = row.Scan(&emailUuid)
	checkErr(err)
	row = database.QueryRow(`SELECT attachments FROM emails WHERE uuid=$1`, emailUuid)
	switch err := row.Scan(&attachmentsUuidArrray); err {
	case sql.ErrNoRows:
		return "", "", err
	default:
		log.Println(attachmentsUuidArrray)
		return emailUuid, attachmentsUuidArrray, nil
	}
}

func GetAllEmails() (emails []Email, err error) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	rows, err := database.Query(`SELECT * FROM emails`)
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var tmpEmailObj Email
		err = rows.Scan(&tmpEmailObj.Uuid, &tmpEmailObj.Date, &tmpEmailObj.Subject, &tmpEmailObj.Sender, &tmpEmailObj.HasPdfAttachment, &tmpEmailObj.AttachmentsUuidArrray, &tmpEmailObj.Completed)
		checkErr(err)
		emails = append(emails, tmpEmailObj)
	}
	return emails, nil
}

func GetAllAttachments() (attachments []Attachment, err error) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	rows, err := database.Query(`SELECT * FROM attachments`)
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var tmpAttachmentObj Attachment
		err = rows.Scan(&tmpAttachmentObj.Uuid, &tmpAttachmentObj.FromEmail, &tmpAttachmentObj.DateOfDownload, &tmpAttachmentObj.OgFilename, &tmpAttachmentObj.NewFilename, &tmpAttachmentObj.FileProcessed)
		checkErr(err)
		attachments = append(attachments, tmpAttachmentObj)
	}
	return attachments, nil
}

func AllAttachmentsComplete(attachmentUuid string) {
	emailUuid, attachmentsUuidArrray, err := GetEmail(attachmentUuid)
	if err == nil {
		isComplete := true
		attachmentsArray := strings.Split(attachmentsUuidArrray, ",")
		for _, item := range attachmentsArray {
			attachmentStatus := AttachmentGetStatus(item)
			if attachmentStatus == 0 {
				isComplete = false
			}
		}
		if isComplete == true {
			EmailIsProcessed(emailUuid)
		}
	}
}

func checkErr(err error, args ...string) {
	if err != nil {
		fmt.Println("Error")
		fmt.Printf("%v: %v", err, args)
	}
}

type Email struct {
	Uuid                  string
	Date                  string
	Subject               string
	Sender                string
	HasPdfAttachment      int8
	AttachmentsUuidArrray string
	Completed             int8
}

type Attachment struct {
	Uuid           string
	FromEmail      string
	DateOfDownload string
	OgFilename     string
	NewFilename    string
	FileProcessed  int8
}
