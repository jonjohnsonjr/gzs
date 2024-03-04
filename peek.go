package gzs

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"io"

	"github.com/jonjohnsonjr/gzs/internal/and"
)

const (
	magicGNU, versionGNU     = "ustar ", " \x00"
	magicUSTAR, versionUSTAR = "ustar\x00", "00"
)

type PeekReader interface {
	io.Reader
	Peek(n int) ([]byte, error)
}

// Peek detects streams of:
// tar
// tar+gzip
// gzip
func Peek(rc io.ReadCloser) (string, io.ReadCloser, io.ReadCloser, error) {
	buf := bufio.NewReaderSize(rc, 1<<16)
	pr := &and.ReadCloser{Reader: buf, CloseFunc: rc.Close}

	// Should be enough to read first block?
	zb, err := buf.Peek(1024)
	if err != nil {
		if err != io.EOF {
			return "", pr, nil, err
		}
	}

	br := bytes.NewReader(zb)
	if ok, zpr, err := gzipPeek(br); err != nil {
		return "", pr, nil, err
	} else if ok {
		zr, err := gzip.NewReader(zpr)
		if err != nil {
			return "", pr, nil, err
		}
		ok, tpr, err := tarPeek(zr)
		if err != nil {
			return "", pr, nil, err
		}
		if ok {
			return "tar+gzip", pr, tpr, nil
		} else {
			return "gzip", pr, tpr, nil
		}
	}

	br = bytes.NewReader(zb)
	if ok, tpr, err := tarPeek(io.NopCloser(br)); err != nil {
		return "", pr, nil, err
	} else if ok {
		return "tar", pr, tpr, nil
	}

	return "", pr, nil, nil
}

// CheckHeader checks whether the first bytes from a PeekReader match an expected header
func checkHeader(pr PeekReader, expectedHeader []byte) (bool, PeekReader, error) {
	header, err := pr.Peek(len(expectedHeader))
	if err != nil {
		// https://github.com/google/go-containerregistry/issues/367
		if err == io.EOF {
			return false, pr, nil
		}
		return false, pr, err
	}
	return bytes.Equal(header, expectedHeader), pr, nil
}

func tarPeek(rc io.ReadCloser) (bool, io.ReadCloser, error) {
	// Make sure it's more than 512
	var pr PeekReader
	if p, ok := rc.(PeekReader); ok {
		pr = p
	} else {
		// For tar peek.
		pr = bufio.NewReaderSize(rc, 1<<16)
	}
	prc := &and.ReadCloser{Reader: pr, CloseFunc: rc.Close}

	block, err := pr.Peek(512)
	if err != nil {
		// https://github.com/google/go-containerregistry/issues/367
		if err == io.EOF {
			return false, prc, nil
		}
		return false, prc, err
	}

	magic := string(block[257:][:6])
	isTar := magic == magicGNU || magic == magicUSTAR
	return isTar, prc, nil
}

// MagicHeader is the start of gzip files.
var gzipMagicHeader = []byte{'\x1f', '\x8b'}

// Peek detects whether the input stream is gzip compressed.
//
// If r implements Peek, we will use that directly, otherwise a small number
// of bytes are buffered to Peek at the gzip header, and the returned
// PeekReader can be used as a replacement for the consumed input io.Reader.
func gzipPeek(r io.Reader) (bool, PeekReader, error) {
	var pr PeekReader
	if p, ok := r.(PeekReader); ok {
		pr = p
	} else {
		// For tar peek.
		pr = bufio.NewReaderSize(r, 1024)
	}
	header, err := pr.Peek(2)
	if err != nil {
		// https://github.com/google/go-containerregistry/issues/367
		if err == io.EOF {
			return false, pr, nil
		}
		return false, pr, err
	}
	return bytes.Equal(header, gzipMagicHeader), pr, nil
}
