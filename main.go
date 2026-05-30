package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf16"
)

var timeTagPattern = regexp.MustCompile(`\[\d{1,2}:\d{2}(?:\.\d{2,3})?\]`)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("请将一个或多个 .lrc / .mp3 文件拖拽到此程序上")
		return
	}

	if failed := processAll(os.Args[1:]); failed > 0 {
		os.Exit(1)
	}
}

func processAll(inputPaths []string) int {
	seen := make(map[string]bool)
	succeeded := 0
	failed := 0

	for _, inputPath := range inputPaths {
		lrcPath, mp3Path, err := resolveInput(inputPath)
		if err != nil {
			fmt.Println(err.Error())
			failed++
			continue
		}

		key, err := filepath.Abs(mp3Path)
		if err != nil {
			key = mp3Path
		}
		if seen[key] {
			fmt.Printf("跳过 | %s | 已处理同名音频\n", inputPath)
			continue
		}
		seen[key] = true

		if err := processPair(lrcPath, mp3Path); err != nil {
			fmt.Println(err.Error())
			failed++
			continue
		}
		succeeded++
	}

	fmt.Printf("完成 | 成功 %d | 失败 %d\n", succeeded, failed)
	return failed
}

func resolveInput(inputPath string) (string, string, error) {
	ext := strings.ToLower(filepath.Ext(inputPath))

	var lrcPath, mp3Path string
	switch ext {
	case ".lrc":
		lrcPath = inputPath
		mp3Path = replaceExt(inputPath, ".mp3")
	case ".mp3":
		mp3Path = inputPath
		lrcPath = replaceExt(inputPath, ".lrc")
	default:
		return "", "", fmt.Errorf("失败 | %s | 请拖拽 .lrc 或 .mp3 文件", inputPath)
	}

	return lrcPath, mp3Path, nil
}

func processPair(lrcPath string, mp3Path string) error {
	if !fileExists(mp3Path) {
		return fmt.Errorf("失败 | %s | 找不到对应的音频文件", lrcPath)
	}
	if !fileExists(lrcPath) {
		return fmt.Errorf("失败 | %s | 找不到对应的歌词文件", mp3Path)
	}

	lyrics, err := readPlainLyrics(lrcPath)
	if err != nil {
		return err
	}
	if lyrics == "" {
		return fmt.Errorf("失败 | %s | 歌词内容为空", lrcPath)
	}

	if err := writeLyrics(mp3Path, lyrics); err != nil {
		return err
	}

	fmt.Printf("成功 | %s\n", mp3Path)
	return nil
}

