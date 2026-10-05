package core

import (
	"context"
	"errors"
	"io"
	"sync/atomic"
)

// Network images are opt-in too: net/http with TLS adds about 4 MB, so
// http and https image sources fail until an app links a fetcher in:
//
//	import _ "github.com/dyike/keel/ui/netimage"
//
// Local paths, file URLs and data URLs always work.

// ImageFetcher downloads a network image source. header holds request
// headers; it returns the status code, the response headers with canonical
// names ("Etag", "Last-Modified", "Cache-Control") and the body, which the
// caller closes.
type ImageFetcher func(ctx context.Context, url string, header map[string]string) (status int, respHeader map[string]string, body io.ReadCloser, err error)

// ErrNoImageFetcher is returned for http and https sources without a fetcher.
var ErrNoImageFetcher = errors.New(`network images are not linked in: import _ "github.com/dyike/keel/ui/netimage"`)

var imageFetcher atomic.Pointer[ImageFetcher]

// SetImageFetcher installs the network image fetcher; ui/netimage calls it
// when imported. Nil removes it.
func SetImageFetcher(f ImageFetcher) {
	if f == nil {
		imageFetcher.Store(nil)
		return
	}
	imageFetcher.Store(&f)
}

// FetchImage downloads url with the installed fetcher.
func FetchImage(ctx context.Context, url string, header map[string]string) (int, map[string]string, io.ReadCloser, error) {
	p := imageFetcher.Load()
	if p == nil {
		return 0, nil, nil, ErrNoImageFetcher
	}
	return (*p)(ctx, url, header)
}
