package tasks

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	Project     = "GoDash"
	Version     = "0.0.0-beta.0"
	CDN         = ""
	TaskMapping = map[string][]string{}
	TaskRunners = map[string]Type{
		"port":          {Port, "port", false},
		"fakeport":      {FakePort, "port", false},
		"ping":          {Ping, "ping", false},
		"http":          {HTTP, "http", false},
		"http-json":     {HTTPJSON, "http", false},
		"http-status":   {HTTPStatus, "http", false},
		"http-regex":    {HTTPREGEXP, "http", false},
		"http-regexp":   {HTTPREGEXP, "http", false},
		"fakeping":      {FakePing, "ping", false},
		"media":         {Media, "media", false},
		"iframe":        {Media, "media", false},
		"feed":          {Feed, "feed", false},
		"fakefeed":      {FakeFeed, "feed", false},
		"dns":           {DNS, "dns", false},
		"dns-cidr":      {DNSCIDR, "dns", false},
		"counter":       {Counter, "counter", true},      // Triggered by callbacks
		"redis-counter": {RedisCounter, "counter", true}, // Triggered by callbacks
	}
)

func init() {
	// Create map of all the task runners.
	for id, t := range TaskRunners {
		TaskMapping[t.Type] = append(TaskMapping[t.Type], id)
	}
}

// Task is the Job
type Task struct {
	Label     string                 `json:"label"`
	Interval  string                 `json:"interval"`
	Task      string                 `json:"task"`
	ID        string                 `json:"id"`
	Date      int64                  `json:"date"`
	Location  string                 `json:"location,omitempty"`
	Once      bool                   `json:"once,omitempty"`
	Spark     *Spark                 `json:"spark"`
	Warn      bool                   `json:"warn,omitempty"`
	Last      interface{}            `json:"last,omitempty"`
	Cancelled bool                   `json:"cancelled,omitempty"`
	Params    map[string]interface{} `json:"params,omitempty"`
	CTX       context.Context        `json:"-"`
	Cancel    func() bool            `json:"-"`
}

// CleanTask is for JSON converting
type CleanTask struct {
	Label     string                 `json:"label"`
	Interval  string                 `json:"interval"`
	Task      string                 `json:"task"`
	ID        string                 `json:"id"`
	Date      int64                  `json:"date"`
	Location  string                 `json:"location,omitempty"`
	Once      bool                   `json:"once,omitempty"`
	Spark     *Spark                 `json:"spark,omitempty"`
	Warn      bool                   `json:"warn,omitempty"`
	Last      interface{}            `json:"last,omitempty"`
	Cancelled bool                   `json:"cancelled,omitempty"`
	Params    map[string]interface{} `json:"-"`
	CTX       context.Context        `json:"-"`
	Cancel    func() bool            `json:"-"`
}

// Hash is used to create task hashes
type Hash struct {
	Label    string `json:"label"`
	Interval string `json:"interval"`
	Task     string `json:"task"`
	ID       string `json:"id,omitempty"`
	Once     bool   `json:"once,omitempty"`
	Machine  string `json:"machine,omitempty"`
}

// Spark is a value and warn flag for the UI
type Spark struct {
	Value interface{} `json:"value,omitempty"`
	Warn  bool        `json:"warn,omitempty"`
}

// Result is the results from a task
type Result struct {
	Task         string      `json:"task"`
	Label        string      `json:"label"`
	ID           string      `json:"id"`
	Date         int64       `json:"date"`
	Notification string      `json:"notification,omitempty"`
	Location     string      `json:"location,omitempty"`
	Spark        *Spark      `json:"spark,omitempty"`
	Warn         bool        `json:"warn,omitempty"`
	Update       interface{} `json:"update,omitempty"`
	Error        error       `json:"error"`
	ErrorString  string      `json:"errormsg,omitempty"`
	Event        string      `json:"event,omitempty"`
	Cancelled    bool        `json:"-"`
}

// Templater lets tasks format strings using dynamic structs
func Templater(str string, data interface{}) string {
	if t, err := template.New("templater").Parse(str); err == nil {
		var tpl bytes.Buffer
		if err := t.Execute(&tpl, data); err == nil {
			return tpl.String()
		}
	}
	return str
}

// NewResult generates a new result object for tasks to return
func NewResult(task Task) Result {
	return Result{
		Task:      task.Task,
		Label:     task.Label,
		ID:        task.ID,
		Date:      time.Now().UnixNano() / int64(time.Millisecond),
		Cancelled: false,
		Event:     "update", // Default SSE event is "update"
	}
}

// TaskArgs is possible arguments passed to a task
type TaskArgs struct {
	Task     Task         `json:"task"`
	Callback func(Result) `json:"-"`
	Stop     func()       `json:"-"`
	Redis    Redis        `json:"-"`
}

// Redis holds the Redis client and context
type Redis struct {
	Enabled bool
	Client  *redis.Client
	Context context.Context
}

// Type binds a method to a UI type
type Type struct {
	Func      func(*TaskArgs) Result
	Type      string `json:"type"`
	Timerless bool   `json:"timerless"` // Triggered by callbacks
}

// Timerless checks if a task should use a timer to update or not
func Timerless(task string) bool {
	return TaskRunners[task].Timerless
}

// CreateRequest generates HTTP requests for tasks
func CreateRequest(method string, url string, body io.Reader) (*http.Request, *http.Client) {
	req, _ := http.NewRequest(method, url, body)
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", Project, Version)) // Add a default user agent
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, // Accept any invalid certs
			},
		},
	}
	return req, client
}
