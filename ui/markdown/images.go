package markdown

import (
	"context"
	"image"

	"github.com/dyike/keel/ui/internal/imageload"
)

// ImageLoader reads an image source off the UI thread.
type ImageLoader func(context.Context, string) (image.Image, error)

// DecodeImage is the default loader: local paths, file, HTTP(S) and data
// URLs; PNG, JPEG, GIF (first frame) and WebP; 16 MiB and 32 megapixels at most.
func DecodeImage(ctx context.Context, source string) (image.Image, error) {
	return imageload.Decode(ctx, source)
}

// ImageLoader sets how local or remote image sources are read. Configure it
// before the first Render. A nil loader uses DecodeImage.
func (d *Doc) ImageLoader(loader ImageLoader) *Doc {
	d.imageLoader = imageload.Loader(loader)
	return d
}

func (d *Doc) bindImages(r *richBlock, spans []span) {
	for i, s := range spans {
		if s.imageURL == "" || r.runs[i].image != nil {
			continue
		}
		if d.images == nil {
			d.images = make(map[string]*imageload.Asset)
		}
		asset := d.images[s.imageURL]
		if asset == nil {
			asset = imageload.Load(s.imageURL, d.imageLoader)
			d.images[s.imageURL] = asset
		}
		r.runs[i].image = &imageload.View{Asset: asset, Alt: s.imageAlt}
	}
}
func (r *richBlock) imagesRevision() uint64 {
	var revision uint64
	for _, rn := range r.runs {
		if rn.image != nil {
			revision += rn.image.Asset.Revision()
		}
	}
	return revision
}
