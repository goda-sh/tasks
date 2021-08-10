package tasks

import (
	"fmt"
	"log"
	"strconv"

	"github.com/go-redis/redis/v8"
	"pkg.goda.sh/utils"
)

// Counter implements Atomic counters for HTTP hooks
func Counter(args *TaskArgs) Result {
	token := args.Task.Params["token"].(string)
	if len(token) > 0 {
		utils.AddAtomicCallback(token, func(ac *utils.AtomicCounter) {
			result := NewResult(args.Task)
			result.Update = struct {
				Count int64 `json:"count"`
			}{
				Count: ac.Get(),
			}
			args.Callback(result)
		})
	} else {
		return Result{
			Error: fmt.Errorf("missing counter token"),
		}
	}
	return Result{}
}

// RedisCounter uses Radis for tracking counts
func RedisCounter(args *TaskArgs) Result {
	if args.Redis.Enabled {
		if err := args.Redis.Client.Ping(args.Redis.Context).Err(); err != nil {
			return Result{
				Error: fmt.Errorf(err.Error()),
			}
		}
		token := args.Task.Params["token"].(string)
		if len(token) > 0 {
			if val, err := args.Redis.Client.Get(args.Redis.Context, token).Result(); err != redis.Nil || err != nil {
				if i, err := strconv.Atoi(val); err == nil {
					utils.AtomicCount(token, int64(i))
				}
			}
			utils.AddAtomicCallback(token, func(ac *utils.AtomicCounter) {
				count := ac.Get()
				result := NewResult(args.Task)
				result.Update = struct {
					Count int64 `json:"count"`
				}{
					Count: count,
				}
				if err := args.Redis.Client.Set(args.Redis.Context, token, count, 0).Err(); err != nil {
					log.Println(err)
				}
				args.Callback(result)
			})
		} else {
			return Result{
				Error: fmt.Errorf("missing counter token"),
			}
		}
	} else {
		return Result{
			Error: fmt.Errorf("redis is not enabled"),
		}
	}
	return Result{}
}
