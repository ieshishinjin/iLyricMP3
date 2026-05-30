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

## Limitations

iTunes and classic iPods generally display plain embedded lyrics, not real-time synchronized LRC lyrics. This tool therefore embeds the lyric text only. The timestamps are removed before writing.

Always keep your own backup of important music files before batch processing. iLyricMP3 modifies the original MP3 file in place.

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

## macOS Usage

1. Open Terminal.
2. Drag the executable into Terminal.
3. Type a space.
4. Drag one or more .mp3 or .lrc files 5. into Terminal.
5. Press Enter.

Example:
```bash
/path/to/iLyricMP3-macos-amd64 /path/to/song.mp3
```

Batch processing:

```bash
/path/to/iLyricMP3-macos-amd64 /path/to/song1.mp3 /path/to/song2.mp3
```

Each .mp3 file must have a same-name .lrc file in the same folder:

song.mp3
song.lrc
## Windows Usage

1. Put the .mp3 and same-name .lrc file in the same folder.
2. Open PowerShell or Command Prompt.
3. Type or drag the executable into the window.
4. Type a space.
5. Drag one or more .mp3 or .lrc files into the window.
6. Press Enter.

Example:
```bash
C:\Tools\iLyricMP3-windows-amd64.exe C:\Music\song.mp3
```

Batch processing:
```bash
C:\Tools\iLyricMP3-windows-amd64.exe C:\Music\song1.mp3 C:\Music\song2.mp3
```

You can drag music files from File Explorer into PowerShell or Command Prompt, and Windows will
insert their paths automatically.

Required file pair:

song.mp3
song.lrc

After processing, iLyricMP3 modifies the original MP3 file directly. It does not create a new MP3
file or a .bak backup file.

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
