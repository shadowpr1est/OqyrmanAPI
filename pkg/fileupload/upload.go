package fileupload

import "io"

type File struct {
	Filename    string
	Reader      ReadSeekCloser
	Size        int64
	ContentType string
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}
