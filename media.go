package tasks

import (
	"math/rand"
)

// Media lets you embed iframes, images, videos etc. on in a dashboard
func Media(args *TaskArgs) Result {
	if args.Task.Once {
		args.Stop()
	}
	t := "iframe"
	if args.Task.Params["type"] != nil {
		t = args.Task.Params["type"].(string)
	}
	result := NewResult(args.Task)
	result.Update = struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}{
		URL: Templater(args.Task.Params["url"].(string), struct {
			CacheKey int
		}{
			CacheKey: rand.Int(),
		}),
		Type: t,
	}
	return result
}
