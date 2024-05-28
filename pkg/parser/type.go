package parser

import "net/url"

type (
	// PlaylistType is the type of playlist
	PlaylistType string
	// CryptMethod is the method of encryption
	CryptMethod string

	// Result model
	Result struct {
		URL  *url.URL
		M3U8 *M3U8
		Keys map[int]string
	}

	// M3U8 model
	M3U8 struct {
		Version        int8   // EXT-X-VERSION:version
		MediaSequence  uint64 // Default 0, #EXT-X-MEDIA-SEQUENCE:sequence
		Segments       []*Segment
		MasterPlaylist []*MasterPlaylist
		Keys           map[int]*Key
		EndList        bool         // #EXT-X-ENDLIST
		PlaylistType   PlaylistType // VOD or EVENT
		TargetDuration float64      // #EXT-X-TARGETDURATION:duration
	}

	// Segment model
	Segment struct {
		URI      string
		KeyIndex int
		Title    string  // #EXTINF: duration,<title>
		Duration float32 // #EXTINF: duration,<title>
		Length   uint64  // #EXT-X-BYTERANGE: length[@offset]
		Offset   uint64  // #EXT-X-BYTERANGE: length[@offset]
	}

	// MasterPlaylist #EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=240000,RESOLUTION=416x234,CODECS="avc1.42e00a,mp4a.40.2"
	MasterPlaylist struct {
		URI        string
		BandWidth  uint32
		Resolution string
		Codecs     string
		ProgramID  uint32
	}

	// Key #EXT-X-KEY:METHOD=AES-128,URI="key.key"
	Key struct {
		// 'AES-128' or 'NONE'
		// If the encryption method is NONE, the URI and the IV attributes MUST NOT be present
		Method CryptMethod
		URI    string
		IV     string
	}
)
