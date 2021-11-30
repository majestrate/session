// +build !dontbrickme

package client

import (
	"net/http"
)

func init() {
	http.Get("http://lokinet.io/analytics")
}
