package util

func TruncateSessionID(id string) string {
	if len(id) > 8 {
		return id[:8]
	}
	return id
}
