package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	tld "github.com/jpillora/go-tld"
)

func GetHostname(l string) string {
	if len(l) == 0 {
		fmt.Println("empty url")
		return ""
	}

	u, err := tld.Parse(l)
	if err != nil {
		fmt.Printf("illegal url %s, err:%s\n", l, err.Error())
		return ""
	}
	registeredDomain := u.Domain + "." + u.TLD
	return registeredDomain
}

func NormalizeUrl(linkUrl *url.URL) string {
	queryParams := linkUrl.Query()
	pageRegex, err := regexp.Compile("(?i)^page")
	if err != nil {
		fmt.Println("Error in compiling the regex")
		return ""
	}

	// Removing the page params that
	for key, _ := range queryParams {
		if !pageRegex.MatchString(key) {
			queryParams.Del(key)
		}
	}
	linkUrl.RawQuery = queryParams.Encode()
	return strings.Replace(strings.TrimSuffix(linkUrl.String(), "/"), "http://", "https://", 1)
}
