package pzip

import (
	"context"
	"os"

	"github.com/pkg/errors"
)

type CLI struct {
	ArchivePath string
	Files       []string
}

func (c *CLI) Archive() error {
	archive, err := os.Create(c.ArchivePath)
	if err != nil {
		return errors.Errorf("ERROR: could not create archive at %s", c.ArchivePath)
	}
	defer archive.Close()

	archiver, err := NewArchiver(archive)
	if err != nil {
		return errors.Wrap(err, "ERROR: could not create archiver")
	}
	defer archiver.Close()

	err = archiver.Archive(context.Background(), c.Files)
	if err != nil {
		return errors.Wrapf(err, "ERROR: could not archive files")
	}

	return nil
}
