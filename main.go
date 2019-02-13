package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"text/template"
	"time"
)

const apikeyBot string = "[apikeyOfTheUserWhoWantToPost]"
const serverURL = "[yourServerURL]"

type commitMessage struct {
	Apikey      string
	WorkPackage string
	Comment     string
	ServerURL   string
}

func main() {
	cmsg := commitMessage{}

	var msgToParse string

	flag.StringVar(&cmsg.Comment, "m", "", "Comment")
	flag.StringVar(&cmsg.WorkPackage, "t", "", "Workpackage to comment to")
	flag.StringVar(&cmsg.Apikey, "k", apikeyBot, "Apikey from Openproject")
	flag.StringVar(&cmsg.ServerURL, "s", serverURL, "Openproject server url")
	flag.StringVar(&msgToParse, "p", "", "Full commit message. The string needs to be formated in the form [TICKETNO]# [TEXT] ")

	flag.Parse()

	if msgToParse != "" {
		cmsg.WorkPackage, cmsg.Comment = parseMessage(msgToParse)
	}

	urlt := "{{.ServerURL}}/api/v3/work_packages/{{.WorkPackage}}/activities?notify=false"
	url := parseTemplate(urlt, cmsg)

	msgt := (`{"comment": { "raw": "{{.Comment}}" } }`)
	msg := parseTemplate(msgt, cmsg)

	data := []byte(msg)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Basic auth is done by username apikey and the apikey as password
	req.SetBasicAuth("apikey", cmsg.Apikey)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	fmt.Println("Sending request...")
	fmt.Println(req.URL)
	fmt.Println(req.Header)
	fmt.Println(req.Body)
	fmt.Println("-------------------------------------------------------------")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error reading response.", err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading body.", err)
	}

	fmt.Printf("%s\n", body)
}

func parseMessage(msg string) (wpNr, cmtmsg string) {
	fmt.Println("parsing...")
	spMsg := strings.SplitAfterN(msg, "#", 2)
	if len(spMsg) != 2 {
		fmt.Println("Malformated commit message")
		os.Exit(0)
	}
	return strings.TrimSuffix(spMsg[0], "#"), spMsg[1]
}

func parseTemplate(temp string, vals commitMessage) string {
	t := template.New("")
	t, _ = t.Parse(temp)
	var buf bytes.Buffer
	if err := t.Execute(&buf, vals); err != nil {
		log.Fatal("Template executing error")
	}
	return buf.String()
}
