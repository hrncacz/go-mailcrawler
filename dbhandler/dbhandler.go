package dbhandler

import (
	"database/sql"
	"fmt"
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
	_, err = database.Exec(fmt.Sprintf("UPDATE emails SET completed = 1 WHERE uuid=\"%s\"", emailUuid))
	checkErr(err)
}

func AttachmentIsProcessed(attachmentUuid string) {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	_, err = database.Exec(fmt.Sprintf("UPDATE attachments SET fileProcessed = 1 WHERE uuid=\"%s\"", attachmentUuid))
}

func EmailGetStatus(emailUuid string) int8 {
	database, _ := sql.Open("sqlite3", dbPath())
	defer database.Close()
	var status int8
	rows, _ := database.Query(fmt.Sprintf("SELECT fileProcessed FROM attachments WHERE uuid=\"%s\"", emailUuid))
	if rows.Next() {
		rows.Scan(&status)
	}
	return status
}

func AttachmentGetStatus(attachmentUuid string) int8 {
	database, _ := sql.Open("sqlite3", dbPath())
	defer database.Close()
	var status int8
	rows, err := database.Query(fmt.Sprintf("SELECT fileProcessed FROM attachments WHERE uuid=\"%v\"", attachmentUuid))
	checkErr(err)
	for rows.Next() {
		rows.Scan(&status)
	}
	return status
}

func GetEmail(attachmentUuid string) Email {
	database, err := sql.Open("sqlite3", dbPath())
	checkErr(err)
	defer database.Close()
	var emailUuid Email
	rows, _ := database.Query(fmt.Sprintf("SELECT * FROM attachments WHERE uuid=\"%s\"", attachmentUuid))
	if rows.Next() {
		rows.Scan(&emailUuid)
	}
	return emailUuid
}

func AllAttachmentsComplete(attachmentUuid string) {
	email := GetEmail(attachmentUuid)
	isComplete := true
	attachmentsArray := strings.Split(email.AttachmentsUuidArrray, ",")
	for _, item := range attachmentsArray {
		attachmentStatus := AttachmentGetStatus(item)
		if attachmentStatus == 0 {
			isComplete = false
		}
	}
	if isComplete == true {
		EmailIsProcessed(email.Uuid)
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
