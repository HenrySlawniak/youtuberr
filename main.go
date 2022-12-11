package main

import (
	"bufio"
	"crypto/md5"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

var (
	links []string
	wg    = sync.WaitGroup{}

	limit           = flag.Int("limit", 0, "Limit the number of videos to download per channel")
	baseDownloadDir = flag.String("dir", "./dl-youtuberr", "Base directory to download videos to")
	inputFile       = flag.String("input", "links.txt", "File to read links from")
	cookiesFile     = flag.String("cookies", "cookies.txt", "File to read cookies from")
)

func init() {
	flag.Parse()
}

func main() {
	err := ensureBinaries()
	if err != nil {
		panic(err)
	}

	err = loadLinks()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Loaded %d links\n", len(links))

	for _, link := range links {
		wg.Add(1)
		go runLinkDownload(link, *limit)
	}

	wg.Wait()
}

func ensureBinaries() error {
	cmd := exec.Command("ffmpeg", "-version")
	err := cmd.Run()
	if err != nil {
		return err
	}

	cmd = exec.Command("yt-dlp", "--version")
	err = cmd.Run()
	if err != nil {
		return err
	} else {
		cmd = exec.Command("yt-dlp", "-U")
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout

		err = cmd.Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func loadLinks() error {
	f, err := os.Open(*inputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	links = []string{}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		links = append(links, scanner.Text())
	}

	return nil
}

func runLinkDownload(link string, limit int) {
	defer wg.Done()

	hasher := md5.New()
	hasher.Write([]byte(link))
	hash := hasher.Sum(nil)

	archiveName := fmt.Sprintf("dl-%x.txt", hash)

	fmt.Printf("%s => %s\n", archiveName, link)

	cmd := exec.Command(
		"yt-dlp",
		"-i",
		"--restrict-filenames",
		"--no-progress",
		"--add-metadata",
		"--write-info-json",
		"--write-description",
		"--write-playlist-metafiles",
		"--video-multistreams",
		"--audio-multistreams",
		"--write-subs",
		"--embed-subs",
		"--embed-chapters",
		"--embed-info-json",
		"--sub-format", "best",
		"--sub-langs", "all",
		"--remux-video", "mkv",
		"--cookies", *cookiesFile,
		"-f", "bestvideo+bestaudio",
		"--download-archive", filepath.Join(*baseDownloadDir, archiveName),
		"--limit-rate", "2M",
		"--max-downloads", fmt.Sprintf("%d", limit),
		"-o", filepath.Join(*baseDownloadDir, "%(uploader)s - %(channel_id)s/%(upload_date)s - %(title)s-%(id)s.%(ext)s"),
		link,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", link, err)
	}
}
