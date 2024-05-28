package parser

import "regexp"

// custom
const (
	PlaylistTypeVOD   PlaylistType = "VOD"
	PlaylistTypeEvent PlaylistType = "EVENT"

	CryptMethodAES  CryptMethod = "AES-128"
	CryptMethodNONE CryptMethod = "NONE"

	extM3U           = "#EXTM3U"
	extInfPrefix     = "#EXTINF:"
	extByteRange     = "#EXT-X-BYTERANGE:"
	extKey           = "#EXT-X-KEY"
	extStreamInf     = "#EXT-X-STREAM-INF:"
	endList          = "#EXT-X-ENDLIST"
	playlistType     = "#EXT-X-PLAYLIST-TYPE:"
	targetDuration   = "#EXT-X-TARGETDURATION:"
	mediaSequence    = "#EXT-X-MEDIA-SEQUENCE:"
	version          = "#EXT-X-VERSION:"
	invalidExtM3U    = "invalid m3u8, missing #EXTM3U in line 1"
	invalidURI       = "invalid EXT-X-STREAM-INF URI, line: %d"
	duplicateExtInf  = "duplicate EXTINF: %s, line: %d"
	duplicateExtByte = "duplicate EXT-X-BYTERANGE: %s, line: %d"
	invalidLine      = "invalid line: %s"
	invalidExtKey    = "invalid EXT-X-KEY: %s, line: %d"
	invalidKeyMethod = "invalid EXT-X-KEY method: %s, line: %d"
)

var linePattern = regexp.MustCompile(`(?P<key>[^=]+)=(?P<value>\"[^\"]*\"|[^,]*)`)
