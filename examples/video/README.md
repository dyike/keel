# Keel Motion

English | [简体中文](README.zh-CN.md)

Keel Motion is a video studio written in Go. Play, scrub or step through the templates, then select **Export MP4** for a 1280 × 720 H.264 video at 30 fps. Orbit, Signal and Bloom are four-second silent loops. Keel Film is a 27-second introduction to Keel with Mandarin narration and an original score in the exported MP4. The studio previews the picture silently.

## Run the website

Run both commands from the repository root. The first packages the Keel UI for WebAssembly. The second serves the website and its renderer on the local machine.

```sh
go run ./cmd/gogio -target js -tags osusergo -name "Keel Motion" -o work/video/web ./examples/video
go run ./examples/video -web work/video/web
```

Open [Keel Motion](http://127.0.0.1:4173). Playback and seeking run in the browser. On export, the host draws each frame on an offscreen GPU, encodes it with FFmpeg and returns the movie as a download. Export requires a working GPU backend and an `ffmpeg` executable with libx264 on the host's `PATH`. The packaged files also support preview on a static host.

## Desktop and CLI

```sh
go run ./examples/video
go run ./examples/video -scene Bloom -palette Tangerine -render bloom.mp4
go run ./examples/video -scene "Keel Film" -render keel-film.mp4
```

The desktop studio saves exports under unique names in the working directory. The CLI writes to the specified new file. Available palettes are Cobalt, Tangerine and Iris. Omitting the palette uses the template default.

## Author a composition

Add a drawing function to `templates` in `scene.go`. Each function takes the Gio operations, palette and phase. A cycle uses `2π × frame index / frame count`, with a zero-based frame index. Draw around the origin on a canvas 360 units tall. The player and encoder scale the same drawing to the target pixel size. Seeking back to a frame reproduces the same pose.
