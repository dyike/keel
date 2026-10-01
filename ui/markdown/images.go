package markdown

import widgets "github.com/dyike/keel/ui/widget"

// ImageLoader sets how local or remote image sources are read. Configure it
// before the first Render. A nil loader uses widget.DecodeImage.
func (d *Doc) ImageLoader(loader widgets.ImageLoader) *Doc { d.imageLoader = loader; return d }

func (d *Doc) bindImages(r *richBlock, spans []span) {
	for i, s := range spans {
		if s.imageURL == "" || r.runs[i].image != nil {
			continue
		}
		if d.images == nil {
			d.images = make(map[string]*widgets.ImageAsset)
		}
		asset := d.images[s.imageURL]
		if asset == nil {
			asset = widgets.LoadImage(s.imageURL, d.imageLoader)
			d.images[s.imageURL] = asset
		}
		r.runs[i].image = &widgets.ImageView{Asset: asset, Alt: s.imageAlt}
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
