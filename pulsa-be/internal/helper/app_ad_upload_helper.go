package helper

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const AppAdUploadDir = "uploads/app-ads"

func SaveAppAdUpload(file multipart.File, header *multipart.FileHeader) (string, error) {
	if file == nil || header == nil {
		return "", fmt.Errorf("file upload tidak valid")
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
	default:
		return "", fmt.Errorf("format gambar harus jpg, jpeg, png, atau webp")
	}

	if err := os.MkdirAll(AppAdUploadDir, 0o755); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("app-ad-%d%s", time.Now().UnixNano(), ext)
	targetPath := filepath.Join(AppAdUploadDir, filename)

	dst, err := os.Create(targetPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return "/uploads/app-ads/" + filename, nil
}

func DeleteManagedAppAdUpload(imageURL string) error {
	path := ResolveManagedAppAdUploadPath(imageURL)
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func ResolveManagedAppAdUploadPath(imageURL string) string {
	imageURL = strings.TrimSpace(imageURL)
	if imageURL == "" {
		return ""
	}

	marker := "/uploads/app-ads/"
	idx := strings.Index(imageURL, marker)
	if idx == -1 {
		return ""
	}

	rel := strings.TrimPrefix(imageURL[idx:], "/")
	clean := filepath.Clean(rel)
	if clean == "." || strings.HasPrefix(clean, "..") {
		return ""
	}
	return clean
}
