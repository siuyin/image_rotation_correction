package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gocv.io/x/gocv"
	"roll_corr/feature_based/go/rolldet"
)

type Config struct {
	Interval int
	Area     float64
	Video    string
}

type Processor struct {
	cap     *gocv.VideoCapture
	cfg     Config
	step    int
	fps     float64
	refKp   []gocv.KeyPoint
	refDesc gocv.Mat
}

func main() {
	cfg := parseConfig()
	cap := openVideo(cfg.Video)
	defer cap.Close()
	processVideo(cap, cfg)
}

func parseConfig() Config {
	interval := flag.Int("i", 1000, "Sampling interval in ms")
	area := flag.Float64("a", 75.0, "Central area percentage")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Println("Usage: roll_corr VIDEO_PATH [-i INTERVAL_MS] [-a AREA_PERCENT]")
		os.Exit(1)
	}
	return Config{*interval, *area, flag.Arg(0)}
}

func openVideo(path string) *gocv.VideoCapture {
	cap, err := gocv.VideoCaptureFile(path)
	if err != nil || !cap.IsOpened() {
		log.Fatalf("Error opening video: %v", err)
	}
	return cap
}

func processVideo(cap *gocv.VideoCapture, cfg Config) {
	fps := cap.Get(gocv.VideoCaptureFPS)
	step := int(float64(cfg.Interval) * fps / 1000.0)
	if step < 1 { step = 1 }

	frame := gocv.NewMat()
	defer frame.Close()
	cap.Read(&frame)
	refKp, refDesc, refCentral := initializeReference(frame, cfg.Area)
	defer refCentral.Close()
	defer refDesc.Close()

	p := Processor{cap, cfg, step, fps, refKp, refDesc}
	p.run()
}

func initializeReference(frame gocv.Mat, area float64) ([]gocv.KeyPoint, gocv.Mat, gocv.Mat) {
	gray := gocv.NewMat()
	defer gray.Close()
	gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)
	refCentral := rolldet.GetCentralArea(gray, area)
	kp, desc := rolldet.DetectAndDescribe(refCentral)
	return kp, desc, refCentral
}

func (p *Processor) run() {
	gray, central, desc, mapping := gocv.NewMat(), gocv.NewMat(), gocv.NewMat(), gocv.NewMat()
	defer func() { gray.Close(); central.Close(); desc.Close(); mapping.Close() }()
	frame := gocv.NewMat()
	defer frame.Close()

	fmt.Println("Virtual Horizon roll correction:")
	idx := 0
	for p.cap.Read(&frame) {
		idx++
		if idx%p.step != 0 { continue }
		gocv.CvtColor(frame, &gray, gocv.ColorBGRToGray)
		central = rolldet.GetCentralArea(gray, p.cfg.Area)
		kp, desc := rolldet.DetectAndDescribe(central)
		matches := rolldet.MatchFeatures(p.refDesc, desc)
		mapping = rolldet.SolveGeometricMapping(p.refKp, kp, matches)
		params := rolldet.DecomposeTransformation(mapping)
		fmt.Println(fmt.Sprintf("%.2fs:\t%.2f°", float64(idx)/p.fps, params.Roll))
		central.Close()
		desc.Close()
		mapping.Close()
	}
}
