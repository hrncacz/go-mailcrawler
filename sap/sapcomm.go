package sapcomm

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/cookiejar"

	"example.com/config"
)

func Sapcomm(arrayOfEmails []SapEmailItem) {
	// testBodyArray := &[]SapEmail{{
	// 	EMLUUID:     "0c0aacc3-1213-4b40-914d-afb544eaertzu5",
	// 	CLIENT:      100,
	// 	RECVDATE:    "20221105",
	// 	SUBJECT:     "tEST MAIL 2",
	// 	SENDER:      "m.hrncirik@vollmann-group.com",
	// 	HASPDF:      1,
	// 	ATTACHMENTS: "0c0aacc3-1213-4b40-914d-afb544ea616e,0c0aacc3-1213-4b40-914d-afb544ea616e,0c0aacc3-1213-4b40-914d-afb544ea616e",
	// 	COMPLETED:   0}, {
	// 	EMLUUID:     "0c0aacc3-1213-4b40-914d-afb54478954436",
	// 	CLIENT:      100,
	// 	RECVDATE:    "20221107",
	// 	SUBJECT:     "tEST MAIL 3",
	// 	SENDER:      "m.hrncirik@vollmann-group.com",
	// 	HASPDF:      1,
	// 	ATTACHMENTS: "0c0aacc3-1213-4b40-914d-afb544ea616e,0c0aacc3-1213-4b40-914d-afb544ea616e,0c0aacc3-1213-4b40-914d-afb544ea616e",
	// 	COMPLETED:   0}}
	c, _ := config.GetConfig()

	requestBodyByteArr, err := json.Marshal(arrayOfEmails)

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}

	urlPath := c.Sap.Path

	client := http.Client{
		Jar: jar,
	}

	var customHeader = map[string]string{"x-csrf-token": "fetch", "Content-Type": "application/json", "Connection": "keep-alive", c.Sap.Username: c.Sap.Password}

	getReq, _ := http.NewRequest("GET", urlPath, nil)
	getReq = configHeader(customHeader, getReq)
	res, _ := client.Do(getReq)
	customHeader["x-csrf-token"] = res.Header.Get("x-csrf-token")
	res.Body.Close()

	postReq, _ := http.NewRequest("POST", urlPath, bytes.NewBuffer(requestBodyByteArr))
	postReq = configHeader(customHeader, postReq)
	res2, _ := client.Do(postReq)
	log.Println(res2.Status)
	res2.Body.Close()
}

func configHeader(customHeader map[string]string, req *http.Request) *http.Request {
	for key, value := range customHeader {
		req.Header.Set(key, value)
	}
	return req
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
