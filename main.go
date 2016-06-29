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
	"runtime"
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

// Options is command line options.
type Options struct {
	LogFilePath string
	IsConc      bool
	NumCPU      uint
	JobList     string
}

// Job is result of command with output.
type Job struct {
	FullCommand []string
	Command     string
	Args        []string
	Start       *time.Time
	End         *time.Time
	Elapsed     *time.Duration
	Output      []byte
	Err         error
}

// NewJob is constructor of Job.
func NewJob(fullCommand []string) *Job {
	var args []string
	if len(fullCommand) >= 2 {
		args = fullCommand[1:]
	}
	return &Job{
		FullCommand: fullCommand,
		Command:     fullCommand[0],
		Args:        args,
	}
}

func main() {
	var opts Options

	flag.StringVar(&opts.LogFilePath, "log", "", "If you need output of commands, please set this flag or set from config file.")
	flag.BoolVar(&opts.IsConc, "c", false, "Execute commands concrrentry.")
	flag.UintVar(&opts.NumCPU, "cpu", 1, "How many CPUs to execution.")
	flag.StringVar(&opts.JobList, "l", "", "List of jobs.")
	flag.Parse()

	// Decide using cpus.
	numCPUs := runtime.NumCPU()
	if int(opts.NumCPU) >= numCPUs {
		runtime.GOMAXPROCS(numCPUs)
	} else if opts.NumCPU == 0 {
		runtime.GOMAXPROCS(1)
	} else {
		runtime.GOMAXPROCS(int(opts.NumCPU))
	}

	stdLogger := log.New(os.Stdout, "exslack: ", log.LstdFlags)

	// Reading config file from ~/.exslackrc
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
	if opts.LogFilePath != "" {
		logFile, err = os.OpenFile(opts.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			stdLogger.Fatalf("Can't open log file %s", opts.LogFilePath)
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

	// Loading jobs.
	var jobs []Job
	if flag.NArg() != 0 {
		jobs = loadJobsFromArgs()
	} else if opts.JobList != "" {
		jobs = loadJobsFromFile(opts.JobList, stdLogger, fileLogger)
	} else {
		if fileLogger != nil {
			fileLogger.Println("Command must be received from -l option or file.")
		}
		stdLogger.Fatalln("Command must be received from -l option or file.")
	}

	// Execute commands.
	resCh := make(chan *Job, len(jobs))
	defer close(resCh)
	if opts.IsConc {
		for i := range jobs {
			start := time.Now()
			jobs[i].Start = &start
			if fileLogger != nil {
				go execWithOutput(&jobs[i], resCh)
			} else {
				go execWithoutOutput(&jobs[i], resCh)
			}
		}
		for i := 0; i < len(jobs); i++ {
			job := <-resCh
			postpro(&config, stdLogger, fileLogger, job)
		}
	} else {
		for i := range jobs {
			start := time.Now()
			jobs[i].Start = &start
			if fileLogger != nil {
				go execWithOutput(&jobs[i], resCh)
			} else {
				go execWithoutOutput(&jobs[i], resCh)
			}
			job := <-resCh
			postpro(&config, stdLogger, fileLogger, job)
		}
	}
}

func postpro(config *Config, stdLogger, fileLogger *log.Logger, job *Job) {
	text := buildText(job.FullCommand, job.Start, job.Elapsed, job.Err)
	body := &WebHookBody{
		Text:      text,
		Channel:   config.Destination,
		Username:  "exslack",
		IconEmoji: ":tada:",
	}
	// Output log to stdout.
	stdLogger.Println(text)

	// If logFile is opened, output to log file.
	if fileLogger != nil {
		fileLogger.Printf("%s\n\n == Output start == \n\n%s\n\n == Output end == \n\n", text, string(job.Output))
	}

	// Post to Slack.
	if err := postToSlack(config.WebHookURL, body); err != nil {
		stdLogger.Fatal("failed to post to Slack.")
	}
}

func loadJobsFromArgs() []Job {
	jobs := make([]Job, 1)
	jobs[0] = *NewJob(flag.Args())
	return jobs
}

func loadJobsFromFile(fname string, stdLogger, fileLogger *log.Logger) []Job {
	f, err := ioutil.ReadFile(fname)
	if err != nil {
		if fileLogger != nil {
			fileLogger.Printf("Command file %s was not found", fname)
		}
		stdLogger.Fatalf("Command file %s was not found", fname)
	}
	commands := strings.Split(strings.Trim(string(f), "\n"), "\n")
	if len(commands) == 0 || commands[0] == "" {
		if fileLogger != nil {
			fileLogger.Println("Command is not defined")
		}
		stdLogger.Fatalln("Command is not defined")
	}
	jobs := make([]Job, len(commands))
	for i := range commands {
		jobs[i] = *NewJob(strings.Fields(commands[i]))
	}
	return jobs
}

func execWithOutput(job *Job, resCh chan *Job) {
	output, err := exec.Command(job.Command, job.Args...).CombinedOutput()
	end := time.Now()
	job.End = &end
	elapsed := end.Sub(*job.Start)
	job.Elapsed = &elapsed
	job.Err = err
	job.Output = output
	resCh <- job
}

func execWithoutOutput(job *Job, resCh chan *Job) {
	err := exec.Command(job.Command, job.Args...).Run()
	end := time.Now()
	job.End = &end
	elapsed := end.Sub(*job.Start)
	job.Elapsed = &elapsed
	job.Err = err
	resCh <- job
}

func buildText(command []string, start *time.Time, elapsed *time.Duration, err error) string {
	if err == nil {
		return fmt.Sprintf("Command %s started on %s is done in %s", strings.Join(command, " "), start, elapsed)
	}
	return fmt.Sprintf("Command %s started on %s is done in %s with error, %s", strings.Join(command, " "), start, elapsed, err.Error())
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
