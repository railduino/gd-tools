package agent

import (
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	DownloadsName = "downloads.json"
)

type Download struct {
	Name     string `json:"name"`
	Releases string `json:"releases"`
	URL      string `json:"url"`
	MD5      string `json:"md5"`
	SHA256   string `json:"sha256"`
	SHA512   string `json:"sha512"`
	FileName string `json:"file_name"`
	DirName  string `json:"dir_name"`
	Binary   string `json:"binary"`
}

type DownloadList struct {
	Downloads []*Download
}

var (
	DownloadsDir string
)

func SetDownloadsDir(path string) {
	if path == "" {
		path = "/root/Downloads"
	}
	DownloadsDir = path
}

func GetDownloadsDir(name string) string {
	if DownloadsDir == "" {
		SetDownloadsDir("")
	}
	if name == "" {
		return DownloadsDir
	}
	return filepath.Join(DownloadsDir, name)
}

// This part is running on Dev
func GetDownload(name string) (*Download, error) {
	download, err := LoadDownload(name)
	if err != nil {
		return nil, err
	}

	if download.MD5 == "" && download.SHA256 == "" && download.SHA512 == "" {
		return nil, fmt.Errorf("unable to verify %s integrity", name)
	}

	return download, nil
}

func LoadDownload(name string) (*Download, error) {
	content, err := os.ReadFile(DownloadsName)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s file (%s): %w", DownloadsName, name, err)
	}

	var downloadList DownloadList
	if err := json.Unmarshal(content, &downloadList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal %s: %w", name, err)
	}

	for _, download := range downloadList.Downloads {
		if download.Name == name {
			return download, nil
		}
	}

	return nil, fmt.Errorf("failed to locate %s download", name)
}

// DownloadsTest checks if there is work to be done
func DownloadsTest(req *Request) bool {
	return req != nil && len(req.Downloads) > 0
}

func DownloadsHandler(req *Request, resp *Response) error {
	if req == nil || len(req.Downloads) == 0 || resp == nil {
		return nil
	}
	downloadsRoot := GetDownloadsDir("")

	if err := os.MkdirAll(downloadsRoot, 0755); err != nil {
		return err
	}

	for _, dwn := range req.Downloads {
		path := filepath.Join(downloadsRoot, dwn.FileName)
		status := fmt.Sprintf("✅ download %s", path)
		if _, err := os.Stat(path); err != nil {
			if !os.IsNotExist(err) {
				return err
			}
			if _, err := RunCommand("curl", "-fsSL", "-o", path, dwn.URL); err != nil {
				return err
			}
			status = fmt.Sprintf("download %s was successful", dwn.Name)
		}

		md5sum, sha256sum, sha512sum, err := computeHashes(path)
		if err != nil {
			return err
		}

		if dwn.MD5 != "" && dwn.MD5 != md5sum {
			return fmt.Errorf("MD5 mismatch for %s", dwn.FileName)
		}
		if dwn.SHA256 != "" && dwn.SHA256 != sha256sum {
			return fmt.Errorf("SHA256 mismatch for %s", dwn.FileName)
		}
		if dwn.SHA512 != "" && dwn.SHA512 != sha512sum {
			return fmt.Errorf("SHA512 mismatch for %s", dwn.FileName)
		}
		resp.Say(status)

		if dwn.Binary != "" {
			_, err := RunCommand("install",
				"-m", "0755",
				"-o", "root",
				"-g", "root",
				path,
				GetBinDir(dwn.Binary),
			)
			if err != nil {
				return err
			}
			resp.Sayf("copied %s to %s", path, GetBinDir(dwn.Binary))
		}
	}

	return nil
}

func computeHashes(path string) (string, string, string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", "", err
	}
	defer f.Close()

	hMd5 := md5.New()
	hSha256 := sha256.New()
	hSha512 := sha512.New()

	// Write simultaneously in all hashes
	if _, err := io.Copy(io.MultiWriter(hMd5, hSha256, hSha512), f); err != nil {
		return "", "", "", err
	}

	md5sum := hex.EncodeToString(hMd5.Sum(nil))
	sha256sum := hex.EncodeToString(hSha256.Sum(nil))
	sha512sum := hex.EncodeToString(hSha512.Sum(nil))

	return md5sum, sha256sum, sha512sum, nil
}
