package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	trfFile       = "transforms.trf"
	globalTrfFile = "global_motions.trf"
)

func runFFmpeg(args []string) error {
	cmd := exec.Command("ffmpeg", args...)
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func parseArgs() string {
	videoPath := flag.String("i", "", "input video path")
	flag.Parse()
	if *videoPath == "" && flag.NArg() > 0 {
		*videoPath = flag.Arg(0)
	}
	if *videoPath == "" {
		fmt.Println("Usage: levelHorizon -i <video_path>")
		os.Exit(1)
	}
	return *videoPath
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
