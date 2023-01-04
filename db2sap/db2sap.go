package db2sap

import (
	"fmt"
	"log"

	// "strconv"
	// "time"

	"example.com/dbhandler"
	// "example.com/mailcrawler"
	//sapcomm "example.com/sap"
)

func CopyRecords() {
	dbEmails, err := dbhandler.GetAllEmails()
	//dbAttachments, err := dbhandler.GetAllAttachments()
	//var decodedEmails []dbhandler.Email
	//var decodedAttachments []dbhandler.Attachment
	//var sapReadyEmails []sapcomm.SapEmailItem
	//var sapReadyAttachments []sapcomm.SapAttachmentItem
	checkErr(err)
	for _, item := range dbEmails {
		log.Println(item.Date)
		// var tmpEmailObj dbhandler.Email
		// tmpEmailObj.Uuid = item.Uuid
		// mailServerTimeTemplate := "Mon, 2 Jan 2006 15:04:05 -0700"
		// mailserverTime, err := time.Parse(mailServerTimeTemplate, mailcrawler.DecodeString(item.Date))
		// sapTimeTemplate := "20060102"
		// tmpEmailObj.Date = mailserverTime.Format(sapTimeTemplate)
		// if err != nil {
		// 	tmpEmailObj.Date = mailcrawler.DecodeString(item.Date)
		// }
		// tmpEmailObj.Subject = mailcrawler.DecodeString(item.Subject)
		// tmpEmailObj.Sender = mailcrawler.GetSender(mailcrawler.DecodeString(item.Sender))
		// tmpEmailObj.HasPdfAttachment = item.HasPdfAttachment
		// tmpEmailObj.AttachmentsUuidArrray = item.AttachmentsUuidArrray
		// tmpEmailObj.Completed = item.Completed
		// decodedEmails = append(decodedEmails, tmpEmailObj)
	}

	// for _, item := range dbAttachments {
	// 	log.Println(item.DateOfDownload)
	// 	var tmpAttachmentObj dbhandler.Attachment
	// 	tmpAttachmentObj.Uuid = item.Uuid
	// 	tmpAttachmentObj.FromEmail = mailcrawler.DecodeString(item.FromEmail)
	// 	if len(item.DateOfDownload) > 8 {
	// 		intUnix, err := strconv.ParseInt(item.DateOfDownload, 10, 64)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		formattedTime := time.Unix(intUnix, 0).Format("20060102")
	// 		tmpAttachmentObj.DateOfDownload = formattedTime
	// 	} else {
	// 		tmpAttachmentObj.DateOfDownload = item.DateOfDownload
	// 	}
	// 	tmpAttachmentObj.OgFilename = mailcrawler.DecodeString(item.OgFilename)
	// 	tmpAttachmentObj.NewFilename = item.NewFilename
	// 	tmpAttachmentObj.FileProcessed = item.FileProcessed
	// 	decodedAttachments = append(decodedAttachments, tmpAttachmentObj)
	// }

	// for _, item := range decodedEmails {
	// 	tmpSapEmailObj := mailcrawler.PrepareEmailForSapLog(item)
	// 	sapReadyEmails = append(sapReadyEmails, tmpSapEmailObj)
	// }

	// for _, item := range decodedAttachments {
	// 	tmpSapAttachmentObj := mailcrawler.PrepareAttachmentForSapLog(item)
	// 	sapReadyAttachments = append(sapReadyAttachments, tmpSapAttachmentObj)
	// }
	// log.Println(sapReadyEmails[25])
	//log.Println(sapReadyAttachments[25].DOWNDATE)
	//sapcomm.Sapcomm(sapReadyEmails, sapReadyAttachments)
}

func checkErr(err error, args ...string) {
	if err != nil {
		fmt.Println("Error")
		fmt.Printf("%v: %v", err, args)
	}
}
