package sapcomm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"example.com/config"
)

func Sapcomm(arrayOfEmails []SapEmailItem, arrayOfAttachments []SapAttachmentItem) {
	c, _ := config.GetConfig()

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	client := http.Client{
		Jar: jar,
	}

	requestEmailBodyByteArr, _ := json.Marshal(arrayOfEmails)

	requestAttachmentBodyByteArr, _ := json.Marshal(arrayOfAttachments)

	sapEmailApiUrl, _ := url.JoinPath(c.Sap.Path, "email")
	sapAttachmentApiUrl, _ := url.JoinPath(c.Sap.Path, "attachment")

	sapApi(sapEmailApiUrl, requestEmailBodyByteArr, c, client)
	sapApi(sapAttachmentApiUrl, requestAttachmentBodyByteArr, c, client)

	client.CloseIdleConnections()
}

func sapApi(urlString string, byteArray []byte, c config.Config, client http.Client) {

	var customHeader = map[string]string{"x-csrf-token": "fetch", "Content-Type": "application/json", "Connection": "keep-alive"}

	getReq, _ := http.NewRequest("GET", urlString, nil)
	getReq.Header.Set("x-csrf-token", "fetch")
	getReq.Header.Set("Content-Type", "application/json")
	getReq.Header.Set("Connection", "keep-alive")
	getReq.SetBasicAuth(c.Sap.Username, c.Sap.Password)
	resGet, _ := client.Do(getReq)
	fmt.Printf("STATUS GET requestu --- %s\n", resGet.Status)
	customHeader["x-csrf-token"] = resGet.Header.Get("x-csrf-token")
	resGet.Body.Close()

	postReq, _ := http.NewRequest("POST", urlString, bytes.NewBuffer(byteArray))
	postReq.Header.Set("x-csrf-token", resGet.Header.Get("x-csrf-token"))
	postReq.Header.Set("Content-Type", "application/json")
	postReq.Header.Set("Connection", "keep-alive")
	postReq.SetBasicAuth(c.Sap.Username, c.Sap.Password)
	resPost, _ := client.Do(postReq)
	fmt.Printf("STATUS POST requestu --- %s\n", resPost.Status)
	resPost.Body.Close()
}

type SapEmailItem struct {
	EMLUUID     string
	CLIENT      int8
	RECVDATE    string
	SUBJECT     string
	SENDER      string
	HASPDF      int8
	ATTACHMENTS string
	COMPLETED   int8
}

type SapAttachmentItem struct {
	UUID          string
	CLIENT        int8
	FROMEMAIL     string
	DOWNDATE      string
	OGFILENAME    string
	NEWFILENAME   string
	FILEPROCESSED int8
}
