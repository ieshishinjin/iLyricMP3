package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeLyricsExtractsRichLyricText(t *testing.T) {
	input := `[00:00.00]{"t":0,"c":[{"tx":"作词: "},{"tx":"陶喆","li":"http://example.com","or":"orpheus://artist"}]}
[00:00.83]{"t":830,"c":[{"tx":"作曲: "},{"tx":"陶喆"}]}
[00:01.20]David Tao
`

	got := normalizeLyrics(input)
	want := "作词: 陶喆\n作曲: 陶喆\nDavid Tao"
	if got != want {
		t.Fatalf("normalizeLyrics() = %q, want %q", got, want)
	}
}

func TestProcessAllHandlesMultipleInputsWithoutWaiting(t *testing.T) {
	dir := t.TempDir()
	firstMP3 := filepath.Join(dir, "first.mp3")
	firstLRC := filepath.Join(dir, "first.lrc")
	secondMP3 := filepath.Join(dir, "second.mp3")
	secondLRC := filepath.Join(dir, "second.lrc")

	writeTestFile(t, firstMP3, []byte{0xFF, 0xFB, 0x90, 0x64})
	writeTestFile(t, firstLRC, []byte("[00:00.00]first lyric"))
	writeTestFile(t, secondMP3, []byte{0xFF, 0xFB, 0x90, 0x64})
	writeTestFile(t, secondLRC, []byte("[00:00.00]second lyric"))

	if failed := processAll([]string{firstMP3, firstLRC, secondMP3}); failed != 0 {
		t.Fatalf("processAll failed count = %d, want 0", failed)
	}

	assertFileContains(t, firstMP3, encodeUTF16LE("first lyric"))
	assertFileContains(t, secondMP3, encodeUTF16LE("second lyric"))
}

func TestEmbedLyricsPreservesID3v23Frames(t *testing.T) {
	title := buildFrameV23("TIT2", []byte{0x03, 'S', 'o', 'n', 'g'})
	artist := buildFrameV23("TPE1", []byte{0x03, 'A', 'r', 't', 'i', 's', 't'})
	cover := buildFrameV23("APIC", []byte{0x03, 'i', 'm', 'a', 'g', 'e', '/', 'j', 'p', 'e', 'g', 0x00, 0x03, 0x00, 0xFF, 0xD8, 0xFF})
	oldLyrics := buildFrameV23("USLT", buildLyricsPayload("old lyrics"))

	body := bytes.Join([][]byte{title, artist, cover, oldLyrics}, nil)
	body = append(body, make([]byte, 16)...)

	input := append(buildID3Header(3, 0, 0, len(body)), body...)
	input = append(input, []byte{0xFF, 0xFB, 0x90, 0x64}...)

	output, err := embedLyrics(input, "new lyrics")
	if err != nil {
		t.Fatal(err)
	}

	tag, ok, err := readID3Tag(output)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ID3 tag")
	}

	assertFrameCount(t, tag.frames, "TIT2", 1)
	assertFrameCount(t, tag.frames, "TPE1", 1)
	assertFrameCount(t, tag.frames, "APIC", 1)
	assertFrameCount(t, tag.frames, "USLT", 1)

	lyricsFrame := findFrame(tag.frames, "USLT")
	if lyricsFrame == nil {
		t.Fatal("expected USLT frame")
	}
	if !bytes.Contains(lyricsFrame, encodeUTF16LE("new lyrics")) {
		t.Fatal("expected new lyrics in output")
	}
	if bytes.Contains(lyricsFrame, encodeUTF16LE("old lyrics")) {
		t.Fatal("old lyrics should be replaced")
	}
}

func TestEmbedLyricsPreservesID3v22Frames(t *testing.T) {
	title := buildFrameV22("TT2", []byte{0x00, 'S', 'o', 'n', 'g'})
	artist := buildFrameV22("TP1", []byte{0x00, 'A', 'r', 't', 'i', 's', 't'})
	cover := buildFrameV22("PIC", []byte{0x00, 'J', 'P', 'G', 0x03, 0x00, 0xFF, 0xD8, 0xFF})

	body := bytes.Join([][]byte{title, artist, cover}, nil)
	body = append(body, make([]byte, 16)...)

	input := append(buildID3Header(2, 0, 0, len(body)), body...)
	input = append(input, []byte{0xFF, 0xFB, 0x90, 0x64}...)

	output, err := embedLyrics(input, "hello")
	if err != nil {
		t.Fatal(err)
	}

	tag, ok, err := readID3Tag(output)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected ID3 tag")
	}

	assertFrameCount(t, tag.frames, "TT2", 1)
	assertFrameCount(t, tag.frames, "TP1", 1)
	assertFrameCount(t, tag.frames, "PIC", 1)
	assertFrameCount(t, tag.frames, "ULT", 1)
}

func assertFrameCount(t *testing.T, frames [][]byte, id string, want int) {
	t.Helper()

	got := 0
	for _, frame := range frames {
		if frameMatchesID(frame, id) {
			got++
		}
	}
	if got != want {
		t.Fatalf("frame %s count = %d, want %d", id, got, want)
	}
}

func findFrame(frames [][]byte, id string) []byte {
	for _, frame := range frames {
		if frameMatchesID(frame, id) {
			return frame
		}
	}
	return nil
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func assertFileContains(t *testing.T, path string, want []byte) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, want) {
		t.Fatalf("%s does not contain expected bytes", path)
	}
}
