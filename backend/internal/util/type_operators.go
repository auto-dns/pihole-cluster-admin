package util

func ErrorString(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}
