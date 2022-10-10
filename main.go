package main

import (
	"GOLANG/config"
	//"GOLANG/ftp"
	"GOLANG/mailcrawler"
	"log"
	//"time"
)

func testChannel(ch chan int) {
	log.Println("Začíná funkce")
	ch <- 355
	log.Println("Končí funkce")
}

func main() {
	c, err := config.GetConfig()
	// ftp.UploadFiles(c.Ftp.Auth)
	if err != nil {
		log.Println("CHYBAAAAAAAAAAAAAAAAAAAAAAAA")
	}

	// log.Println("Spouštím channel")
	// values := make(chan int, 3)
	// defer close(values)
	// go testChannel(values)
	// go testChannel(values)
	// go testChannel(values)

	// value := <-values
	// log.Println(value)
	// time.Sleep(5 * time.Second)
	mailcrawler.MailCrawler(c.Email)
}
