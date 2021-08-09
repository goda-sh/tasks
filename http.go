package tasks

import (
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/tidwall/gjson"
)

// HTTP pulls content from a web server
func HTTP(args *TaskArgs) Result {
	result := NewResult(args.Task)
	req, client := CreateRequest("GET", args.Task.Params["url"].(string), nil)
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.Error = err
		} else {
			c := string(contents)
			result.Warn = len(c) == 0 || !(resp.StatusCode >= 200 && resp.StatusCode <= 299)
			if result.Warn {
				result.Notification = fmt.Sprintf("an invalid status code has been found: %d", resp.StatusCode)
			}
			result.Update = struct {
				Content string `json:"content"`
			}{
				Content: c,
			}
		}
	}
	return result
}

// HTTPStatus gets the status code from a web server
func HTTPStatus(args *TaskArgs) Result {
	result := NewResult(args.Task)
	req, client := CreateRequest("HEAD", args.Task.Params["url"].(string), nil)
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
	} else {
		defer resp.Body.Close()
		if err != nil {
			result.Error = err
		} else {
			valid := false
			codes := args.Task.Params["codes"].([]interface{})
			if len(codes) == 1 {
				valid = resp.StatusCode == int(codes[0].(float64))
			} else {
				valid = resp.StatusCode >= int(codes[0].(float64)) && resp.StatusCode <= int(codes[1].(float64))
			}
			result.Warn = !valid
			if result.Warn {
				result.Notification = fmt.Sprintf("an invalid status code has been found: %d", resp.StatusCode)
			}
			result.Update = struct {
				Content string `json:"content"`
			}{
				Content: fmt.Sprintf("%d", resp.StatusCode),
			}
		}
	}
	return result
}

// HTTPJSON lets you parse JSON on a remote web host
func HTTPJSON(args *TaskArgs) Result {
	result := NewResult(args.Task)
	req, client := CreateRequest("GET", args.Task.Params["url"].(string), nil)
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.Error = err
		} else {
			value := gjson.Get(string(contents), args.Task.Params["query"].(string)).String()
			l := len(value) == 0
			result.Warn = l || !(resp.StatusCode >= 200 && resp.StatusCode <= 299)
			if result.Warn {
				if l {
					result.Notification = "no value returned in JSON query!"
				} else {
					result.Notification = fmt.Sprintf("an invalid status code has been found: %d", resp.StatusCode)
				}
			}
			result.Update = struct {
				Content string `json:"content"`
			}{
				// TODO: Add error checking + multiple queries
				Content: value,
			}
		}
	}
	return result
}

// HTTPREGEXP lets you parse HTML with REGEXP
func HTTPREGEXP(args *TaskArgs) Result {
	result := NewResult(args.Task)
	req, client := CreateRequest("GET", args.Task.Params["url"].(string), nil)
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.Error = err
		} else {
			if regex, ok := args.Task.Params["regex"].(string); ok {
				content := string(contents)
				matcher, err := regexp.Compile(regex)
				if err != nil {
					result.Error = err
				} else {
					if !matcher.MatchString(content) {
						result.Error = fmt.Errorf("no match")
					} else {
						value := matcher.FindStringSubmatch(content)
						if len(value) == 1 {
							value = []string{"", value[0]}
						}
						l := len(value[1]) == 0
						result.Warn = l || !(resp.StatusCode >= 200 && resp.StatusCode <= 299)
						if result.Warn {
							if l {
								result.Notification = "no value returned in HTML query!"
							} else {
								result.Notification = fmt.Sprintf("an invalid status code has been found: %d", resp.StatusCode)
							}
						}
						result.Update = struct {
							Content string `json:"content"`
						}{
							// TODO: Add error checking + multiple queries
							Content: value[1],
						}
					}
				}
			}
		}
	}
	return result
}
