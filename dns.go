package tasks

import (
	"fmt"
	"io/ioutil"
	"net"

	"github.com/tidwall/gjson"
)

// DNS checks if a domain resolves to anything
func DNS(args *TaskArgs) Result {
	result := NewResult(args.Task)
	provider := "https://cloudflare-dns.com/dns-query"
	if args.Task.Params["provider"] != nil {
		provider = args.Task.Params["provider"].(string)
	}
	target := "example.org"
	if args.Task.Params["target"] != nil {
		target = args.Task.Params["target"].(string)
	}
	request := "A"
	if args.Task.Params["type"] != nil {
		request = args.Task.Params["type"].(string)
	}
	req, client := CreateRequest("GET", fmt.Sprintf("%s?name=%s&type=%s", provider, target, request), nil)
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
				result.Notification = fmt.Sprintf("invalid %s record has been detected! Status code: %s", request, status)
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
	provider := "https://cloudflare-dns.com/dns-query"
	if args.Task.Params["provider"] != nil {
		provider = args.Task.Params["provider"].(string)
	}
	target := "example.org"
	if args.Task.Params["target"] != nil {
		target = args.Task.Params["target"].(string)
	}
	request := "A"
	if args.Task.Params["type"] != nil {
		request = args.Task.Params["type"].(string)
	}
	ranges := []string{}
	if args.Task.Params["ranges"] != nil {
		for _, cidr := range args.Task.Params["ranges"].([]interface{}) {
			ranges = append(ranges, cidr.(string))
		}
	}
	if len(ranges) > 0 {
		req, client := CreateRequest("GET", fmt.Sprintf("%s?name=%s&type=%s", provider, target, request), nil)
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
				for _, addr := range gjson.Get(string(contents), "Answer.#.data").Array() {
					if valid {
						break
					}
					addr := fmt.Sprintf("%s/32", addr.String()) // should always be a single IP address
					for _, cidr := range ranges {
						_, n, err := net.ParseCIDR(cidr)
						if err != nil {
							break
						}
						ip, _, err := net.ParseCIDR(addr)
						if err != nil {
							break
						}
						if valid = n.Contains(ip); valid {
							break
						}
					}
				}
				result.Warn = !valid
				if result.Warn {
					result.Notification = fmt.Sprintf("%s record is not within the valid CIDR ranges!", request)
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
