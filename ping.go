package tasks

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-ping/ping"
	"pkg.goda.sh/utils"
)

// Ping sends ICMP requests to a specific target host
func Ping(args *TaskArgs) Result {
	result := NewResult(args.Task)
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"count": 3,
		"high":  75,
	})
	pinger, err := ping.NewPinger(params.Get("target").String())
	if err != nil {
		result.Error = err
	}
	pinger.Count = int(params.Get("count").Int64())
	pinger.SetPrivileged(true)
	err = pinger.Run() // Blocks until finished.
	if err != nil {
		result.Error = err
	}
	if result.Error == nil {
		pinged := pinger.Statistics()
		avg := int(pinged.AvgRtt / time.Millisecond)
		result.Warn = avg >= int(params.Get("high").Int64())
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
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"high":  75,
		"range": []int64{
			1, 100,
		},
	})
	high := int(params.Get("high").Int64())
	mm := params.Get("range").Ints()
	min := int(mm[0])
	max := int(mm[1])
	recv := r.Intn(max-min) + min
	avg := int(time.Duration((r.Intn(max-min)+min)*1000) / time.Microsecond)
	result.Warn = avg >= high
	loss := 0
	if r.Intn(100) >= high {
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
