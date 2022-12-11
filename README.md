# Youtuberr

A hacky bad program that only exists because I didn't want to write it in BASH.

## Install

### Binaries

Without these two binaries, the program will panic on startup

`ffmpeg` - installed and on your PATH

`yt-dlp` - installed and on your PATH

### Files

`links.txt` - a file containing one youtube link per line, configurable via the `-input` flag

`cookies.txt` - a netscape cookies formatted copy of your cookies for youtube.com, configurable via the `-cookies` flag

## Command line flags

```
youtuberr -help
Usage of C:\Users\henry\Desktop\youtuberr\youtuberr.exe:
  -cookies string
        File to read cookies from (default "cookies.txt")
  -dir string
        Base directory to download videos to (default "./dl-youtuberr")
  -input string
        File to read links from (default "links.txt")
  -limit int
        Limit the number of videos to download per channel
```
