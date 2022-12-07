package main

import (
	"log"

	"example.com/config"
	"example.com/dbhandler"

	//"example.com/ftpup"
	"example.com/mailcrawler"
)

func main() {
	c, err := config.GetConfig()
	log.Println("Config loaded")
	log.Println("Init database")
	dbhandler.InitTables()
	log.Println("Checking emails")
	mailcrawler.MailCrawler(c.Email)
	log.Println("Uploading and Downloading files")
	//ftpup.FtpUploadDownload(c.Ftp)
	if err != nil {
		log.Println("CHYBAAAAAAAAAAAAAAAAAAAAAAAA")
	}
}
