package parser

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"loki/pkg/tools"
	"net/url"
	"strconv"
	"strings"
)

// parse parses the M3U8 content from the provided reader
func parse(reader io.Reader) (*M3U8, error) {
	lines, err := readLines(reader)
	if err != nil {
		return nil, err
	}

	m3u8 := &M3U8{
		Keys: make(map[int]*Key),
	}

	if err := processLines(lines, m3u8); err != nil {
		return nil, err
	}

	return m3u8, nil
}

// readLines reads all lines from the provided reader
func readLines(reader io.Reader) ([]string, error) {
	s := bufio.NewScanner(reader)
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

// processLines processes each line of the M3U8 content
func processLines(lines []string, m3u8 *M3U8) error {
	if len(lines) == 0 || lines[0] != extM3U {
		return errors.New(invalidExtM3U)
	}

	var (
		seg      *Segment
		key      *Key
		keyIndex = 0
		extInf   bool
		extByte  bool
	)

	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		switch {
		case strings.HasPrefix(line, playlistType):
			if err := parsePlaylistType(line, m3u8, i); err != nil {
				return err
			}
		case strings.HasPrefix(line, targetDuration):
			if err := parseTargetDuration(line, m3u8); err != nil {
				return err
			}
		case strings.HasPrefix(line, mediaSequence):
			if err := parseMediaSequence(line, m3u8); err != nil {
				return err
			}
		case strings.HasPrefix(line, version):
			if err := parseVersion(line, m3u8); err != nil {
				return err
			}
		case strings.HasPrefix(line, extStreamInf):
			mp, err := parseMasterPlaylist(line)
			if err != nil {
				return err
			}
			i++
			if i >= len(lines) || lines[i] == "" || strings.HasPrefix(lines[i], "#") {
				return fmt.Errorf(invalidURI, i+1)
			}
			mp.URI = lines[i]
			m3u8.MasterPlaylist = append(m3u8.MasterPlaylist, mp)
		case strings.HasPrefix(line, extInfPrefix):
			if extInf {
				return fmt.Errorf(duplicateExtInf, line, i+1)
			}
			if seg == nil {
				seg = new(Segment)
			}
			if err := parseExtInf(line, seg); err != nil {
				return err
			}
			extInf = true
			seg.KeyIndex = keyIndex
		case strings.HasPrefix(line, extByteRange):
			if extByte {
				return fmt.Errorf(duplicateExtByte, line, i+1)
			}
			if seg == nil {
				seg = new(Segment)
			}
			if err := parseExtByteRange(line, seg, i); err != nil {
				return err
			}
			extByte = true
		case strings.HasPrefix(line, extKey):
			keyIndex++
			key = new(Key)
			if err := parseExtKey(line, key, keyIndex, m3u8); err != nil {
				return err
			}
		case line == endList:
			m3u8.EndList = true
		case !strings.HasPrefix(line, "#"):
			if extInf {
				seg.URI = line
				m3u8.Segments = append(m3u8.Segments, seg)
				seg = nil
				extInf = false
				extByte = false
			} else {
				return fmt.Errorf(invalidLine, line)
			}
		}
	}
	return nil
}

func parsePlaylistType(line string, m3u8 *M3U8, lineNumber int) error {
	if _, err := fmt.Sscanf(line, "#EXT-X-PLAYLIST-TYPE:%s", &m3u8.PlaylistType); err != nil {
		return err
	}
	isValid := m3u8.PlaylistType == "" || m3u8.PlaylistType == "VOD" || m3u8.PlaylistType == "EVENT"
	if !isValid {
		return fmt.Errorf("invalid playlist type: %s, line: %d", m3u8.PlaylistType, lineNumber+1)
	}
	return nil
}

func parseTargetDuration(line string, m3u8 *M3U8) error {
	_, err := fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%f", &m3u8.TargetDuration)
	return err
}

func parseMediaSequence(line string, m3u8 *M3U8) error {
	_, err := fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &m3u8.MediaSequence)
	return err
}

