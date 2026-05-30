package main

import (
	"bytes"
	"testing"
)

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
