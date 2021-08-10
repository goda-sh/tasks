package tasks

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"pkg.goda.sh/utils"
)

// Port checks if a port is open on a target host
func Port(args *TaskArgs) Result {
	result := NewResult(args.Task)
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"method":  "tcp",
		"timeout": 10,
	})
	method := params.Get("method").String()
	target := params.Get("target").String()
	timeout := params.Get("timeout").Int64()
	conn, err := net.DialTimeout(method, target, time.Duration(timeout)*time.Second)
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
