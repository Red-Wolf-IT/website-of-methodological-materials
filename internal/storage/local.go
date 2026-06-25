package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const uploadsURLPrefix = "/uploads/"

type FileStorage struct {
	dir string
}

func NewFileStorage(dir string) (*FileStorage, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve storage dir: %w", err)
	}

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	return &FileStorage{dir: absDir}, nil
}

// Save сохраняет файл и возвращает web-путь для БД (/uploads/...)
func (s *FileStorage) Save(manualID uuid.UUID, originalName string, src io.Reader) (string, error) {
	safeName := filepath.Base(originalName)
	if safeName == "" || safeName == "." {
		return "", fmt.Errorf("invalid file name")
	}

	filename := fmt.Sprintf("%s_%s", manualID.String(), safeName)
	diskPath := filepath.Join(s.dir, filename)

	dst, err := os.Create(diskPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(diskPath)
		return "", fmt.Errorf("write file: %w", err)
	}

	return uploadsURLPrefix + filename, nil
}

// Open открывает файл по web-пути из БД
func (s *FileStorage) Open(webPath string) (io.ReadCloser, error) {
	diskPath, err := s.resolveWebPath(webPath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(diskPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %w", err)
		}
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

func (s *FileStorage) Remove(webPath string) error {
	if webPath == "" {
		return nil
	}

	diskPath, err := s.resolveWebPath(webPath)
	if err != nil {
		return err
	}

	if err := os.Remove(diskPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove file: %w", err)
	}

	return nil
}

func (s *FileStorage) resolveWebPath(webPath string) (string, error) {
	if !strings.HasPrefix(webPath, uploadsURLPrefix) {
		return "", fmt.Errorf("invalid file path")
	}

	filename := filepath.Base(strings.TrimPrefix(webPath, uploadsURLPrefix))
	if filename == "" || filename == "." {
		return "", fmt.Errorf("invalid file path")
	}

	diskPath := filepath.Join(s.dir, filename)

	absDisk, err := filepath.Abs(diskPath)
	if err != nil {
		return "", fmt.Errorf("resolve file path: %w", err)
	}

	rel, err := filepath.Rel(s.dir, absDisk)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path traversal detected")
	}

	return absDisk, nil
}
