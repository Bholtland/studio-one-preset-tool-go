package file

import (
	"context"
	"encoding/xml"
	"fmt"
	"github.com/saracen/fastzip"
	"io"
	"os"
	"path"
	"path/filepath"
)

func WriteXML[T interface{}](content T, filePath string) error {
	// Create XML file
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	if _, err = file.WriteString(xml.Header); err != nil {
		return err
	}

	// Write XML content
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(content)
	if err != nil {
		return err
	}

	// Close file
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}

func ReadXML[T interface{}](filePath string) (*T, error) {
	file, err := os.Open(path.Join(filePath))
	if err != nil {
		return nil, fmt.Errorf("Error opening XML file: %w", err)
	}
	defer file.Close()

	rawXML, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading XML file: %w", err)
	}

	var unmarshalledXML *T

	err = xml.Unmarshal(rawXML, &unmarshalledXML)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling XML: %w", err)
	}

	return unmarshalledXML, nil
}

func Copy(src string, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func Compress(ctx context.Context, srcPath string, destPath string, destFileName string) error {
	if err := os.MkdirAll(destPath, os.ModeDir); err != nil {
		return err
	}

	// Create archive file
	w, err := os.Create(path.Join(destPath, destFileName))
	if err != nil {
		return err
	}
	defer w.Close()

	// Create new Archiver
	a, err := fastzip.NewArchiver(w, srcPath)
	if err != nil {
		return err
	}
	defer a.Close()

	// Walk directory, adding the files we want to add
	files := make(map[string]os.FileInfo)
	err = filepath.Walk(srcPath, func(pathname string, info os.FileInfo, err error) error {
		files[pathname] = info
		return nil
	})

	// Archive
	if err = a.Archive(ctx, files); err != nil {
		return err
	}

	return nil
}

func Extract(ctx context.Context, src string, dst string) error {
	if err := os.MkdirAll(dst, os.ModeTemporary); err != nil {
		return err
	}

	extractor, err := fastzip.NewExtractor(src, dst)
	if err != nil {
		return err
	}
	defer extractor.Close()

	return extractor.Extract(ctx)
}
