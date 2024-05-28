package parser

import (
	"errors"
	"fmt"
	"net/url"

	"loki/pkg/tools"
)

// Parse parses the provided endpoint and returns a Result
func Parse(endpoint string) (*Result, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %v", err)
	}

	body, err := tools.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer body.Close()

	m3u8, err := parse(body)
	if err != nil {
		return nil, fmt.Errorf("parse m3u8 failed: %v", err)
	}

	if len(m3u8.MasterPlaylist) > 0 {
		sf := m3u8.MasterPlaylist[0]
		return Parse(tools.ResolveURL(u, sf.URI))
	}

	if len(m3u8.Segments) == 0 {
		return nil, errors.New("no TS file description found in the M3U8 file")
	}

	result := &Result{
		URL:  u,
		M3U8: m3u8,
		Keys: make(map[int]string),
	}

	if err := fetchKeys(result, u); err != nil {
		return nil, err
	}

	return result, nil
}
