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

func getFPS(videoPath string) float64 {
	cmd := exec.Command("ffprobe", "-v", "error", "-select_streams", "v:0", "-show_entries", "stream=r_frame_rate", "-of", "csv=p=0", videoPath)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("failed to get FPS: %v\n", err)
		os.Exit(1)
	}
	parts := strings.Split(strings.TrimSpace(string(out)), "/")
	if len(parts) == 2 {
		num, _ := strconv.ParseFloat(parts[0], 64)
		den, _ := strconv.ParseFloat(parts[1], 64)
		return num / den
	}
	fps, _ := strconv.ParseFloat(parts[0], 64)
	return fps
}

func getRefFrameIndex(videoPath string) int {
	cmd := exec.Command("ffprobe", "-v", "error", "-f", "lavfi",
		fmt.Sprintf("movie=%s,signalstats", videoPath),
		"-show_entries", "frame_tags=lavfi.signalstats.YAVG", "-of", "json", "-read_intervals", "%+#12")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var result struct {
		Frames []struct {
			Tags struct {
				YAVG string `json:"lavfi.signalstats.YAVG"`
			} `json:"tags"`
		} `json:"frames"`
	}
	if err := json.Unmarshal(out, &result); err == nil {
		for i, frame := range result.Frames {
			yavg, _ := strconv.ParseFloat(frame.Tags.YAVG, 64)
			if yavg > 20.0 {
				return i
			}
		}
	}
	return 0
}

func mustGetVideoMetadata(videoPath string, intervalMs int) Metadata {
	fps := getFPS(videoPath)
	refIdx := getRefFrameIndex(videoPath)
	frameInterval := int(fps * float64(intervalMs) / 1000.0)
	if frameInterval < 1 {
		frameInterval = 1
	}
	return Metadata{FPS: fps, RefFrameIndex: refIdx, FrameInterval: frameInterval}
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
	interval := flag.Int("m", 1000, "output interval in milliseconds")
	flag.Parse()
	if flag.NArg() == 0 {
		fmt.Println("Usage: levelHorizon <video_path> [-m interval_ms]")
		os.Exit(1)
	}
	return Config{
		VideoPath: flag.Arg(0),
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

func runTransformPass(videoPath string, meta Metadata) {
	fmt.Println("Running pass 2: vidstabtransform with streaming...")

	args := []string{"-i", videoPath, "-vf", "vidstabtransform=input=" + trfFile + ":tripod=true:debug=true", "-f", "null", "-"}
	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Start(); err != nil {
		fmt.Printf("Error starting pass 2: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nRotation Correction Results (relative to frame", meta.RefFrameIndex, "):")
	fmt.Printf("%-10s %-20s\n", "Frame", "Correction Angle (deg)")

	streamResults(meta)

	if err := cmd.Wait(); err != nil {
		fmt.Printf("Error in pass 2: %v\n", err)
		os.Exit(1)
	}
}

func waitForFile(filename string) {
	for {
		if _, err := os.Stat(filename); err == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func outputResult(frameIdx int, angle float64, refAngle float64, meta Metadata) {
	if frameIdx%meta.FrameInterval == 0 {
		fmt.Printf("%-10d %-20.4f\n", frameIdx, angle-refAngle)
	}
}

func processFrames(reader *bufio.Reader, meta Metadata) {
	var refAngle float64
	refAngleSet := false
	frameIdx := 0
	var angleBuffer []float64

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				time.Sleep(100 * time.Millisecond)
				continue
			}
			break
		}
		angleDeg, ok := parseLine(line)
		if !ok {
			continue
		}
		if frameIdx == meta.RefFrameIndex {
			refAngle = angleDeg
			refAngleSet = true
			for i, bufAngle := range angleBuffer {
				outputResult(i, bufAngle, refAngle, meta)
			}
		}
		if !refAngleSet {
			angleBuffer = append(angleBuffer, angleDeg)
		} else {
			outputResult(frameIdx, angleDeg, refAngle, meta)
		}
		frameIdx++
	}
}

func streamResults(meta Metadata) {
	waitForFile(globalTrfFile)
	
	// Use tail -f to stream the file
	cmd := exec.Command("tail", "-f", globalTrfFile)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err := cmd.Start(); err != nil {
		return
	}
	defer cmd.Process.Kill()
	
	processFrames(bufio.NewReader(stdout), meta)
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
	config := parseArgs()
	defer os.Remove(trfFile)
	defer os.Remove(globalTrfFile)

	runDetectPass(config.VideoPath)
	meta := mustGetVideoMetadata(config.VideoPath, config.Interval)
	runTransformPass(config.VideoPath, meta)
}