func replaceExt(path string, ext string) string {
	return strings.TrimSuffix(path, filepath.Ext(path)) + ext
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func readPlainLyrics(lrcPath string) (string, error) {
	data, err := os.ReadFile(lrcPath)
	if err != nil {
		return "", fmt.Errorf("失败 | %s | 读取失败: %v", lrcPath, err)
	}

	lyrics := normalizeLyrics(string(data))

	return lyrics, nil
}

func normalizeLyrics(text string) string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")

	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))

	for _, line := range lines {
		line = timeTagPattern.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		line = plainLyricLine(line)
		if line != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func plainLyricLine(line string) string {
	if !strings.HasPrefix(line, "{") {
		return line
	}

	var rich struct {
		Content []struct {
			Text string `json:"tx"`
		} `json:"c"`
	}
	if err := json.Unmarshal([]byte(line), &rich); err != nil || len(rich.Content) == 0 {
		return line
	}

	var builder strings.Builder
	for _, part := range rich.Content {
		builder.WriteString(part.Text)
	}

	return strings.TrimSpace(builder.String())
}

func writeLyrics(mp3Path string, lyrics string) error {
	info, err := os.Stat(mp3Path)
	if err != nil {
		return fmt.Errorf("失败 | %s | 打开失败: %v", mp3Path, err)
	}

	data, err := os.ReadFile(mp3Path)
	if err != nil {
		return fmt.Errorf("失败 | %s | 读取失败: %v", mp3Path, err)
	}

	output, err := embedLyrics(data, lyrics)
	if err != nil {
		return fmt.Errorf("失败 | %s | %v", mp3Path, err)
	}

	if err := os.WriteFile(mp3Path, output, info.Mode()); err != nil {
		return fmt.Errorf("失败 | %s | 写入失败: %v", mp3Path, err)
	}

	return nil
}

type id3Tag struct {
	version     byte
	revision    byte
	flags       byte
	prefix      []byte
	frames      [][]byte
	audio       []byte
	frameCount  int
	parsedBytes int
}

func embedLyrics(data []byte, lyrics string) ([]byte, error) {
	tag, ok, err := readID3Tag(data)
	if err != nil {
		return nil, err
	}
	if !ok {
		return buildNewID3v23Tag(lyrics, data), nil
	}

	var frame []byte
	switch tag.version {
	case 2:
		tag.frames = filterFrames(tag.frames, "ULT")
		frame = buildFrameV22("ULT", buildLyricsPayload(lyrics))
	case 3:
		tag.frames = filterFrames(tag.frames, "USLT")
		frame = buildFrameV23("USLT", buildLyricsPayload(lyrics))
	case 4:
		tag.frames = filterFrames(tag.frames, "USLT")
		frame = buildFrameV24("USLT", buildLyricsPayload(lyrics))
	default:
		return nil, fmt.Errorf("不支持的 ID3v2.%d 标签", tag.version)
	}

	tag.frames = append(tag.frames, frame)
	return buildExistingID3Tag(tag), nil
}

func readID3Tag(data []byte) (id3Tag, bool, error) {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return id3Tag{}, false, nil
	}

	version := data[3]
	if version < 2 || version > 4 {
		return id3Tag{}, false, fmt.Errorf("不支持的 ID3v2.%d 标签", version)
	}

	flags := data[5]
	if flags&0x80 != 0 {
		return id3Tag{}, false, fmt.Errorf("不支持带 unsynchronisation 标记的 ID3 标签，已停止写入以避免破坏元数据")
	}

	tagSize, ok := decodeSynchsafe(data[6:10])
	if !ok {
		return id3Tag{}, false, fmt.Errorf("ID3 标签大小无效")
	}

	tagEnd := 10 + tagSize
	if flags&0x10 != 0 {
		tagEnd += 10
	}
	if tagEnd > len(data) {
		return id3Tag{}, false, fmt.Errorf("ID3 标签长度超过文件大小")
	}

	body := data[10 : 10+tagSize]
	framesOffset := extendedHeaderSize(version, flags, body)
	if framesOffset > len(body) {
		return id3Tag{}, false, fmt.Errorf("ID3 扩展头无效")
	}

	frames, parsed := parseFrames(version, body[framesOffset:])
	if len(frames) == 0 && hasNonZeroBytes(body[framesOffset:]) {
		return id3Tag{}, false, fmt.Errorf("无法可靠解析原 ID3 帧，已停止写入以避免删除封面、作者等元数据")
	}
	if hasNonZeroBytes(body[framesOffset+parsed:]) {
		return id3Tag{}, false, fmt.Errorf("只解析到部分 ID3 帧，已停止写入以避免删除封面、作者等元数据")
	}

	return id3Tag{
		version:     version,
		revision:    data[4],
		flags:       flags &^ 0x10,
		prefix:      append([]byte(nil), body[:framesOffset]...),
		frames:      frames,
		audio:       data[tagEnd:],
		frameCount:  len(frames),
		parsedBytes: parsed,
	}, true, nil
}

func extendedHeaderSize(version byte, flags byte, body []byte) int {
	if flags&0x40 == 0 || len(body) < 4 {
		return 0
	}

	switch version {
	case 3:
		size := int(binary.BigEndian.Uint32(body[:4]))
		if size < 0 || 4+size > len(body) {
			return 0
		}
		return 4 + size
	case 4:
		size, ok := decodeSynchsafe(body[:4])
		if !ok || size > len(body) {
			return 0
		}
		return size
	default:
		return 0
	}
}

func parseFrames(version byte, body []byte) ([][]byte, int) {
	frames := make([][]byte, 0)
	offset := 0

	for {
		headerSize := frameHeaderSize(version)
		if len(body)-offset < headerSize {
			break
		}
		header := body[offset : offset+headerSize]
		if isPadding(header) {
			break
		}

		id := frameID(version, header)
		if !isFrameID(id) {
			break
		}

		size, ok := frameSize(version, header)
		if !ok || size <= 0 || offset+headerSize+size > len(body) {
			break
		}

		raw := body[offset : offset+headerSize+size]
		frames = append(frames, append([]byte(nil), raw...))
		offset += headerSize + size
	}

	return frames, offset
}

func frameHeaderSize(version byte) int {
	if version == 2 {
		return 6
	}
	return 10
}

func frameID(version byte, header []byte) string {
	if version == 2 {
		return string(header[:3])
	}
	return string(header[:4])
}

