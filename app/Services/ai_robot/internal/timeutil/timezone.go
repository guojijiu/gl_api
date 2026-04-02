package timeutil

import "time"

var shanghaiLocation = loadShanghaiLocation()

func NowInShanghai() time.Time {
	return time.Now().In(shanghaiLocation)
}

func FormatInShanghai(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.In(shanghaiLocation).Format(time.RFC3339)
}

func loadShanghaiLocation() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil && loc != nil {
		return loc
	}
	return time.FixedZone("CST", 8*3600)
}
