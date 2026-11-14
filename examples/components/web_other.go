//go:build !js

package main

import "github.com/dyike/keel/ui/core"

func setupGalleryNavigation(*gallery) {}
func galleryReady()                   {}

func switchGalleryBundle(string, string) {}

func galleryLayout(w core.Widget, gtx core.C) core.D { return w.Layout(gtx) }
