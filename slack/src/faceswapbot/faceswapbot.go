package main

import (
	"flag"
	"fmt"
	"github.com/nlopes/slack"
	"log"
	"os"
	"strings"
	"reflect"
	"net/http"
	"time"
	"io/ioutil"
	"os/exec"
	"bytes"
)

var (
	botToken string
	faceSwapExec string
)

func main() {
	var (
		isDebug  bool
	)

	flag.StringVar(&botToken, "token", "", "Your SlackBot Token")
	flag.BoolVar(&isDebug, "debug", false, "Debug")
	flag.Parse()

	if botToken == "" {
		fmt.Println("Slack SlackBot token cannot be empty")
		return
	}

	fileTypes := []string{"jpg", "png"}

	api := slack.New(botToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(isDebug)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	for msg := range rtm.IncomingEvents {
		//fmt.Print("Event Received: ")
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			info := rtm.GetInfo()
			botId := fmt.Sprintf("<@%s>", info.User.ID)

			if ev.User != info.User.ID && strings.Contains(ev.Text, botId) && ev.SubType == "file_share" {
				isFileAllowed, _ := inArray(ev.File.Filetype, fileTypes)

				if isFileAllowed {
					// Handle the file
					handleFile(rtm, ev)
				}
			}

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return

		default:
			// Ignore other events..
		}
	}
}

func inArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

func handleFile(rtm *slack.RTM, ev *slack.MessageEvent) {
	// Download image from private url to temp file
	file := SaveTempFile(GetFile(ev.File))

	// Pass the temp file to the face recognition executable
	swappedFile := FaceSwap(file)

	// Get resulting image and upload it to the channel
	params := slack.FileUploadParameters{
		Title: "Foo",
		Reader: bytes.NewReader(swappedFile),
	}
	rtm.UploadFile(params)
}

func GetFile(file *slack.File) []byte {
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	request, err := http.NewRequest(http.MethodGet, file.URLPrivateDownload, nil)
	if err != nil {
		log.Fatalf("Error creating request %s", err)
	}
	request.Header.Add("Authorization", "Bearer "+botToken)

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("error downloading file\n%v\n%v", file, err)
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("error downloading file\n%v\n%v", file, err)
	}

	return body
}

func SaveTempFile(b []byte) string {
	file, err := ioutil.TempFile("", "slack_image")
	if err != nil {
		log.Fatalf("error saving file: %s", err)
	}
	if _, err = file.Write(b); err != nil {
		log.Fatalf("error writing file: %s", err)
	}
	if err = file.Close(); err != nil {
		log.Fatalf("error closing file: %s", err)
	}
	return file.Name()
}

func FaceSwap(file string) []byte {
	out, err := exec.Command(faceSwapExec, file).Output()
	if err != nil {
		log.Fatalf("Coulnd't swap faces: %s %s", file, err)
	}

	return out
}
