package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	outDir = "output_frames"
	tgtW   = 640
)

type Probe struct {
	Streams []struct {
		W    int    `json:"width"`
		H    int    `json:"height"`
		Rate string `json:"avg_frame_rate"`
	} `json:"streams"`
}

func initStorage() {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("mkdir failed: %v\n", err)
		os.Exit(1)
	}
}

func parseRate(rate string) float64 {
	p := strings.Split(rate, "/")
	if len(p) != 2 {
		return 0
	}
	n, _ := strconv.ParseFloat(p[0], 64)
	d, _ := strconv.ParseFloat(p[1], 64)
	if d == 0 {
		return 0
	}
	return n / d
}

func getMeta(path string) (int, int, float64) {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0",
		"-show_entries", "stream=width,height,avg_frame_rate", "-of", "json", path)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("ffprobe failed: %v\n", err)
		os.Exit(1)
	}
	var b Probe
	json.Unmarshal(out, &b)
	if len(b.Streams) == 0 {
		fmt.Println("no stream"); os.Exit(1)
	}
	s := b.Streams[0]
	return s.W, s.H, parseRate(s.Rate)
}

func calcH(w, h int) int {
	if w <= 0 {
		return 0
	}
	return int(float64(tgtW) * float64(h) / float64(w)) &^ 1
}

func startFF(path string) (*exec.Cmd, io.ReadCloser) {
	cmd := exec.Command("ffmpeg", "-i", path, "-vf", fmt.Sprintf("scale=%d:-2", tgtW),
		"-f", "image2pipe", "-pix_fmt", "rgb24", "-vcodec", "rawvideo", "-")
	out, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("pipe failed: %v\n", err)
		os.Exit(1)
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("start failed: %v\n", err)
		os.Exit(1)
	}
	return cmd, out
}

func toImg(buf []byte, w, h int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si, di := (y*w+x)*3, (y*w+x)*4
			m.Pix[di] = buf[si]
			m.Pix[di+1] = buf[si+1]
			m.Pix[di+2] = buf[si+2]
			m.Pix[di+3] = 255
		}
	}
	return m
}

func save(m *image.RGBA, idx int, fps float64) {
	name := fmt.Sprintf("frame_offset_%08.3f.jpg", float64(idx)/fps)
	f, err := os.Create(filepath.Join(outDir, name))
	if err != nil {
		return
	}
	defer f.Close()
	jpeg.Encode(f, m, &jpeg.Options{Quality: 90})
}

func loop(out io.ReadCloser, h int, fps float64) {
	buf := make([]byte, tgtW*h*3)
	for i := 0; ; i++ {
		if _, err := io.ReadFull(out, buf); err != nil {
			break
		}
		save(toImg(buf, tgtW, h), i, fps)
		if i%100 == 0 {
			fmt.Printf("at %d\r", i)
		}
	}
}

func usage() {
	fmt.Printf("Usage: %s <video_path>\n", os.Args[0])
	fmt.Println("\nExtracts frames from a video and saves them as JPEGs.")
	fmt.Println("The output is stored in the 'output_frames' directory.")
	fmt.Println("\nOptions:")
	fmt.Println("  -h, --help  Show this help message")
}

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" {
		usage()
		return
	}
	initStorage()
	w, h, fps := getMeta(os.Args[1])
	th := calcH(w, h)
	cmd, out := startFF(os.Args[1])
	loop(out, th, fps)
	cmd.Wait()
	fmt.Println("\ndone")
}
