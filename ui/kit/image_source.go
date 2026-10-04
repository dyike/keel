package kit

import (
	"context"
	"time"

	"github.com/dyike/keel/ui/core"
	"github.com/dyike/keel/ui/el"
)

// Source loads an HTTP(S) URL, file path/URL or data URL asynchronously. Same
// source is a no-op; Retry forces a refresh. Empty clears and cancels the load.
// Call from the UI thread or core.Update, like other component mutations.
func (v *ImageView) Source(source string) *ImageView {
	if source != "" && source == v.source {
		return v
	}
	v.stopLoad()
	v.source = source
	v.setPixels(nil)
	if source != "" {
		v.startLoad()
	}
	return v
}

// Cache replaces the shared default (64MiB estimated budget). Nil disables
// caching. Changing cache cancels and restarts the current source.
func (v *ImageView) Cache(cache *ImageCache) *ImageView {
	if v.cache == cache {
		return v
	}
	v.stopLoad()
	v.cache = cache
	if v.source != "" {
		v.setPixels(nil)
		v.startLoad()
	}
	return v
}
func (v *ImageView) Loading() bool                          { return v.loading }
func (v *ImageView) ImageError() error                      { return v.imageErr }
func (v *ImageView) LoadingContent(view el.View) *ImageView { v.loadingContent = view; return v }
func (v *ImageView) Fallback(view el.View) *ImageView       { v.fallback = view; return v }

// Retry refreshes the source, bypassing its cached result. Without Source it
// invokes OnRetry for an error state. Disabled components ignore user retries.
func (v *ImageView) Retry() {
	if v.disabled {
		return
	}
	if v.source != "" {
		v.stopLoad()
		v.cache.Delete(v.source)
		v.setPixels(nil)
		v.startLoad()
		return
	}
	if v.err != "" && v.retry != nil {
		v.err = ""
		v.retry()
	}
}
func (v *ImageView) stopLoad() {
	if v.cancel != nil {
		v.cancel()
		v.cancel = nil
	}
	v.revision++
	v.loading = false
	v.imageErr = nil
}
func (v *ImageView) startLoad() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	v.cancel, v.loading = cancel, true
	revision, source, cache := v.revision, v.source, v.cache
	go func() {
		defer cancel()
		img, err := cache.load(ctx, source)
		core.Update(func() {
			if v.revision != revision {
				return
			}
			v.loading, v.cancel, v.imageErr = false, nil, err
			if err != nil {
				v.err = err.Error()
			} else {
				v.setPixels(img)
			}
		})
	}()
}
