package tools

import (
	"fmt"
	"strings"
)

const (
	networkSplit = "@"
)

func ParseNetwork(str string) (network, addr string, err error) {
	if idx := strings.Index(str, networkSplit); idx == -1 {
		err = fmt.Errorf("addr: \"%s\" error, must be network@address:port or network@unixsocket, network is like 'tcp' and so on", str)
		return
	} else {
		network = str[:idx]
		addr = str[idx+1:]
		return
	}
}
