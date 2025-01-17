package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/eikendev/hackenv/internal/images"
	"github.com/eikendev/hackenv/internal/settings"
	progressbar "github.com/schollz/progressbar/v3"
)

type GetCommand struct {
	Force  bool `short:"f" long:"force" description:"Force to download the new image"`
	Update bool `short:"u" long:"update" description:"Allow update to the latest image"`
}

func (c *GetCommand) Execute(args []string) error {
	settings.Runner = c
	return nil
}

// https://golang.org/pkg/crypto/sha256/#example_New_file
func calculateFileChecksum(path string) string {
	log.Printf("Calculating checksum of %s\n", path)

	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open file: %s\n", err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		log.Fatalf("Failed to copy file content: %s\n", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// https://stackoverflow.com/a/11693049
func downloadImage(path, url string) {
	log.Printf("Downloading image to %s\n", path)

	out, err := os.Create(path)
	if err != nil {
		log.Fatalf("Cannot write image file: %s\n", err)
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("Cannot download image file: %s\n", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Cannot download image file: bad status %s\n", resp.Status)
	}

	bar := progressbar.DefaultBytes(
		resp.ContentLength,
		"downloading",
	)

	_, err = io.Copy(io.MultiWriter(out, bar), resp.Body)
	if err != nil {
		log.Fatalf("Cannot write image file: %s\n", err)
	}

	log.Println("Download successful")
}

func validateChecksum(localPath, checksum string) {
	if newChecksum := calculateFileChecksum(localPath); newChecksum != checksum {
		checksumMsg := fmt.Sprintf("Downloaded image has bad checksum: %s instead of %s", newChecksum, checksum)

		err := os.Remove(localPath)
		if err != nil {
			log.Fatalf("%s. Unable to remove file.\n", checksumMsg)
		}

		log.Fatalf("%s. File removed.\n", checksumMsg)
	}

	log.Println("Checksum validated successfully")
}

func (c *GetCommand) Run(s *settings.Settings) {
	image := images.GetImageDetails(s.Type)
	info := image.GetDownloadInfo(true)

	log.Printf("Found file %s with checksum %s\n", info.Filename, info.Checksum)

	localPath := image.GetLocalPath(info.Version)

	// https://stackoverflow.com/a/12518877
	if _, err := os.Stat(localPath); err == nil {
		// The image already exists.

		if !c.Update && !c.Force {
			log.Println("An image is already installed; update with --update")
			return
		}

		localVersion := image.FileVersion(localPath)

		if !c.Force && image.VersionComparer.Eq(info.Version, localVersion) {
			log.Println("Latest image is already installed; force with --force")
			return
		}
	} else if !os.IsNotExist(err) {
		log.Fatalf("Unable to get file information for path %s\n", localPath)
	}

	downloadImage(localPath, image.ArchiveURL+"/"+info.Filename)

	validateChecksum(localPath, info.Checksum)
}
