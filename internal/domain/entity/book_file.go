package entity

import (
	"fmt"
	"net/http"

	"github.com/google/uuid"
)

// BookFileFormat — typed enum of allowed book file formats.
// The format encodes the file category (audio vs document) — IsAudio() is derived, not client-supplied.
type BookFileFormat string

const (
	BookFileFormatPDF  BookFileFormat = "pdf"
	BookFileFormatEPUB BookFileFormat = "epub"
	BookFileFormatMP3  BookFileFormat = "mp3"

	// MaxDocumentSize — 50 MB hard limit for PDF/EPUB uploads.
	MaxDocumentSize int64 = 50 * 1024 * 1024
	// MaxAudioSize — 200 MB hard limit for MP3 uploads.
	MaxAudioSize int64 = 200 * 1024 * 1024
)

// IsAudio returns true only for audio formats.
// Derived deterministically from the format — never accepted from the client.
func (f BookFileFormat) IsAudio() bool {
	return f == BookFileFormatMP3
}

// IsValid reports whether the format is one of the allowed values.
func (f BookFileFormat) IsValid() bool {
	switch f {
	case BookFileFormatPDF, BookFileFormatEPUB, BookFileFormatMP3:
		return true
	}
	return false
}

// MaxSize returns the maximum allowed file size for this format.
func (f BookFileFormat) MaxSize() int64 {
	if f.IsAudio() {
		return MaxAudioSize
	}
	return MaxDocumentSize
}

// ValidateMagicBytes checks that the first 512 bytes of the file match the declared format.
// Protects against renamed files (e.g. shell.php → book.pdf).
func (f BookFileFormat) ValidateMagicBytes(buf []byte) error {
	detected := http.DetectContentType(buf)
	switch f {
	case BookFileFormatPDF:
		if detected != "application/pdf" {
			return fmt.Errorf("%w: file bytes do not match pdf format (got %s)", ErrValidation, detected)
		}
	case BookFileFormatEPUB:
		// EPUB is a ZIP archive — DetectContentType returns application/zip.
		if detected != "application/zip" {
			return fmt.Errorf("%w: file bytes do not match epub format (got %s)", ErrValidation, detected)
		}
	case BookFileFormatMP3:
		if !isMPEGAudio(buf) {
			return fmt.Errorf("%w: file bytes do not match mp3 format", ErrValidation)
		}
	}
	return nil
}

// isMPEGAudio detects MP3 by checking for an ID3v2 tag or MPEG frame sync word.
// http.DetectContentType is unreliable for audio/mpeg — manual byte check is required.
func isMPEGAudio(buf []byte) bool {
	if len(buf) < 3 {
		return false
	}
	// ID3v2 tag: "ID3"
	if buf[0] == 0x49 && buf[1] == 0x44 && buf[2] == 0x33 {
		return true
	}
	// MPEG frame sync: 0xFF followed by byte with top 3 bits set (0xE0)
	if buf[0] == 0xFF && (buf[1]&0xE0) == 0xE0 {
		return true
	}
	return false
}

type BookFile struct {
	ID      uuid.UUID      `db:"id"`
	BookID  uuid.UUID      `db:"book_id"`
	Format  BookFileFormat `db:"format"`
	FileURL string         `db:"file_url"`
	IsAudio bool           `db:"is_audio"`
}
