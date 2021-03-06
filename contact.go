package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"text/template"
	"time"
)

type recaptchaResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []int     `json:"error-codes"`
}

type EmailData struct {
	SenderEmail    string
	RecipientEmail string
	Name           string
	ReturnEmail    string
	Message        string
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		log.Print("Received contact POST Request from " + r.RemoteAddr)
		contactRequest := make(map[string]string)
		r.ParseForm()
		for k, v := range r.Form {
			contactRequest[strings.TrimSpace(k)] = strings.TrimSpace(strings.Join(v, " "))
		}

		log.Print("contact request: ", contactRequest)

		if checkCaptcha(contactRequest["captcha"]).Success {
			log.Print("contact POST request successful, sending mail...")
			sendMail(contactRequest["name"],
				contactRequest["email"],
				contactRequest["message"])
		} else {
			log.Print("contact POST request with failed captcha")
		}
	}
}

func checkCaptcha(response string) (r recaptchaResponse) {
	secret, err := loadCaptchaSecret()
	if err != nil {
		log.Print("Failed to load captcha secret:", err)
		return
	}
	resp, err := http.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {secret}, "response": {response}})
	if err != nil {
		log.Print("Post error: ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Print("Read error: could not read body: ", err)
		return
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Print("Read error: got invalid JSON: ", err)
		return
	}
	return
}

func loadCaptchaSecret() (string, error) {
	captchaSecretFile := "captcha-secret.txt"
    dat, err := ioutil.ReadFile(captchaSecretFile)
    if err != nil {
    	return "", err
    }
    return strings.TrimSpace(string(dat)), nil
}

func sendMail(name string, email string, msg string) {
	context := &EmailData{
		SenderEmail:    "contact@enochtsang.com",
		RecipientEmail: "echtsang@gmail.com",
		Name:           name,
		ReturnEmail:    email,
		Message:        msg,
	}

	if strings.ContainsAny((*context).Name, "\r\n") {
		log.Print("Found malicious emailData: ", (*context).Name)
		return
	}

	client, err := smtp.Dial("127.0.0.1:25")
	check(err, false)
	defer client.Close()

	client.Mail(context.SenderEmail)
	client.Rcpt(context.RecipientEmail)

	writeCloser, err := client.Data()
	check(err, false)
	defer writeCloser.Close()

	t, err := template.ParseFiles("templates/email.tmpl")
	check(err, false)

	var tpl bytes.Buffer
	err = t.Execute(&tpl, context)
	check(err, false)

	buf := bytes.NewBufferString(tpl.String())
	if _, err = buf.WriteTo(writeCloser); err != nil {
		log.Fatal(err)
	}
}
