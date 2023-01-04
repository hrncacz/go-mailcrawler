package main

import (
	"log"

	"example.com/config"
	"example.com/db2sap"
	"example.com/dbhandler"

	//"example.com/ftpup"
	"example.com/mailcrawler"
)

func main() {
	c, err := config.GetConfig()
	log.Println("Config loaded")
	log.Println("Init database")
	dbhandler.InitTables()
	db2sap.CopyRecords()
	if err != nil {
		log.Println("CHYBAAAAAAAAAAAAAAAAAAAAAAAA")
		log.Println("Checking emails")
		mailcrawler.MailCrawler(c.Email)
		log.Println("Uploading and Downloading files")
		//ftpup.FtpUploadDownload(c.Ftp)
	}
}
