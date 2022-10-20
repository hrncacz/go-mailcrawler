package main

import (
	"log"

	"example.com/config"
	"example.com/dbhandler"
	"example.com/ftpup"
	"example.com/mailcrawler"
)

func main() {
	c, err := config.GetConfig()
	log.Printf("Config loaded: %v", c)
	log.Println("Init database")
	dbhandler.InitTables()
	mailcrawler.MailCrawler(c.Email)
	ftpup.FtpUploadDownload(c.Ftp)
	if err != nil {
		log.Println("CHYBAAAAAAAAAAAAAAAAAAAAAAAA")
	}
}
