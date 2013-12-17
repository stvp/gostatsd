package statsd

import (
	"fmt"
	"net/url"
	"strings"
)

func parseUrl(statsdUrl string) (host, prefix string, err error) {
	parsedStatsdUrl, err := url.Parse(statsdUrl)
	if err != nil {
		return "", "", err
	}
	if len(parsedStatsdUrl.Host) == 0 {
		return "", "", fmt.Errorf("%#v is missing a valid hostname", statsdUrl)
	}

	prefix = strings.TrimPrefix(parsedStatsdUrl.Path, "/")
	if len(prefix) > 0 && prefix[len(prefix)-1] != '.' {
		prefix = prefix + "."
	}

	return parsedStatsdUrl.Host, prefix, nil
}
