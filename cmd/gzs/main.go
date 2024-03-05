package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"

	"github.com/jonjohnsonjr/gzs"
)

const spanSize = 1 << 22

func main() {
	if err := run(context.Background()); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	if len(os.Args) < 2 {
		return fmt.Errorf("usage: %s index", os.Args[0])
	}

	switch os.Args[1] {
	case "index":
		return index()
	case "ls":
		return ls()
	case "cat":
		return cat()
	}

	return fmt.Errorf("usage: %s <index|ls|cat>", os.Args[0])
}

func index() error {
	rc := os.Stdin
	if len(os.Args) == 3 {
		f, err := os.Open(os.Args[2])
		if err != nil {
			return err
		}
		rc = f
	}

	indexer, _, _, _, err := gzs.NewIndexer(rc, os.Stdout, spanSize, "application/tar+gzip")
	if err != nil {
		return err
	}
	if indexer == nil {
		return fmt.Errorf("couldn't index for some reason")
	}
	for {
		// Make sure we hit the end.
		_, err := indexer.Next()
		if errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return fmt.Errorf("indexer.Next: %w", err)
		}
	}

	toc, err := indexer.TOC()
	if err != nil {
		return err
	}

	log.Printf("wrote %d checkpoints for %d files", len(toc.Checkpoints), len(toc.Files))

	return indexer.Close()
}

type fileReader struct {
	f *os.File
}

func (fr *fileReader) Reader(ctx context.Context, off int64, end int64) (io.ReadCloser, error) {
	return io.NopCloser(io.NewSectionReader(fr.f, off, end-off)), nil
}

func ls() error {
	if len(os.Args) != 3 {
		return fmt.Errorf("usage: ls file")
	}

	f, err := os.Open(os.Args[2])
	if err != nil {
		return err
	}

	fr := &fileReader{f}

	idx, err := gzs.NewIndex(fr, nil, nil)
	if err != nil {
		return err
	}

	toc := idx.TOC()

	for _, f := range toc.Files {
		hdr := gzs.TarHeader(&f).FileInfo()
		fmt.Println(fs.FormatFileInfo(hdr))
	}

	return nil
}

func cat() error {
	if len(os.Args) != 5 {
		return fmt.Errorf("usage: cat file original.tar.gz index.gzs")
	}

	idxf, err := os.Open(os.Args[4])
	if err != nil {
		return err
	}

	idxfr := &fileReader{idxf}

	idx, err := gzs.NewIndex(idxfr, nil, nil)
	if err != nil {
		return err
	}

	ogf, err := os.Open(os.Args[3])
	if err != nil {
		return err
	}

	ogfr := &fileReader{ogf}

	tocf, err := idx.Locate(os.Args[2])
	if err != nil {
		return err
	}

	rc, err := gzs.ExtractFile(context.TODO(), idx, ogfr, tocf)
	if err != nil {
		return err
	}

	if _, err := io.Copy(os.Stdout, rc); err != nil {
		return err
	}

	return nil
}
