// Inspiration from http://blog.zikes.me/post/how-i-ruined-office-productivity-with-a-slack-bot/
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
)

var (
	botToken           string
	faceSwapperCommand string
)

// Start thg slackbot and listen to messages where this bot is mentioned
func main() {
	var (
		isDebug  bool
	)

	flag.StringVar(&botToken, "token", "", "Your SlackBot Token")
	flag.StringVar(&faceSwapperCommand, "faceSwapper", "", "The command to use for swapping out faces. Should take the input file as an argument")
	flag.BoolVar(&isDebug, "debug", false, "Debug")
	flag.Parse()

	if botToken == "" {
		fmt.Println("Slack SlackBot token cannot be empty")
		return
	}

	if faceSwapperCommand == "" {
		fmt.Println("No face swapper command detected")
		return
	}

	fileTypes := []string{"jpg", "png"}

	api := slack.New(botToken)
	logger := log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)
	slack.SetLogger(logger)
	api.SetDebug(isDebug)

	rtm := api.NewRTM()
	go rtm.ManageConnection()

	// Listen for real-time-messages
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			info := rtm.GetInfo()
			botId := fmt.Sprintf("<@%s>", info.User.ID)

			// If the posting user is _not_ the bot, the message is a file_share, and the file is an image then process it
			if ev.User != info.User.ID && strings.Contains(ev.Text, botId) && ev.SubType == "file_share" {
				index := inArray(ev.File.Filetype, fileTypes)

				if index > -1 {
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

// Checks if the value is in the passed array
// Returns -1 if not found, else the index of the array
func inArray(val interface{}, array interface{}) (index int) {
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				return
			}
		}
	}

	return
}

// Download the image from Slack and pass it to the face swapper
// Upload the manipulated image back to slack
func handleFile(rtm *slack.RTM, ev *slack.MessageEvent) {
	// Download image from private url to temp file
	file := SaveTempFile(GetFile(ev.File))

	defer os.Remove(file.Name()) // clean up

	// Pass the temp file to the face recognition executable
	swappedFile := FaceSwap(file)

	defer os.Remove(swappedFile.Name()) // clean up

	// Get resulting image and upload it to the channel
	params := slack.FileUploadParameters{
		Title: "Foo",
		Reader: swappedFile,
		Filename: "foo",
		Channels: []string{ev.Channel},
	}
	uploaded, err := rtm.UploadFile(params)

	if uploaded == nil && err != nil {
		log.Fatalf("error uploading file\n%v", err)
	}
}

// Download a file from Slack
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

// Saves the downloaded file to a temporary file and returns it
func SaveTempFile(b []byte) *os.File {
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
	return file
}

// Swap the faces using the face swapper command
func FaceSwap(file *os.File) *os.File {
	tmpFile, err := ioutil.TempFile("", "slack_image")
	if err != nil {
		log.Fatalf("error saving file: %s", err)
	}

	defer os.Remove(tmpFile.Name()) // clean up

	// Stupid TempFile cannot be prefixed
	outName := tmpFile.Name() + ".png"

	err = os.Rename(tmpFile.Name(), outName)
	if err != nil {
		log.Fatalf("error renaming file: %s", err)
	}

	cmd := exec.Command("python", faceSwapperCommand, file.Name(), "-o", outName)

	outFile, err := os.Open(outName)
	if err != nil {
		log.Fatalf("error opening the file: %s", err)
	}

	err = cmd.Run()
	if err != nil {
		log.Fatalf("Coulnd't swap faces: %s %s", file.Name(), err)
	}

	return outFile
}
