//go:build !js

package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"image"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dyike/keel/gio/gpu/headless"
	"github.com/dyike/keel/gio/op"
)

//go:embed film/soundtrack.m4a
var filmSoundtrack []byte

// renderMovie writes a private temporary file. Callers publish it only after
// the encoder has flushed the final frame and closed the MP4 container.
func renderMovie(ctx context.Context, dir string, c composition) (string, error) {
	encoder, err := exec.LookPath("ffmpeg")
	if err != nil {
		return "", fmt.Errorf("install FFmpeg with libx264 and add ffmpeg to PATH")
	}
	w, err := headless.NewWindow(videoWidth, videoHeight)
	if err != nil {
		return "", fmt.Errorf("open offscreen GPU: %w", err)
	}
	defer w.Release()
	f, err := os.CreateTemp(dir, "keel-"+templates[c.Scene].name+"-*.mp4")
	if err != nil {
		return "", err
	}
	path := f.Name()
	keep := false
	defer func() {
		if !keep {
			os.Remove(path)
		}
	}()
	if err := f.Close(); err != nil {
		return "", err
	}
	args := []string{"-hide_banner", "-loglevel", "error", "-y", "-f", "rawvideo", "-pixel_format", "rgba", "-video_size", fmt.Sprintf("%dx%d", videoWidth, videoHeight),
		"-framerate", fmt.Sprint(fps), "-i", "pipe:0"}
	if templates[c.Scene].name == filmName {
		audio, err := os.CreateTemp(dir, "keel-score-*.m4a")
		if err != nil {
			return "", err
		}
		defer os.Remove(audio.Name())
		_, writeErr := audio.Write(filmSoundtrack)
		closeErr := audio.Close()
		if writeErr != nil {
			return "", writeErr
		}
		if closeErr != nil {
			return "", closeErr
		}
		args = append(args, "-i", audio.Name(), "-map", "0:v:0", "-map", "1:a:0", "-c:a", "copy")
	}
	args = append(args, "-c:v", "libx264", "-crf", "18", "-pix_fmt", "yuv420p", "-movflags", "+faststart", path)
	cmd := exec.CommandContext(ctx, encoder, args...)
	var diagnostic bytes.Buffer
	cmd.Stderr = &diagnostic
	input, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	if err = cmd.Start(); err != nil {
		input.Close()
		return "", err
	}
	var ops op.Ops
	frame := image.NewRGBA(image.Rect(0, 0, videoWidth, videoHeight))
	for n := range c.frames() {
		if err = ctx.Err(); err != nil {
			break
		}
		ops.Reset()
		c.draw(&ops, frame.Rect.Size(), n)
		if err = w.Frame(&ops); err != nil {
			break
		}
		if err = w.Screenshot(frame); err != nil {
			break
		}
		if _, err = input.Write(frame.Pix); err != nil {
			break
		}
	}
	closeErr := input.Close()
	waitErr := cmd.Wait()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if waitErr != nil {
		return "", fmt.Errorf("FFmpeg: %w %s", waitErr, diagnostic.String())
	}
	if err != nil {
		return "", err
	}
	if closeErr != nil {
		return "", closeErr
	}
	keep = true
	return path, nil
}

func saveMovie(c composition) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	path, err := renderMovie(context.Background(), dir, c)
	if err != nil {
		return "", err
	}
	return "Saved " + filepath.Base(path) + " in " + dir, nil
}
