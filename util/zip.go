package util

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/mholt/archives"
	"github.com/yeka/zip"
)

type SeekReaderAt interface {
	io.ReaderAt
	io.Seeker
}

type FileInArchive struct {
	io.ReadCloser
	info fs.FileInfo
}

func (af FileInArchive) Stat() (fs.FileInfo, error) { return af.info, nil }

type Zip struct {
	Password         string
	EncryptionMethod zip.EncryptionMethod
}

func (z Zip) archiveOneFile(ctx context.Context, zw *zip.Writer, idx int, file archives.FileInfo) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	name := file.NameInArchive
	if name == "" {
		name = file.Name()
	}

	w, err := zw.Encrypt(name, z.Password, z.EncryptionMethod)
	if err != nil {
		return fmt.Errorf("creating header for file %d: %s: %w", idx, file.Name(), err)
	}

	fileReader, err := file.Open()
	if err != nil {
		return err
	}
	defer func() {
		_ = fileReader.Close()
	}()

	if _, err := io.Copy(w, fileReader); err != nil {
		return fmt.Errorf("writing file %d: %s: %w", idx, file.Name(), err)
	}

	return nil
}

func (z Zip) Archive(ctx context.Context, output io.Writer, files []archives.FileInfo) error {
	zw := zip.NewWriter(output)
	defer func() {
		_ = zw.Close()
	}()

	for i, file := range files {
		if err := z.archiveOneFile(ctx, zw, i, file); err != nil {
			return err
		}
	}

	return nil
}

func (z Zip) Extract(ctx context.Context, sourceArchive io.Reader, handleFile archives.FileHandler) error {
	sra, ok := sourceArchive.(SeekReaderAt)
	if !ok {
		return fmt.Errorf("input type must be an io.ReaderAt and io.Seeker because of zip format constraints")
	}

	size, err := StreamSizeBySeeking(sra)
	if err != nil {
		return fmt.Errorf("determining stream size: %w", err)
	}

	zr, err := zip.NewReader(sra, size)
	if err != nil {
		return err
	}

	for i, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}

		if f.IsEncrypted() {
			f.SetPassword(z.Password)
		}

		info := f.FileInfo()
		file := archives.FileInfo{
			FileInfo:      info,
			Header:        f.FileHeader,
			NameInArchive: f.Name,
			Open: func() (fs.File, error) {
				openedFile, err := f.Open()
				if err != nil {
					return nil, err
				}

				return FileInArchive{openedFile, info}, nil
			},
		}

		err := handleFile(ctx, file)
		if errors.Is(err, fs.SkipAll) {
			break
		}
		if err != nil {
			return fmt.Errorf("handling file %d: %s: %w", i, f.Name, err)
		}
	}

	return nil
}

// Interface guard
var (
	_ archives.Archiver  = Zip{}
	_ archives.Extractor = Zip{}
)
