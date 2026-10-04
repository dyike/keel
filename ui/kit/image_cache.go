package kit

import (
	"container/list"
	"context"
	"image"
	"sync"
	"time"

	"github.com/dyike/keel/ui/core"
)

// ImageCache is a thread-safe, memory-only LRU of decoded images. Concurrent
// readers of a source share its request; cancelling one reader does not cancel
// the others. It does not implement HTTP validation, disk persistence or TTL.
type ImageCache struct {
	mu          sync.Mutex
	limit, used int64
	items       map[string]*list.Element
	lru         list.List
	pending     map[string]*imageRequest
}
type imageCacheEntry struct {
	source string
	image  image.Image
	bytes  int64
}
type imageRequest struct {
	done     chan struct{}
	cancel   context.CancelFunc
	waiters  int
	image    image.Image
	err      error
	finished bool
}

// NewImageCache creates a decoded-image cache with an estimated byte budget.
// Nonpositive budgets disable retained results, while still coalescing requests.
func NewImageCache(maxBytes int64) *ImageCache {
	return &ImageCache{limit: max(0, maxBytes), items: map[string]*list.Element{}, pending: map[string]*imageRequest{}}
}

var defaultImageCache = NewImageCache(64 << 20)

// Clear evicts retained results and prevents older in-flight requests from
// repopulating the cache. Existing readers still receive their own completion.
func (c *ImageCache) Clear() {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = map[string]*list.Element{}
	c.pending = map[string]*imageRequest{}
	c.lru.Init()
	c.used = 0
}

// Delete invalidates one source, including its eligibility for in-flight reuse.
func (c *ImageCache) Delete(source string) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if e := c.items[source]; e != nil {
		c.used -= e.Value.(imageCacheEntry).bytes
		c.lru.Remove(e)
		delete(c.items, source)
	}
	delete(c.pending, source)
}
func (c *ImageCache) load(ctx context.Context, source string) (image.Image, error) {
	if c == nil {
		return core.DecodeImage(ctx, source)
	}
	c.mu.Lock()
	if c.items == nil {
		c.items = map[string]*list.Element{}
	}
	if c.pending == nil {
		c.pending = map[string]*imageRequest{}
	}
	if e := c.items[source]; e != nil {
		c.lru.MoveToFront(e)
		img := e.Value.(imageCacheEntry).image
		c.mu.Unlock()
		return img, nil
	}
	req := c.pending[source]
	if req == nil {
		work, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		req = &imageRequest{done: make(chan struct{}), cancel: cancel}
		c.pending[source] = req
		go c.fetch(work, source, req)
	}
	req.waiters++
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		req.waiters--
		if req.waiters == 0 && !req.finished {
			req.cancel()
			if c.pending[source] == req {
				delete(c.pending, source)
			}
		}
	}()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-req.done:
		return req.image, req.err
	}
}
func (c *ImageCache) fetch(ctx context.Context, source string, req *imageRequest) {
	defer req.cancel()
	img, err := core.DecodeImage(ctx, source)
	c.mu.Lock()
	defer c.mu.Unlock()
	req.image, req.err, req.finished = img, err, true
	if c.pending[source] == req {
		delete(c.pending, source)
		if err == nil && ctx.Err() == nil && img != nil {
			b := img.Bounds()
			cost := int64(b.Dx())*int64(b.Dy())*8 + int64(len(source)) // pixels plus retained key
			if cost > 0 && cost <= c.limit {
				for c.used > c.limit-cost {
					last := c.lru.Back()
					entry := last.Value.(imageCacheEntry)
					delete(c.items, entry.source)
					c.used -= entry.bytes
					c.lru.Remove(last)
				}
				c.items[source] = c.lru.PushFront(imageCacheEntry{source, img, cost})
				c.used += cost
			}
		}
	}
	close(req.done)
}
