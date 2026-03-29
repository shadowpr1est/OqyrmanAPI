package fileupload

import "io"

type File struct {
	Filename    string
	Reader      io.ReadSeeker
	Size        int64
	ContentType string
}
