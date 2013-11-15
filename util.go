package statsd

import (
	"net/url"
	"strings"
)

func parseUrl(statsdUrl string) (host, prefix string, err error) {
	parsedStatsdUrl, err := url.Parse(statsdUrl)
	if err != nil {
		return "", "", err
	}
	prefix = strings.TrimPrefix(parsedStatsdUrl.Path, "/")
	if len(prefix) > 0 {
		prefix = prefix + "."
	}
	return parsedStatsdUrl.Host, prefix, nil
}
