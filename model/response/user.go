package response

import (
	"zldface_server/recognition"
)

type FaceMatchResult struct {
	recognition.Closest
	FilePath string `json:"filePath"`
}
