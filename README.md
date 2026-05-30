# iLyricMP3

iLyricMP3 is a small command-line tool that embeds `.lrc` lyrics into an `.mp3` file so the lyrics can be displayed by iTunes, Apple Music, and synced iPods.

It writes lyrics directly into the MP3 file's ID3 tag. No new MP3 file is created.

## What It Does

- Finds a same-name `.lrc` and `.mp3` pair in the same folder.
- Removes LRC time tags such as `[00:12.34]`.
- Converts rich lyric JSON lines from sources such as NetEase Cloud Music into readable text.
- Writes the plain lyrics into the MP3 `USLT` lyrics frame.
- Uses UTF-16 lyrics encoding for better iTunes/iPod compatibility.
- Preserves existing ID3 metadata such as title, artist, album, and cover art.
- Creates a `.bak` backup before modifying the MP3.

## Limitations

iTunes and classic iPods generally display plain embedded lyrics, not real-time synchronized LRC lyrics. This tool therefore embeds the lyric text only. The timestamps are removed before writing.

Always keep your own backup of important music files before batch processing. iLyricMP3 modifies the original MP3 file in place and also creates a `.bak` backup next to the MP3.

## Download

Download the latest release from:

<https://github.com/ieshishinjin/iLyricMP3/releases>

Choose the file for your platform:

| Platform | File |
| --- | --- |
| macOS Apple Silicon | `iLyricMP3-macos-arm64` |
| macOS Intel | `iLyricMP3-macos-amd64` |
| Windows 64-bit | `iLyricMP3-windows-amd64.exe` |
| Linux 64-bit | `iLyricMP3-linux-amd64` |

## Usage

Put the `.mp3` and `.lrc` files in the same folder with the same base name:

```text
song.mp3
song.lrc
```

Then run iLyricMP3 with one or more file paths:

```bash
iLyricMP3 song.lrc
```

or:

```bash
iLyricMP3 song.mp3
```

Batch usage:

```bash
iLyricMP3 song1.mp3 song2.mp3 song3.mp3
```

On success, it prints:

```text
成功 | /path/to/song.mp3
```

## macOS Usage

On macOS, this is a command-line executable, not a `.app` bundle. The most reliable way to use it is through Terminal.

For Intel Mac:

```bash
chmod +x /path/to/iLyricMP3-macos-amd64
/path/to/iLyricMP3-macos-amd64 /path/to/song.lrc
```

For Apple Silicon Mac:

```bash
chmod +x /path/to/iLyricMP3-macos-arm64
/path/to/iLyricMP3-macos-arm64 /path/to/song.lrc
```

You can also type the executable path, add a space, then drag the `.mp3` or `.lrc` file from Finder into the Terminal window. Terminal will insert the file path automatically.

To process multiple songs, type the executable path, add a space, then drag several `.mp3` files into the Terminal window and press Enter. Each MP3 must have a same-name `.lrc` file in the same folder.

If macOS blocks the executable because it was downloaded from the internet, run:

```bash
xattr -d com.apple.quarantine /path/to/iLyricMP3-macos-amd64
```

Then run the command again.

## iTunes / Apple Music / iPod Notes

After embedding lyrics, import the MP3 into iTunes or Apple Music.

If the song was already in your library, iTunes/Music may keep cached metadata. Remove the song from the library and import the modified MP3 again.

For iPod syncing:

1. Embed the lyrics into the MP3 with iLyricMP3.
2. Re-import the modified MP3 into iTunes/Music if needed.
3. Sync the song to your iPod.
4. Open the song's lyrics view on the iPod.

## Build From Source

Requires Go 1.21 or later.

```bash
go build -o iLyricMP3 .
```

Cross-compile release binaries:

```bash
GOOS=darwin GOARCH=arm64 go build -o releases/iLyricMP3-macos-arm64 .
GOOS=darwin GOARCH=amd64 go build -o releases/iLyricMP3-macos-amd64 .
GOOS=windows GOARCH=amd64 go build -o releases/iLyricMP3-windows-amd64.exe .
GOOS=linux GOARCH=amd64 go build -o releases/iLyricMP3-linux-amd64 .
```

## License

MIT
