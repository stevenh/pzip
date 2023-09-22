package pzip

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/klauspost/compress/zip"
	derrors "github.com/pkg/errors"
	"github.com/pzip/pool"
)

type extractor struct {
	outputDir      string
	archiveReader  *zip.ReadCloser
	fileWorkerPool pool.WorkerPool[zip.File]
	concurrency    int
}

func NewExtractor(outputDir string) *extractor {
	absOutputDir, _ := filepath.Abs(outputDir)
	e := &extractor{outputDir: absOutputDir, concurrency: runtime.GOMAXPROCS(0)}

	fileExecutor := func(file *zip.File) error {
		if err := e.extractFile(file); err != nil {
			return derrors.Wrapf(err, "ERROR: could not extract file %s", file.Name)
		}

		return nil
	}

	fileWorkerPool, _ := pool.NewFileWorkerPool(fileExecutor, &pool.Config{Concurrency: e.concurrency, Capacity: 1000})
	e.fileWorkerPool = fileWorkerPool
	return e
}

func (e *extractor) Extract(ctx context.Context, archivePath string) (err error) {
	e.archiveReader, err = zip.OpenReader(archivePath)
	if err != nil {
		return derrors.Errorf("ERROR: could not read archive at %s: %v", archivePath, err)
	}

	e.fileWorkerPool.Start(ctx)

	for _, file := range e.archiveReader.File {
		e.fileWorkerPool.Enqueue(file)
	}

	if err = e.fileWorkerPool.Close(); err != nil {
		return derrors.Wrap(err, "ERROR: could not close file worker pool")
	}

	return err
}

func (e *extractor) Close() error {
	err := e.archiveReader.Close()
	if err != nil {
		return derrors.New("ERROR: could not close archive reader")
	}

	return nil
}

func (e *extractor) extractFile(file *zip.File) (err error) {
	outputPath := e.outputPath(file.Name)

	if err = os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil { // TODO: need to set correct file mode as specified by file
		return derrors.Errorf("ERROR: could not directories %s: %+v", outputPath, err)
	}

	if e.isDir(file.Name) {
		return nil
	}

	outputFile, err := os.Create(e.outputPath(file.Name))
	if err != nil {
		return derrors.Errorf("ERROR: could not create file %s: %v", outputPath, err)
	}
	defer func() {
		err = errors.Join(err, outputFile.Close())
	}()

	fileContent, _ := file.Open()
	defer func() {
		err = errors.Join(err, fileContent.Close())
	}()

	_, err = io.Copy(outputFile, fileContent)
	if err != nil {
		return derrors.Errorf("ERROR: could not decompress file %s", file.Name)
	}

	return nil
}

func (e *extractor) isDir(name string) bool {
	return strings.HasSuffix(filepath.ToSlash(name), "/")
}

func (e *extractor) outputPath(name string) string {
	return filepath.Join(e.outputDir, name)
}
