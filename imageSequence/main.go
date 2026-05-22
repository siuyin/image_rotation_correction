package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const outDir = "output_frames"

type Probe struct {
	Streams []struct {
		W int `json:"width"`
		H int `json:"height"`
	} `json:"streams"`
}

func initStorage() {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf("mkdir failed: %v\n", err)
		os.Exit(1)
	}
}

func getMeta(path string) (int, int) {
	p := []string{"-v", "error", "-select_streams", "v:0", "-show_entries", "stream=width,height", "-of", "json"}
	if strings.HasPrefix(path, "rtsp://") {
		p = append([]string{"-rtsp_transport", "tcp"}, p...)
	}
	out, _ := exec.Command("ffprobe", append(p, path)...).Output()
	var b Probe
	json.Unmarshal(out, &b)
	if len(b.Streams) == 0 {
		fmt.Println("no stream"); os.Exit(1)
	}
	return b.Streams[0].W, b.Streams[0].H
}

func calcH(w, h, tgtW int) int {
	if w <= 0 {
		return 0
	}
	return int(float64(tgtW) * float64(h) / float64(w)) &^ 1
}

func isStream(p string) bool {
	return strings.HasPrefix(p, "rtsp://") || strings.HasPrefix(p, "http://") ||
		strings.HasPrefix(p, "https://") || strings.HasPrefix(p, "rtmp://")
}

func startFF(path string, w int, fps float64) (*exec.Cmd, io.ReadCloser) {
	vf := fmt.Sprintf("fps=%f,scale=%d:-2:flags=neighbor", fps, w)
	args := []string{"-i", path, "-vf", vf, "-f", "image2pipe", "-pix_fmt", "rgb24", "-vcodec", "rawvideo", "-sws_flags", "neighbor", "-threads", "0", "-"}
	if isStream(path) {
		args = append([]string{"-fflags", "nobuffer+discardcorrupt", "-flags", "low_delay"}, args...)
	}
	if strings.HasPrefix(path, "rtsp://") {
		args = append([]string{"-rtsp_transport", "tcp"}, args...)
	}
	cmd := exec.Command("ffmpeg", args...)
	out, _ := cmd.StdoutPipe()
	cmd.Start()
	return cmd, out
}

func toImg(buf []byte, w, h int) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			si, di := (y*w+x)*3, (y*w+x)*4
			m.Pix[di], m.Pix[di+1], m.Pix[di+2], m.Pix[di+3] = buf[si], buf[si+1], buf[si+2], 255
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
	jpeg.Encode(f, m, &jpeg.Options{Quality: 50})
}

func loop(out io.ReadCloser, w, h int, fps float64) {
	buf := make([]byte, w*h*3)
	for i := 0; ; i++ {
		if _, err := io.ReadFull(out, buf); err != nil {
			break
		}
		save(toImg(buf, w, h), i, fps)
		if i%10 == 0 {
			fmt.Printf("at frame %d (%.2fs)\r", i, float64(i)/fps)
		}
	}
}

func usage() {
	fmt.Printf("Usage: %s [options] <video_path>\n", os.Args[0])
	fmt.Println("\nExtracts frames from a video and saves them as JPEGs.")
	fmt.Println("The output is stored in the 'output_frames' directory.")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	w := flag.Int("w", 640, "output image width")
	ms := flag.Int("m", 1000, "step interval in milliseconds")
	flag.Parse()
	if flag.NArg() < 1 {
		usage()
		return
	}
	initStorage()
	vW, vH := getMeta(flag.Arg(0))
	th, tFPS := calcH(vW, vH, *w), 1000.0/float64(*ms)
	cmd, out := startFF(flag.Arg(0), *w, tFPS)
	loop(out, *w, th, tFPS)
	cmd.Wait()
	fmt.Println("\ndone")
}
