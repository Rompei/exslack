package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	homedir "github.com/mitchellh/go-homedir"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

// Config is object for config file.
type Config struct {
	WebHookURL  string `json:"webHookURL"`
	Destination string `json:"destination"`
	LogFile     string `json:"logFile"`
}

// WebHookBody is body of slack webhook.
type WebHookBody struct {
	Text      string `json:"text"`
	Channel   string `json:"channel"`
	Username  string `json:"username"`
	IconEmoji string `json:"icon_emoji"`
}

func main() {
	var (
		logFilePath string
	)

	flag.StringVar(&logFilePath, "log", "", "If you need output of commands, please set this flag or set from config file.")
	flag.Parse()

	stdLogger := log.New(os.Stdout, "exslack: ", log.LstdFlags)

	// Reading config file from ~/.exslackrc
	if num := flag.NArg(); num != 1 {
		stdLogger.Fatalf("The number of arguments is wrong %d", num)
	}
	configFilePath, err := homedir.Expand("~/.exslackrc")
	if err != nil {
		panic(err)
	}
	f, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		stdLogger.Fatal("Config file ~/.exslackrc was not found")
	}
	var config Config
	err = json.Unmarshal(f, &config)
	if err != nil {
		stdLogger.Fatalf("Can't read config file %s", err.Error)
	}

	if config.WebHookURL == "" || config.Destination == "" {
		stdLogger.Fatalf("Config file is not valid.")
	}

	// If there is log output, open log file.
	var logFile *os.File
	var fileLogger *log.Logger
	if logFilePath != "" {
		logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			stdLogger.Fatalf("Can't open log file %s", logFilePath)
		}
		defer logFile.Close()
		fileLogger = log.New(logFile, "exslack: ", log.LstdFlags)
	} else if config.LogFile != "" {
		logFile, err = os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			stdLogger.Fatalf("Can't open log file %s", config.LogFile)
		}
		defer logFile.Close()
		fileLogger = log.New(logFile, "exslack: ", log.LstdFlags)
	}

	// Reading commands from file.
	commandFile := flag.Arg(0)
	f, err = ioutil.ReadFile(commandFile)
	if err != nil {
		if fileLogger != nil {
			fileLogger.Printf("Command file %s was not found", commandFile)
		}
		stdLogger.Fatalf("Command file %s was not found", commandFile)
	}
	_commands := strings.Split(strings.Trim(string(f), "\n"), "\n")
	commands := make([][]string, len(_commands))
	for i := range _commands {
		commands[i] = strings.Fields(_commands[i])
	}

	// Executing commands.
	for i := range commands {
		start := time.Now()
		var output []byte
		var err error
		if fileLogger != nil {
			output, err = exec.Command(commands[i][0], commands[i][1:]...).CombinedOutput()
		} else {
			err = exec.Command(commands[i][0], commands[i][1:]...).Run()
		}
		elapsed := time.Now().Sub(start)
		text := buildText(commands[i], &start, &elapsed, output, err)
		body := &WebHookBody{
			Text:    text,
			Channel: config.Destination,
		}

		// Output log to stdout.
		stdLogger.Println(text)

		// If logFile is opened, output to log file.
		if fileLogger != nil {
			if err != nil {
				fileLogger.Printf("%s\n\n", text)
			} else {
				fileLogger.Printf("%s\n\n == Output start == \n\n%s\n\n == Output end == \n\n", text, string(output))
			}
		}

		// Post to Slack.
		if err = postToSlack(config.WebHookURL, body); err != nil {
			stdLogger.Fatal("failed to post to Slack.")
		}
	}
}

func buildText(command []string, start *time.Time, elapsed *time.Duration, output []byte, err error) string {
	if err == nil {
		return fmt.Sprintf("Command %s started on %s is done in %s", strings.Join(command, " "), start, elapsed)
	}
	if len(output) == 0 {
		return fmt.Sprintf("Command %s started on %s is done in %s with error, %s", strings.Join(command, " "), start, elapsed, err.Error)
	}
	return fmt.Sprintf("Command %s started on %s is done in %s with error, %s", strings.Join(command, " "), start, elapsed, string(output))
}

func postToSlack(url string, body *WebHookBody) error {

	// Posting slack incoming webhooks.

	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(b)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
