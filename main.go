package main

import (
	"archive/zip"
	"io"
	"io/fs"
	"os"
)

type Archiver struct {
	Dest *os.File
	w    *zip.Writer
}

func NewArchiver(archive *os.File) *Archiver {
	return &Archiver{Dest: archive, w: zip.NewWriter(archive)}
}

func (a *Archiver) Archive(files ...*os.File) error {
	for _, file := range files {
		a.WriteFile(file)
	}

	a.w.Close()
	return nil
}

func (a *Archiver) WriteFile(file *os.File) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}

	writer, err := a.createFile(info)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	if err != nil {
		return err
	}

	return nil
}

func (a *Archiver) createFile(info fs.FileInfo) (io.Writer, error) {
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return nil, err
	}

	writer, err := a.w.CreateHeader(header)
	if err != nil {
		return nil, err
	}

	return writer, nil
}

func main() {
}
