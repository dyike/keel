package kit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/dyike/keel/ui/core"
)

// maxDiskImageBytes caps one download, matching the decoder's limit.
const maxDiskImageBytes = 16 << 20

// imageDiskCache keeps HTTP(S) image bytes on disk between runs. Each source
// is a body file plus a small JSON record of its validators. Within ttl a
// stored copy is used as is; after it, the copy is revalidated with
// If-None-Match / If-Modified-Since, and a 304 keeps it. When the network
// fails, a stale copy is still used. Total size stays under max by evicting
// the least recently used files.
type imageDiskCache struct {
	mu  sync.Mutex
	dir string
	max int64
	ttl time.Duration
	now func() time.Time
}

type imageDiskRecord struct {
	Source       string    `json:"source"`
	ETag         string    `json:"etag,omitempty"`
	LastModified string    `json:"last_modified,omitempty"`
	Checked      time.Time `json:"checked"`
}

// Disk also stores HTTP(S) downloads under dir, up to maxBytes in total, so
// they survive restarts. A copy younger than ttl is used without a request;
// an older one is revalidated with its ETag or Last-Modified. Data URLs and
// local files are never copied. An empty dir turns the disk cache off.
func (c *ImageCache) Disk(dir string, maxBytes int64, ttl time.Duration) *ImageCache {
	c.mu.Lock()
	defer c.mu.Unlock()
	if dir == "" || maxBytes <= 0 {
		c.disk = nil
		return c
	}
	c.disk = &imageDiskCache{dir: dir, max: maxBytes, ttl: max(0, ttl), now: time.Now}
	return c
}

// read decodes a source, through the disk cache when it applies.
func (c *ImageCache) read(ctx context.Context, source string) (*imageMedia, error) {
	c.mu.Lock()
	disk := c.disk
	c.mu.Unlock()
	if disk == nil || !(strings.HasPrefix(source, "https://") || strings.HasPrefix(source, "http://")) {
		return loadImageMedia(ctx, source)
	}
	data, err := disk.get(ctx, source)
	if err != nil {
		return nil, err
	}
	m, err := decodeImageMedia(data)
	if err != nil {
		disk.remove(source) // a corrupt copy must not stick
	}
	return m, err
}

func (d *imageDiskCache) paths(source string) (body, record string) {
	sum := sha256.Sum256([]byte(source))
	name := filepath.Join(d.dir, hex.EncodeToString(sum[:16]))
	return name + ".img", name + ".json"
}

func (d *imageDiskCache) get(ctx context.Context, source string) ([]byte, error) {
	bodyPath, recordPath := d.paths(source)
	var rec imageDiskRecord
	var cached []byte
	if raw, err := os.ReadFile(recordPath); err == nil && json.Unmarshal(raw, &rec) == nil && rec.Source == source {
		if data, err := os.ReadFile(bodyPath); err == nil {
			cached = data
		}
	}
	now := d.now()
	if cached != nil && now.Sub(rec.Checked) < d.ttl {
		d.touch(bodyPath, now)
		return cached, nil
	}
	header := map[string]string{}
	if cached != nil {
		if rec.ETag != "" {
			header["If-None-Match"] = rec.ETag
		}
		if rec.LastModified != "" {
			header["If-Modified-Since"] = rec.LastModified
		}
	}
	status, respHeader, body, err := core.FetchImage(ctx, source, header)
	if err != nil {
		if cached != nil && ctx.Err() == nil {
			return cached, nil // offline: a stale copy beats nothing
		}
		return nil, err
	}
	defer body.Close()
	switch {
	case status == 304 && cached != nil: // Not Modified
		rec.Checked = now
		d.save(source, bodyPath, recordPath, nil, rec)
		return cached, nil
	case status < 200 || status >= 300:
		if cached != nil && status >= 500 {
			return cached, nil
		}
		return nil, fmt.Errorf("image HTTP status %d", status)
	}
	data, err := io.ReadAll(io.LimitReader(body, maxDiskImageBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxDiskImageBytes {
		return nil, fmt.Errorf("image exceeds %d bytes", maxDiskImageBytes)
	}
	if strings.Contains(respHeader["Cache-Control"], "no-store") {
		return data, nil
	}
	d.save(source, bodyPath, recordPath, data, imageDiskRecord{Source: source, ETag: respHeader["Etag"], LastModified: respHeader["Last-Modified"], Checked: now})
	return data, nil
}

// save writes the record, and the body when data is not nil, then evicts.
// Failures only cost a later download.
func (d *imageDiskCache) save(source, bodyPath, recordPath string, data []byte, rec imageDiskRecord) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if os.MkdirAll(d.dir, 0o755) != nil {
		return
	}
	if data != nil {
		if int64(len(data)) > d.max || writeFileAtomic(bodyPath, data) != nil {
			return
		}
	}
	raw, _ := json.Marshal(rec)
	if writeFileAtomic(recordPath, raw) != nil {
		return
	}
	d.touch(bodyPath, d.now())
	d.evict(bodyPath)
}

func (d *imageDiskCache) remove(source string) {
	bodyPath, recordPath := d.paths(source)
	d.mu.Lock()
	defer d.mu.Unlock()
	os.Remove(bodyPath)
	os.Remove(recordPath)
}

// touch marks a body as used; its modification time orders eviction.
func (d *imageDiskCache) touch(path string, at time.Time) { os.Chtimes(path, at, at) }

// evict removes the least recently used bodies until the total fits, never
// the one just written.
func (d *imageDiskCache) evict(keep string) {
	entries, err := os.ReadDir(d.dir)
	if err != nil {
		return
	}
	type body struct {
		path string
		size int64
		used time.Time
	}
	var bodies []body
	var total int64
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".img") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		bodies = append(bodies, body{filepath.Join(d.dir, e.Name()), info.Size(), info.ModTime()})
		total += info.Size()
	}
	slices.SortFunc(bodies, func(a, b body) int { return a.used.Compare(b.used) })
	for _, b := range bodies {
		if total <= d.max {
			break
		}
		if b.path == keep {
			continue
		}
		if os.Remove(b.path) == nil {
			os.Remove(strings.TrimSuffix(b.path, ".img") + ".json")
			total -= b.size
		}
	}
}

func writeFileAtomic(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}
