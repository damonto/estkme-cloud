package lpac

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	DownloadUrl = "https://github.com/estkme-group/lpac/releases/download/%s/%s"
)

var (
	ErrUnsupportedArch = errors.New("lpac does not currently have a binary file for this architecture, you must build it yourself from the source code")
)

func Download(dir, version string) error {
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	if !shouldDownload(dir, version) {
		slog.Info("lpac already downloaded", "version", version)
		return nil
	}
	return download(dir, version)
}

func packageName() string {
	switch runtime.GOOS {
	case "linux":
		switch runtime.GOARCH {
		case "amd64", "386":
			return "lpac-linux-x86_64.zip"
		default:
			return "lpac-linux-" + runtime.GOARCH + ".zip"
		}
	case "darwin":
		return "lpac-darwin-universal.zip"
	default:
		return ""
	}
}

func download(dir, version string) error {
	if err := setupDstDir(dir); err != nil {
		return err
	}

	path, err := downloadFile(fmt.Sprintf(DownloadUrl, version, packageName()), dir)
	if err != nil {
		return err
	}

	if err := unzip(path, dir); err != nil {
		return err
	}

	if _, err := os.Create(filepath.Join(dir, version)); err != nil {
		slog.Warn("failed to create version file", "version", version, "error", err)
	}

	if err := os.Remove(path); err != nil {
		slog.Warn("failed to remove zip file", "path", path, "error", err)
	}
	return nil
}

func unzip(filePath string, dir string) error {
	zipReader, err := zip.OpenReader(filePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		dst, err := os.OpenFile(filepath.Join(dir, f.Name), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		src, err := f.Open()
		if err != nil {
			return err
		}
		io.Copy(dst, src)
		src.Close()
		dst.Close()
	}
	return nil
}

func downloadFile(url string, dir string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return "", ErrUnsupportedArch
		}
		return "", errors.New("failed to download lpac: " + resp.Status)
	}
	defer resp.Body.Close()
	filePath := filepath.Join(dir, "lpac.zip")
	out, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return filePath, err
}

func setupDstDir(dir string) error {
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return nil
}

func shouldDownload(dir, version string) bool {
	versionFile := filepath.Join(dir, version)
	stat, err := os.Stat(versionFile)
	if os.IsNotExist(err) {
		return true
	}
	if stat.Name() != version {
		return true
	}
	return false
}
