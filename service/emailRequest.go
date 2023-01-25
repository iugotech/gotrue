package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type EmailRequestService struct {
	ApiKey string
	URL    string
}

type MailRequest struct {
	To               []string    `json:"to"`
	Bcc              []string    `json:"bcc"`
	Subject          string      `json:"subject"`
	TemplateFileName string      `json:"templateFileName"`
	TemplateData     interface{} `json:"templateData"`
}

func NewEmailRequestService() *EmailRequestService {

	return &EmailRequestService{
		ApiKey: "44ffe46c-320a-4a31-b0d4-345b75909f6e",
		URL:    "http://mail-service:9003/api/v1/mail",
	}
}

func (rs *EmailRequestService) SendEmailWithTemplate(to []string, bcc []string, subject string, templateFileName string, templateData interface{}) error {

	emailRequest := &MailRequest{
		To:               to,
		Bcc:              bcc,
		Subject:          subject,
		TemplateFileName: templateFileName,
		TemplateData:     templateData,
	}

	jsonData, err := json.Marshal(emailRequest)
	if err != nil {
		fmt.Println(err)
	}

	req, err := http.NewRequest(http.MethodPost, rs.URL, bytes.NewBuffer(jsonData))
	req.Header.Set("x-mail-api-key", rs.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	return err
}

// curl --location --request POST 'http://10.180.32.11:9003/api/v1/mail' --header 'x-mail-api-key: 44ffe46c-320a-4a31-b0d4-345b75909f6e' --header 'Content-Type: application/json' --data '{
//     "to": ["engin.ozatay@gmail.com"],
//     "bcc": [],
//     "subject": "Test",
//     "templateFileName": "resetPassword",
//     "templateData": {
//         "URL": "https://www.iugo.tech"
//     }
// }'

// curl --location --request POST 'http://104.248.249.210:9003/api/v1/mail' --header 'x-mail-api-key: 44ffe46c-320a-4a31-b0d4-345b75909f6e' --header 'Content-Type: application/json' --data '{
//     "to": ["engin.ozatay@gmail.com"],
//     "bcc": [],
//     "subject": "Test",
//     "templateFileName": "resetPassword",
//     "templateData": {
//         "URL": "https://www.iugo.tech"
//     }
// }'