func parseVersion(line string, m3u8 *M3U8) error {
	_, err := fmt.Sscanf(line, "#EXT-X-VERSION:%d", &m3u8.Version)
	return err
}

func parseMasterPlaylist(line string) (*MasterPlaylist, error) {
	params := parseLineParameters(line)
	if len(params) == 0 {
		return nil, errors.New("empty parameter")
	}
	mp := new(MasterPlaylist)
	for k, v := range params {
		switch k {
		case "BANDWIDTH":
			bandwidth, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, err
			}
			mp.BandWidth = uint32(bandwidth)
		case "RESOLUTION":
			mp.Resolution = v
		case "PROGRAM-ID":
			programID, err := strconv.ParseUint(v, 10, 32)
			if err != nil {
				return nil, err
			}
			mp.ProgramID = uint32(programID)
		case "CODECS":
			mp.Codecs = v
		}
	}
	return mp, nil
}

func parseExtInf(line string, seg *Segment) error {
	var s string
	if _, err := fmt.Sscanf(line, "#EXTINF:%s", &s); err != nil {
		return err
	}
	if strings.Contains(s, ",") {
		split := strings.Split(s, ",")
		seg.Title = split[1]
		s = split[0]
	}
	duration, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return err
	}
	seg.Duration = float32(duration)
	return nil
}

func parseExtByteRange(line string, seg *Segment, lineNumber int) error {
	var b string
	if _, err := fmt.Sscanf(line, "#EXT-X-BYTERANGE:%s", &b); err != nil {
		return err
	}
	if b == "" {
		return fmt.Errorf("invalid EXT-X-BYTERANGE, line: %d", lineNumber+1)
	}
	if strings.Contains(b, "@") {
		split := strings.Split(b, "@")
		offset, err := strconv.ParseUint(split[1], 10, 64)
		if err != nil {
			return err
		}
		seg.Offset = uint64(offset)
		b = split[0]
	}
	length, err := strconv.ParseUint(b, 10, 64)
	if err != nil {
		return err
	}
	seg.Length = uint64(length)
	return nil
}

func validateCryptMethod(method string) error {
	validMethods := []string{"NONE", "AES-128", "SAMPLE-AES"}
	for _, validMethod := range validMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("invalid CryptMethod: %s", method)
}

func parseExtKey(line string, key *Key, keyIndex int, m3u8 *M3U8) error {
	params := parseLineParameters(line)
	if len(params) == 0 {
		return fmt.Errorf(invalidExtKey, line, keyIndex)
	}
	method := params["METHOD"]
	if err := validateCryptMethod(method); err != nil {
		return err
	}
	key.Method = CryptMethod(method)
	key.URI = params["URI"]
	key.IV = params["IV"]
	m3u8.Keys[keyIndex] = key
	return nil
}

func parseLineParameters(line string) map[string]string {
	matches := linePattern.FindAllStringSubmatch(line, -1)
	params := make(map[string]string)
	for _, match := range matches {
		params[match[1]] = strings.Trim(match[2], "\"")
	}
	return params
}

// fetchKeys retrieves decryption keys for the M3U8 segments
func fetchKeys(result *Result, baseURL *url.URL) error {
	for idx, key := range result.M3U8.Keys {
		switch key.Method {
		case "", CryptMethodNONE:
			continue
		case CryptMethodAES:
			keyURL := tools.ResolveURL(baseURL, key.URI)
			keyData, err := fetchKey(keyURL)
			if err != nil {
				return fmt.Errorf("extract key failed: %v", err)
			}
			result.Keys[idx] = keyData
		default:
			return fmt.Errorf("unknown or unsupported encryption method: %s", key.Method)
		}
	}
	return nil
}

// fetchKey requests and reads the decryption key from the specified URL
func fetchKey(keyURL string) (string, error) {
	body, err := tools.Get(keyURL)
	if err != nil {
		return "", fmt.Errorf("request key URL failed: %v", err)
	}
	defer body.Close()

	keyData, err := io.ReadAll(body)
	if err != nil {
		return "", fmt.Errorf("read key data failed: %v", err)
	}

	return string(keyData), nil
}
