package tasks

import (
	"math/rand"

	"pkg.goda.sh/utils"
)

// Media lets you embed iframes, images, videos etc. on in a dashboard
func Media(args *TaskArgs) Result {
	if args.Task.Once {
		args.Stop()
	}
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"type": "iframe",
	})
	result := NewResult(args.Task)
	result.Update = struct {
		URL  string `json:"url"`
		Type string `json:"type"`
	}{
		URL: Templater(params.Get("url").String(), struct {
			CacheKey int
		}{
			CacheKey: rand.Int(),
		}),
		Type: params.Get("type").String(),
	}
	return result
}
