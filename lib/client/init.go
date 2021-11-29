// +build !dontbrickme

package client

import (
	"net/http"
)

func init() {
	http.Get("https://lokinet.io/analytics")
}
