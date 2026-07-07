package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)



type Interface interface {
  UploadPhotoProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error)
	DeletePhotoProfile(ctx context.Context, fileURL string) error
	UploadPostImages(ctx context.Context, files []*multipart.FileHeader) ([]string, error)
	DeletePostImages(ctx context.Context, fileURLs []string) error
}

type LocalStorage struct {
  uploadDir string
}

func NewLocalStorage(uploadDir string) Interface {
  if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
    _ = os.MkdirAll(uploadDir, os.ModePerm)
  }
  return &LocalStorage{uploadDir: uploadDir}
}

func (s *LocalStorage) UploadPhotoProfile(ctx context.Context, id uuid.UUID, file *multipart.FileHeader) (string, error) {
  ext := filepath.Ext(file.Filename)
  if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
    return "", fmt.Errorf("ekstensi file %s tidak diizinkan", ext)
  }

  uniqueName := fmt.Sprintf("%s_%d%s", uuid.New().String(), time.Now().Unix(), ext)
	dstPath := filepath.Join(s.uploadDir, uniqueName)

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return "/" + dstPath, nil
}

func (s *LocalStorage) DeletePhotoProfile(ctx context.Context, fileURL string) error {
	filePath := strings.TrimPrefix(fileURL, "/")

	if _, err := os.Stat(filePath); err == nil {
		return os.Remove(filePath)
	}
	
	return nil
}

func (s *LocalStorage) UploadPostImages(ctx context.Context, files []*multipart.FileHeader) ([]string, error) {
	var imgURLs []string
	for _, img := range files {
		ext := filepath.Ext(img.Filename)
		if ext != ".jpg" && ext != ".png" && ext != ".jpeg" {
			return nil, fmt.Errorf("ekstensi file %s tidak diizinkan", ext)
		}
		uniqueName := fmt.Sprintf("post_%s_%d%s", uuid.New().String(), time.Now().UnixNano(), ext)
		dstPath := filepath.Join(s.uploadDir, uniqueName)

		src, err := img.Open()
		if err != nil {
			return nil, err
		}
		defer src.Close()

		dst, err := os.Create(dstPath)
		if err != nil {
			return nil, err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return nil, err
		}

		imgURLs = append(imgURLs, dstPath)
	}
	return imgURLs, nil
}

func (s *LocalStorage) DeletePostImages(ctx context.Context, imgURLs []string) error {
	for _, imgURL := range imgURLs {
		filePath := strings.TrimPrefix(imgURL, "/")
		if _, err := os.Stat(filePath); err == nil {
			return os.Remove(filePath)
		}
	}
	return nil
}
