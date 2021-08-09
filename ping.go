package tasks

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-ping/ping"
)

// Ping sends ICMP requests to a specific target host
func Ping(args *TaskArgs) Result {
	result := NewResult(args.Task)
	pinger, err := ping.NewPinger(args.Task.Params["target"].(string))
	if err != nil {
		result.Error = err
	}
	pinger.Count = int(args.Task.Params["count"].(float64))
	pinger.SetPrivileged(true)
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		result.Error = err
	}
	if result.Error == nil {
		pinged := pinger.Statistics()
		avg := int(pinged.AvgRtt / time.Millisecond)
		if high, ok := args.Task.Params["high"].(float64); ok {
			result.Warn = avg >= int(high)
		}
		if pinged.PacketLoss > 0 {
			result.Warn = true
		}
		result.Spark = &Spark{
			avg,
			result.Warn,
		}
		if result.Warn {
			result.Notification = fmt.Sprintf("ping of %dms detected with %f%% packet loss!", avg, pinged.PacketLoss)
		}
		result.Update = struct {
			Sent int     `json:"sent"`
			Recv int     `json:"recv"`
			Loss float64 `json:"loss"`
			Avg  int     `json:"avg"`
			Jitt int     `json:"jitt"`
		}{
			Sent: pinged.PacketsSent,
			Recv: pinged.PacketsRecv,
			Loss: pinged.PacketLoss,
			Avg:  avg,
			Jitt: int(pinged.StdDevRtt / time.Millisecond),
		}
	}
	return result
}

// FakePing generates fake data for demo dashboards
func FakePing(args *TaskArgs) Result {
	result := NewResult(args.Task)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	mm := args.Task.Params["range"].([]interface{})
	min := int(mm[0].(float64))
	max := int(mm[1].(float64))
	recv := r.Intn(max-min) + min
	avg := int(time.Duration((r.Intn(max-min)+min)*1000) / time.Microsecond)
	result.Warn = avg >= int(args.Task.Params["high"].(float64))
	loss := 0
	if r.Intn(100) >= 75 {
		loss = max - recv
	}
	if loss > 0 {
		result.Warn = true
	}
	if result.Warn {
		result.Notification = fmt.Sprintf("ping of %dms detected with %d%% packet loss!", avg, loss)
	}
	result.Spark = &Spark{
		avg,
		result.Warn,
	}
	result.Update = struct {
		Sent int `json:"sent"`
		Recv int `json:"recv"`
		Loss int `json:"loss"`
		Avg  int `json:"avg"`
		Jitt int `json:"jitt"`
	}{
		Sent: max,
		Recv: recv,
		Loss: loss,
		Avg:  avg,
		Jitt: int(time.Duration(r.Intn(50000)) / time.Microsecond),
	}
	return result
}
