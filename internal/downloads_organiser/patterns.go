package downloads_organiser

import (
	"regexp"
)

// The tv show season format. E.g. S07E01 (will fail without the 0 in S07)
var (
	ReShow  = regexp.MustCompile(`(?i)(.+)(S[0-9]+)(E[0-9]+).*`)
	ReMovie = regexp.MustCompile(`^([^()\n]*)\(?([1-2][0-9][0-9][0-9])\)?.*(1080p|2160p|720p).*$`)
)