func frameSize(version byte, header []byte) (int, bool) {
	switch version {
	case 2:
		return int(header[3])<<16 | int(header[4])<<8 | int(header[5]), true
	case 4:
		return decodeSynchsafe(header[4:8])
	default:
		return int(binary.BigEndian.Uint32(header[4:8])), true
	}
}

func filterFrames(frames [][]byte, removeID string) [][]byte {
	filtered := make([][]byte, 0, len(frames))
	for _, frame := range frames {
		if frameMatchesID(frame, removeID) {
			continue
		}
		filtered = append(filtered, frame)
	}
	return filtered
}

func frameMatchesID(frame []byte, id string) bool {
	if len(id) == 3 {
		return len(frame) >= 3 && string(frame[:3]) == id
	}
	return len(frame) >= 4 && string(frame[:4]) == id
}

func buildLyricsPayload(lyrics string) []byte {
	payload := []byte{0x01, 'c', 'h', 'i', 0xFF, 0xFE, 0x00, 0x00}
	payload = append(payload, encodeUTF16LE(lyrics)...)
	return payload
}

func encodeUTF16LE(text string) []byte {
	units := utf16.Encode([]rune(text))
	data := make([]byte, 0, len(units)*2)
	for _, unit := range units {
		data = append(data, byte(unit), byte(unit>>8))
	}
	return data
}

func buildFrameV22(id string, payload []byte) []byte {
	frame := make([]byte, 6+len(payload))
	copy(frame[:3], id)
	frame[3] = byte((len(payload) >> 16) & 0xFF)
	frame[4] = byte((len(payload) >> 8) & 0xFF)
	frame[5] = byte(len(payload) & 0xFF)
	copy(frame[6:], payload)
	return frame
}

func buildFrameV23(id string, payload []byte) []byte {
	frame := make([]byte, 10+len(payload))
	copy(frame[:4], id)
	binary.BigEndian.PutUint32(frame[4:8], uint32(len(payload)))
	copy(frame[10:], payload)
	return frame
}

func buildFrameV24(id string, payload []byte) []byte {
	frame := make([]byte, 10+len(payload))
	copy(frame[:4], id)
	encodeSynchsafe(frame[4:8], len(payload))
	copy(frame[10:], payload)
	return frame
}

func buildNewID3v23Tag(lyrics string, audio []byte) []byte {
	body := buildFrameV23("USLT", buildLyricsPayload(lyrics))
	body = append(body, make([]byte, 2048)...)
	output := make([]byte, 0, 10+len(body)+len(audio))
	output = append(output, buildID3Header(3, 0, 0, len(body))...)
	output = append(output, body...)
	output = append(output, audio...)
	return output
}

func buildExistingID3Tag(tag id3Tag) []byte {
	body := make([]byte, 0, len(tag.prefix)+tag.parsedBytes+2048)
	body = append(body, tag.prefix...)
	body = append(body, bytes.Join(tag.frames, nil)...)
	body = append(body, make([]byte, 2048)...)

	output := make([]byte, 0, 10+len(body)+len(tag.audio))
	output = append(output, buildID3Header(tag.version, tag.revision, tag.flags, len(body))...)
	output = append(output, body...)
	output = append(output, tag.audio...)
	return output
}

func buildID3Header(version byte, revision byte, flags byte, size int) []byte {
	header := []byte{'I', 'D', '3', version, revision, flags, 0, 0, 0, 0}
	encodeSynchsafe(header[6:10], size)
	return header
}

func decodeSynchsafe(raw []byte) (int, bool) {
	if len(raw) != 4 {
		return 0, false
	}

	size := 0
	for _, b := range raw {
		if b&0x80 != 0 {
			return 0, false
		}
		size = (size << 7) | int(b)
	}
	return size, true
}

func encodeSynchsafe(dst []byte, size int) {
	dst[0] = byte((size >> 21) & 0x7F)
	dst[1] = byte((size >> 14) & 0x7F)
	dst[2] = byte((size >> 7) & 0x7F)
	dst[3] = byte(size & 0x7F)
}

func isFrameID(id string) bool {
	if len(id) != 3 && len(id) != 4 {
		return false
	}

	for _, ch := range id {
		if (ch < 'A' || ch > 'Z') && (ch < '0' || ch > '9') {
			return false
		}
	}
	return true
}

func isPadding(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return false
		}
	}
	return true
}

func hasNonZeroBytes(data []byte) bool {
	for _, b := range data {
		if b != 0 {
			return true
		}
	}
	return false
}
