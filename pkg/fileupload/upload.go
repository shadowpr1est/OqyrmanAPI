package fileupload

import "io"

type File struct {
	Filename    string
	Reader      io.Reader
	Size        int64
	ContentType string
}
