# iLyricMP3 v1.0.0

First public release.

## Features

- Embed `.lrc` lyrics into same-name `.mp3` files.
- Write lyrics into the MP3 `USLT` ID3 frame.
- Save lyrics as UTF-16 for better iTunes and iPod compatibility.
- Preserve existing ID3 metadata frames, including title, artist, album, and cover art.
- Create a `.bak` backup before modifying the MP3.
- Support macOS Apple Silicon, macOS Intel, Windows 64-bit, and Linux 64-bit builds.

## Usage

Place the files in the same folder:

```text
song.mp3
song.lrc
```

Run:

```bash
iLyricMP3 song.lrc
```

or:

```bash
iLyricMP3 song.mp3
```

The MP3 file is modified in place.

## Downloads

- `iLyricMP3-macos-arm64`
- `iLyricMP3-macos-amd64`
- `iLyricMP3-windows-amd64.exe`
- `iLyricMP3-linux-amd64`

## Notes

iTunes and classic iPods normally show embedded plain lyrics rather than synchronized LRC timing. iLyricMP3 removes LRC time tags and writes the remaining lyric text.
