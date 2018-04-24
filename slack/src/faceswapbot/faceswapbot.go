// Inspiration from http://blog.zikes.me/post/how-i-ruined-office-productivity-with-a-slack-bot/
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"reflect"
	"net/http"
	"time"
	"io/ioutil"
	"github.com/nlopes/slack"
	"../github.com/digiaonline/faceswapbot"
	"github.com/spf13/viper"
)

var (
	botToken    string
	faceSwapper faceswapbot.FaceReplacer
)

// Start thg slackbot and listen to messages where this bot is mentioned
func main() {
	var (
		config  string
		isDebug bool
	)

	//flag.StringVar(&botToken, "token", "", "Your SlackBot Token")
	flag.StringVar(&config, "config", "", "Config file")
	flag.BoolVar(&isDebug, "debug", false, "Debug")
	flag.Parse()

	viper.SetConfigFile(config)
	err := viper.ReadInConfig()
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	botToken = viper.GetString("slack.token")

	if botToken == "" {
		panic(fmt.Errorf("Slack token cannot be empty\n"))
	}

	faceSwapper = faceswapbot.PySwapper{
		Executable:  viper.GetString("pyswapper.paths.executable"),
		FacesPath:   viper.GetString("pyswapper.paths.faces"),
		BodiesPath:  viper.GetString("pyswapper.paths.bodies"),
		SuccessPath: viper.GetString("pyswapper.paths.success"),
	}

	// Todo: should come from config allow images by some other method
	fileTypes := []string{"jpg", "png", "gif"}

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
			if ev.User != info.User.ID && strings.Contains(ev.Text, botId) {
				if ev.SubType == "file_share" {
					index := inArray(ev.File.Filetype, fileTypes)

					// If the file is of supported type
					if index > -1 {
						// Handle the file
						if strings.Contains(strings.ToLower(ev.Text), "bomb") {
							handleFile(rtm, ev, "bomb")
						} else if strings.Contains(strings.ToLower(ev.Text), "success") {
							handleFile(rtm, ev, "success")
						} else {
							handleFile(rtm, ev, "")
						}
					} else {
						rtm.SendMessage(rtm.NewOutgoingMessage(
							"Supported file types are: " + strings.Join(fileTypes, ", "),
							ev.Channel,
						))
					}
				} else if strings.Contains(strings.ToLower(ev.Text), "help") {
					rtm.SendMessage(rtm.NewOutgoingMessage(
						"Available commands are:\n"+
							"No parameters for face swapping purpouses :facepalm:\n"+
							"`bomb` to explicitly photobomb the image :bomb:\n"+
							"`success` to _successify_ the image :success:",
						ev.Channel,
					))
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
func handleFile(rtm *slack.RTM, ev *slack.MessageEvent, command string) {
	// Download image from private url to temp file
	file := SaveTempFile(GetFile(ev.File))

	defer os.Remove(file.Name()) // clean up

	var swappedFile *os.File
	var err error

	// Pass the temp file to the face recognition executable
	switch command {
	case "success":
		swappedFile, err = faceSwapper.Success(file)
	case "bomb":
		swappedFile, err = faceSwapper.PhotoBomb(file)
	default:
		swappedFile, err = faceSwapper.SwapFaces(file)
	}

	if err != nil {
		log.Fatalf("error swapping faces: %s", err)
	}

	defer os.Remove(swappedFile.Name()) // clean up

	// Get resulting image and upload it to the channel
	params := slack.FileUploadParameters{
		Title:    "Foo",
		Reader:   swappedFile,
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
