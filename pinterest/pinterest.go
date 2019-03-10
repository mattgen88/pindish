package pinterest

import (
	"errors"
	"net"
	"net/http"
	"time"
)

var ErrAPIOverLimit = errors.New("API over limit")

var netTransport = &http.Transport{
	Dial: (&net.Dialer{
		Timeout: 5 * time.Second,
	}).Dial,
	TLSHandshakeTimeout: 5 * time.Second,
}
var netClient = &http.Client{
	Timeout:   time.Second * 10,
	Transport: netTransport,
}
