package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/xerrors"
)

var (
	links []string
	wg    = sync.WaitGroup{}
	tick  time.Duration

	limit           = flag.Int("limit", 0, "Limit the number of videos to download per channel")
	baseDownloadDir = flag.String("dir", "./dl-youtuberr", "Base directory to download videos to")
	inputFile       = flag.String("input", "links.txt", "File to read links from")
	cookiesFile     = flag.String("cookies", "cookies.txt", "File to read cookies from")
	listMode        = flag.String("list-mode", "serial", "Mode to run in. parallel, or serial.")
	once            = flag.Bool("run-once", false, "Run once and exit")
	tickerDuration  = flag.String("ticker", "1h", "Duration to wait between runs")
	limitRate       = flag.String("limit-rate", "2M", "Limit the download rate to this value. 0 for no limit.")
)

func init() {
	flag.Parse()
	var err error
	tick, err = time.ParseDuration(*tickerDuration)
	if err != nil {
		panic(xerrors.Errorf("failed to parse ticker duration: %w", err))
	}

	tick = time.Hour
}

func main() {
	err := ensureBinaries()
	if err != nil {
		panic(xerrors.Errorf("failed to ensure binaries: %w", err))
	}

	err = loadLinks()
	if err != nil {
		panic(xerrors.Errorf("failed to load links: %w", err))
	}

	fmt.Printf("Loaded %d links\n", len(links))

	if *once {
		work()
	} else {
		ticker := time.NewTicker(tick)
		work()

		for {
			select {
			case <-ticker.C:
				work()
			}
		}
	}

}

func work() {
	switch *listMode {
	case "parallel":
		runParallel()
	case "serial":
		runSerial()
	default:
		fmt.Printf("Invalid list mode, %s is not defined, defaulting to serial\n", *listMode)
		runSerial()
	}
}

func runSerial() {
	for _, link := range links {
		runLinkDownload(link, *limit)
	}
}

func runParallel() {
	for _, link := range links {
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

const archiveName = "youtuberr.txt"

func runLinkDownload(link string, limit int) {
	wg.Add(1)
	defer wg.Done()

	fmt.Printf("%s => %s\n", archiveName, link)

	progressOption := "--no-progress"
	if *listMode == "serial" || len(links) == 1 {
		progressOption = "--progress"
	}

	cmd := exec.Command(
		"yt-dlp",
		"-i",
		"--restrict-filenames",
		progressOption,
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
		"--limit-rate", *limitRate,
		"-P", "temp:.temp",
		"--max-downloads", fmt.Sprintf("%d", limit),
		"-o", filepath.Join(*baseDownloadDir, "%(uploader)s [%(channel_id)s]/%(upload_date)s - %(title)s [%(id)s].%(ext)s"),
		link,
	)

	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error downloading %s: %v\n", link, err)
	}
}
