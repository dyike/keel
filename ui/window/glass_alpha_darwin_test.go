//go:build darwin && !ios && !nometal

package window

import (
	"image"
	"image/color"
	"runtime"
	"testing"
	"unsafe"

	"github.com/dyike/keel/gio/gpu"
	"github.com/dyike/keel/gio/op"
	"github.com/dyike/keel/gio/op/clip"
	"github.com/dyike/keel/gio/op/paint"
	"github.com/dyike/keel/ui/internal/appkit"
	"github.com/ebitengine/purego"
)

type glassTextureRegion struct{ X, Y, Z, Width, Height, Depth uintptr }

// Core Animation must receive linear premultiplied pixels. Encoding RGB to
// sRGB before CA divides by alpha makes translucent highlights turn white.
func TestGlassMetalPremultipliedPixels(t *testing.T) {
	appkit.Pool(func() {
		device := appkit.ID(appkit.Call(appkit.Sym("MTLCreateSystemDefaultDevice")))
		if device == 0 {
			t.Skip("Metal unavailable")
		}
		defer appkit.Release(device)
		queue := appkit.Send(device, "newCommandQueue")
		defer appkit.Release(queue)
		desc := appkit.Send(appkit.Class("MTLTextureDescriptor"), "texture2DDescriptorWithPixelFormat:width:height:mipmapped:", mtlPixelFormatRGBA16Float, 8, 8, 0)
		appkit.Send(desc, "setStorageMode:", 0) // shared CPU/GPU storage
		appkit.Send(desc, "setUsage:", 5)       // shader read + render target
		texture := appkit.Send(device, "newTextureWithDescriptor:", uintptr(desc))
		if texture == 0 {
			t.Skip("shared Metal texture unavailable")
		}
		defer appkit.Release(texture)
		renderer, err := gpu.New(gpu.Metal{Device: uintptr(device), Queue: uintptr(queue), PixelFormat: mtlPixelFormatRGBA16Float})
		if err != nil {
			t.Fatal(err)
		}
		defer renderer.Release()
		var ops op.Ops
		paint.FillShape(&ops, color.NRGBA{R: 255, A: 128}, clip.Rect(image.Rect(0, 0, 4, 8)).Op())
		renderer.Clear(color.NRGBA{})
		if err := renderer.Frame(&ops, gpu.MetalRenderTarget{Texture: uintptr(texture)}, image.Pt(8, 8)); err != nil {
			t.Fatal(err)
		}
		buffer := appkit.Send(queue, "commandBuffer")
		appkit.Send(buffer, "commit")
		appkit.Send(buffer, "waitUntilCompleted")
		var pixels [8 * 8 * 8]byte
		// The reflected native call takes a uintptr, so pin the destination
		// before converting it; stack growth must not move the readback buffer.
		var pinned runtime.Pinner
		pinned.Pin(&pixels[0])
		defer pinned.Unpin()
		var read func(appkit.ID, appkit.SEL, uintptr, uintptr, glassTextureRegion, uintptr)
		purego.RegisterFunc(&read, appkit.Sym("objc_msgSend"))
		read(texture, appkit.Sel("getBytes:bytesPerRow:fromRegion:mipmapLevel:"), uintptr(unsafe.Pointer(&pixels[0])), 8*8, glassTextureRegion{Width: 8, Height: 8, Depth: 1}, 0)
		// float16(128/255) is 0x3804; RGB and alpha must contain the same
		// linear premultiplied value, with no sRGB encoding between them.
		if got := [8]byte(pixels[:8]); got != ([8]byte{4, 56, 0, 0, 0, 0, 4, 56}) {
			t.Fatalf("premultiplied RGBA16Float = %v", got)
		}
		if got := [8]byte(pixels[7*8 : 8*8]); got != ([8]byte{}) {
			t.Fatalf("clear RGBA16Float = %v", got)
		}
	})
}
