package tasks

import (
	"fmt"
	"math/rand"
	"net"
	"time"
)

// Port checks if a port is open on a target host
func Port(args *TaskArgs) Result {
	result := NewResult(args.Task)
	method := "tcp"
	timeout := 10
	if m, ok := args.Task.Params["method"].(string); ok {
		method = m
	}
	if t, ok := args.Task.Params["timeout"].(float64); ok {
		timeout = int(t)
	}
	conn, err := net.DialTimeout(method, args.Task.Params["target"].(string), time.Duration(timeout)*time.Second)
	if err != nil {
		result.Error = err
	} else if conn != nil {
		conn.Close()
	}
	result.Warn = conn == nil || result.Error != nil
	if result.Warn {
		result.Notification = fmt.Sprintf("Port checker could not connect to %s target within %d seconds!", method, timeout)
	}
	result.Update = struct {
		Connected bool `json:"connected"`
	}{
		Connected: !result.Warn,
	}
	return result
}

// FakePort generates a fake port result for demo dashboards
func FakePort(args *TaskArgs) Result {
	result := NewResult(args.Task)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	conn := r.Intn(100)
	result.Warn = conn%2 == 0
	if result.Warn {
		result.Notification = fmt.Sprintf("Port checker could not connect to %s target within %d seconds!", "FAKE", 10)
	}
	result.Update = struct {
		Connected bool `json:"connected"`
	}{
		Connected: !result.Warn,
	}
	return result
}
