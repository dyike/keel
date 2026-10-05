// Package netimage lets Keel load images from http and https URLs. Import
// it for its side effect where the app shows network images:
//
//	import _ "github.com/dyike/keel/ui/netimage"
//
// It is separate because net/http and TLS add about 4 MB to a binary that
// may never fetch an image.
package netimage

import (
	"context"
	"io"
	"net/http"

	"github.com/dyike/keel/ui/core"
)

func init() { core.SetImageFetcher(Fetch) }

// Client is the HTTP client used for images, http.DefaultClient unless set
// before the first request.
var Client = http.DefaultClient

// Fetch is the core.ImageFetcher this package installs.
func Fetch(ctx context.Context, url string, header map[string]string) (int, map[string]string, io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	for k, v := range header {
		req.Header.Set(k, v)
	}
	resp, err := Client.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	h := make(map[string]string, len(resp.Header))
	for k := range resp.Header {
		h[http.CanonicalHeaderKey(k)] = resp.Header.Get(k)
	}
	return resp.StatusCode, h, resp.Body, nil
}
