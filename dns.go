package tasks

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/tidwall/gjson"
	"pkg.goda.sh/utils"
)

// DNS checks if a domain resolves to anything
func DNS(args *TaskArgs) Result {
	result := NewResult(args.Task)
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"provider": "https://cloudflare-dns.com/dns-query",
		"target":   "example.org",
		"request":  "A",
	})
	req, client := CreateRequest("GET", fmt.Sprintf("%s?name=%s&type=%s", params.Get("provider").String(), params.Get("target").String(), params.Get("request").String()), nil)
	req.Header.Set("Accept", "application/dns-json")
	resp, err := client.Do(req)
	if err != nil {
		result.Error = err
	} else {
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			result.Error = err
		} else {
			status := gjson.Get(string(contents), "Status").String()
			result.Warn = status != "0"
			if result.Warn {
				result.Notification = fmt.Sprintf("invalid %s record has been detected! Status code: %s", params.Get("request").String(), status)
			}
			result.Update = struct {
				Valid bool `json:"valid"`
			}{
				Valid: !result.Warn,
			}
		}
	}
	if result.Error != nil {
		result.Notification = fmt.Sprintf("a DNS error has occurred: %q", result.Error)
	}
	return result
}

// DNSCIDR validates a dns address with CIDR ranges
func DNSCIDR(args *TaskArgs) Result {
	result := NewResult(args.Task)
	params := utils.ParamsParser(args.Task.Params, utils.DefaultParams{
		"provider": "https://cloudflare-dns.com/dns-query",
		"target":   "example.org",
		"request":  "A",
	})
	ranges := params.Get("ranges").Strings()
	if len(ranges) > 0 {
		req, client := CreateRequest("GET", fmt.Sprintf("%s?name=%s&type=%s", params.Get("provider").String(), params.Get("target").String(), params.Get("request").String()), nil)
		req.Header.Set("Accept", "application/dns-json")
		resp, err := client.Do(req)
		if err != nil {
			result.Error = err
		} else {
			defer resp.Body.Close()
			contents, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				result.Error = err
			} else {
				valid := false
			search:
				for _, cidr := range ranges {
					if _, ipn, err := net.ParseCIDR(cidr); err == nil {
						for _, addr := range gjson.Get(string(contents), "Answer.#.data").Array() {
							if valid = ipn.Contains(net.ParseIP(addr.String())); valid {
								break search
							}
						}
					}
				}
				result.Warn = !valid
				if result.Warn {
					result.Notification = fmt.Sprintf("%s record is not within the valid CIDR ranges!", params.Get("request").String())
				}
				result.Update = struct {
					Valid bool `json:"valid"`
				}{
					Valid: !result.Warn,
				}
			}
		}
	} else {
		result.Cancelled = args.Task.Cancel()
		result.Error = fmt.Errorf("missing ranges")
	}
	if result.Error != nil {
		result.Notification = fmt.Sprintf("a DNS-CIDR error has occurred: %q", result.Error)
	}
	return result
}
