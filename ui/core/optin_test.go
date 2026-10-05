package core

import (
	"context"
	"errors"
	"testing"
)

// Without ui/netimage and ui/highlight linked in, network images fail with
// a message naming the package to import, and nothing highlights.
func TestOptInFeaturesAreOffByDefault(t *testing.T) {
	if _, err := ReadImageSource(context.Background(), "https://example.com/a.png"); !errors.Is(err, ErrNoImageFetcher) {
		t.Fatalf("network image without a fetcher: %v", err)
	}
	if _, err := ReadImageSource(context.Background(), "data:image/png;base64,iVBORw0KGgo="); errors.Is(err, ErrNoImageFetcher) {
		t.Fatal("data URLs need no fetcher")
	}
	if CurrentHighlighter() != nil {
		t.Fatal("no highlighter is installed by default")
	}
}
