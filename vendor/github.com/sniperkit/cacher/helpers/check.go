package helpers

import (
	"bytes"
)

func AddSlashIfNeeded(url string) string {
	var buffer bytes.Buffer
	buffer.WriteString(url)
	if url[len(url)-1] != '/' {
		buffer.WriteString("/")
	}

	return buffer.String()
}

func CheckMapHasKey(inputMap map[string]interface{}, key string) string {
	if val, ok := inputMap[key]; ok {
		if val != nil {
			return val.(string)
		}
	}
	return ""
}
