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
	"io"
	"bytes"
	"math/rand"
	"bufio"
	"path/filepath"
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
						// Todo, we need a command parser to extract out the possible commands
						if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(ev.Text)), "BOMB") {
							handleFile(rtm, ev, "bomb", false)
						} else if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(ev.Text)), "BOMB BW") {
							handleFile(rtm, ev, "bomb", true)
						} else if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(ev.Text)), "SUCCESS") {
							handleFile(rtm, ev, "success", false)
						} else if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(ev.Text)), "SUCCESS BW") {
							handleFile(rtm, ev, "success", true)
						} else if strings.HasSuffix(strings.ToUpper(strings.TrimSpace(ev.Text)), "BW") {
							handleFile(rtm, ev, "", true)
						} else {
							handleFile(rtm, ev, "", false)
						}
					} else {
						rtm.SendMessage(rtm.NewOutgoingMessage(
							"Supported file types are: "+strings.Join(fileTypes, ", "),
							ev.Channel,
						))
					}
				} else {
					botIdStr := fmt.Sprintf("@%s", info.User.Name)
					rtm.SendMessage(rtm.NewOutgoingMessage(
						"Ad `" + botIdStr + "` as a comment to image uploads for face swapping purposes :facepalm:\n" +
							"Additional commands:\n"+
							"`bomb` to explicitly photobomb the image :bomb:\n"+
							"`success` to _successify_ the image :success:\n" +
							"You can append `bw` to use a Black&White filter on the image\n" +
							"e.g. `" + botIdStr + " bomb bw`",
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
func handleFile(rtm *slack.RTM, ev *slack.MessageEvent, command string, bw bool) {
	var swappedFile *os.File
	var err error

	// Download image from private url to temp file
	file := saveTempFile(getFileData(ev.File))
	tempFileName := file.Name() + "." + ev.File.Filetype
	err = os.Rename(file.Name(), tempFileName)
	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(
			":robot_face: + :bug: :arrow_right: :feelsgood:, plz contact your local IT support for support :troll:",
			ev.Channel,
		))
		log.Printf("error swapping faces: %s\n", err)
		return
	}

	file, _ = os.Open(tempFileName)

	defer os.Remove(tempFileName) // clean up

	// Pass the temp file to the face recognition executable
	switch command {
	case "success":
		swappedFile, err = faceSwapper.Success(file, bw)
	case "bomb":
		swappedFile, err = faceSwapper.PhotoBomb(file, bw)
	default:
		swappedFile, err = faceSwapper.SwapFaces(file, bw)
	}

	if err != nil {
		rtm.SendMessage(rtm.NewOutgoingMessage(
			":robot_face: + :bug: :arrow_right: :feelsgood:, plz contact your local IT support for support :troll:",
			ev.Channel,
		))
		log.Printf("error swapping faces: %s\n", err)
		return
	}

	defer os.Remove(swappedFile.Name()) // clean up

	fileName, _ := getRandomFileName()

	// Get resulting image and upload it to the channel
	params := slack.FileUploadParameters{
		Title:    fileName,
		Reader:   swappedFile,
		Filename: fileName + filepath.Ext(swappedFile.Name()),
		Channels: []string{ev.Channel},
	}
	uploaded, err := rtm.UploadFile(params)

	if uploaded == nil && err != nil {
		log.Fatalf("error uploading file\n%v", err)
	}
}

// Download a file from Slack and return its content
func getFileData(file *slack.File) []byte {
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
func saveTempFile(b []byte) *os.File {
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

// Get a random file name
func getRandomFileName() (string, error) {
	var (
		line string
	)

	//_, filename, _, _ := runtime.Caller(0)
	//file, _ := os.Open(path.Join(path.Dir(filename), "/data/names.txt"))
	file, _ := os.Open(viper.GetString("filenames"))
	defer file.Close()

	numberOfLines, _ := lineCounter(file)

	randLineNumber := random(0, numberOfLines-1)

	file.Seek(0, 0)
	scanner := bufio.NewScanner(file)

	i := 0
	for scanner.Scan() {
		line = scanner.Text()
		if i >= randLineNumber && line != "" {
			return line, nil
		}
		i++
	}

	return "", scanner.Err()
}

// https://stackoverflow.com/questions/24562942/golang-how-do-i-determine-the-number-of-lines-in-a-file-efficiently
func lineCounter(r io.Reader) (int, error) {
	buf := make([]byte, 32*1024)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		count += bytes.Count(buf[:c], lineSep)

		switch {
		case err == io.EOF:
			return count, nil

		case err != nil:
			return count, err
		}
	}
}

// http://golangcookbook.blogspot.fi/2012/11/generate-random-number-in-given-range.html
func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
