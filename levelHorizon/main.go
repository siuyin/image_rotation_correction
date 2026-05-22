package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metadata struct {
	FPS             float64
	RefFrameIndex   int
	FrameInterval   int // Number of frames between outputs
}

func getVideoMetadata(videoPath string, intervalMs int) (Metadata, error) {
	// 1. Get FPS using ffprobe
	fpsCmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=r_frame_rate", "-of", "csv=p=0", videoPath)
	fpsOut, err := fpsCmd.Output()
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to get FPS: %v", err)
	}
	
	fpsParts := strings.Split(strings.TrimSpace(string(fpsOut)), "/")
	var fps float64
	if len(fpsParts) == 2 {
		num, _ := strconv.ParseFloat(fpsParts[0], 64)
		den, _ := strconv.ParseFloat(fpsParts[1], 64)
		fps = num / den
	} else {
		fps, _ = strconv.ParseFloat(fpsParts[0], 64)
	}

	// 2. Find first non-black frame in first 12 frames
	// Using signalstats to detect luminance (YAVG)
	refFrameIndex := 0
	signalCmd := exec.Command("ffprobe", "-v", "error", "-f", "lavfi",
		fmt.Sprintf("movie=%s,signalstats", videoPath),
		"-show_entries", "frame_tags=lavfi.signalstats.YAVG", "-of", "json", "-read_intervals", "%+#12")
	
	signalOut, err := signalCmd.Output()
	if err == nil {
		var result struct {
			Frames []struct {
				Tags struct {
					YAVG string `json:"lavfi.signalstats.YAVG"`
				} `json:"tags"`
			} `json:"frames"`
		}
		if err := json.Unmarshal(signalOut, &result); err == nil {
			for i, frame := range result.Frames {
				yavg, _ := strconv.ParseFloat(frame.Tags.YAVG, 64)
				if yavg > 20.0 {
					refFrameIndex = i
					break
				}
			}
		}
	}

	frameInterval := int(fps * float64(intervalMs) / 1000.0)
	if frameInterval < 1 {
		frameInterval = 1
	}

	return Metadata{
		FPS:           fps,
		RefFrameIndex: refFrameIndex,
		FrameInterval: frameInterval,
	}, nil
}

const (
	trfFile       = "transforms.trf"
	globalTrfFile = "global_motions.trf"
)

func runFFmpeg(args []string) error {
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type Config struct {
	VideoPath string
	Interval  int
}

func parseArgs() Config {
	videoPath := flag.String("i", "", "input video path")
	interval := flag.Int("m", 1000, "output interval in milliseconds")
	flag.Parse()
	if *videoPath == "" && flag.NArg() > 0 {
		*videoPath = flag.Arg(0)
	}
	if *videoPath == "" {
		fmt.Println("Usage: levelHorizon -i <video_path> [-m interval_ms]")
		os.Exit(1)
	}
	return Config{
		VideoPath: *videoPath,
		Interval:  *interval,
	}
}

func runDetectPass(videoPath string) {
	fmt.Println("Running pass 1: vidstabdetect...")
	args := []string{"-i", videoPath, "-vf", "vidstabdetect=tripod=1:result=" + trfFile, "-f", "null", "-"}
	if err := runFFmpeg(args); err != nil {
		fmt.Printf("Error in pass 1: %v\n", err)
		os.Exit(1)
	}
}

func runTransformPass(videoPath string) {
	fmt.Println("Running pass 2: vidstabtransform with debug=true...")
	args := []string{"-i", videoPath, "-vf", "vidstabtransform=input=" + trfFile + ":tripod=true:debug=true", "-f", "null", "-"}
	if err := runFFmpeg(args); err != nil {
		fmt.Printf("Error in pass 2: %v\n", err)
		os.Exit(1)
	}
}

func parseAndPrintResults() {
	file, err := os.Open(globalTrfFile)
	if err != nil {
		fmt.Printf("Error opening %s: %v\n", globalTrfFile, err)
		os.Exit(1)
	}
	defer file.Close()
	fmt.Println("\nRotation Correction Results (relative to frame 0):")
	fmt.Printf("%-10s %-20s\n", "Frame", "Correction Angle (deg)")
	processFile(os.Stdout, file)
}

func processFile(w io.Writer, file *os.File) {
	scanner := bufio.NewScanner(file)
	frameIdx := 0
	for scanner.Scan() {
		if angleDeg, ok := parseLine(scanner.Text()); ok {
			fmt.Fprintf(w, "%-10d %-20.4f\n", frameIdx, angleDeg)
			frameIdx++
		}
	}
}

func parseLine(line string) (float64, bool) {
	if strings.HasPrefix(line, "#") || strings.TrimSpace(line) == "" {
		return 0, false
	}
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return 0, false
	}
	angleRad, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return 0, false
	}
	return angleRad * 180.0 / 3.141592653589793, true
}

func main() {
	videoPath := parseArgs()
	defer os.Remove(trfFile)
	defer os.Remove(globalTrfFile)

	runDetectPass(videoPath)
	runTransformPass(videoPath)
	parseAndPrintResults()
}
