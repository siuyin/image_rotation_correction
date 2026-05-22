package main

import (
	"os"
	"os/exec"
	"testing"
)

// generateTestVideo creates a 1-second 320x240 test video at 10fps
func generateTestVideo(t *testing.T, path string) {
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=10", "-t", "1", "-vcodec", "libx264", path)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate test video: %v", err)
	}
}

func TestParseRate(t *testing.T) {
	tests := []struct {
		in  string
		exp float64
	}{
		{"30/1", 30.0},
		{"60000/1001", 59.94005994005994},
		{"invalid", 0.0},
		{"30/0", 0.0},
	}
	for _, tt := range tests {
		res := parseRate(tt.in)
		if res != tt.exp && (res-tt.exp) > 0.001 {
			t.Errorf("parseRate(%s) = %f; want %f", tt.in, res, tt.exp)
		}
	}
}

func TestCalcH(t *testing.T) {
	tests := []struct {
		w, h, exp int
	}{
		{1920, 1080, 360}, // 640 * 1080 / 1920 = 360
		{1280, 720, 360},  // 640 * 720 / 1280 = 360
		{0, 720, 0},
		{100, 105, 672},   // (640 * 105 / 100) = 672
	}
	for _, tt := range tests {
		res := calcH(tt.w, tt.h)
		if res != tt.exp {
			t.Errorf("calcH(%d, %d) = %d; want %d", tt.w, tt.h, res, tt.exp)
		}
	}
}

func TestToImg(t *testing.T) {
	w, h := 2, 2
	buf := make([]byte, w*h*3)
	for i := range buf {
		buf[i] = byte(i)
	}
	m := toImg(buf, w, h)
	if m.Rect.Dx() != w || m.Rect.Dy() != h {
		t.Errorf("wrong image dimensions: %vx%v", m.Rect.Dx(), m.Rect.Dy())
	}
	// Check first pixel (RGB)
	if m.Pix[0] != 0 || m.Pix[1] != 1 || m.Pix[2] != 2 || m.Pix[3] != 255 {
		t.Errorf("wrong pixel values: %v", m.Pix[0:4])
	}
}

func TestGetMeta(t *testing.T) {
	path := "test_meta.mp4"
	generateTestVideo(t, path)
	defer os.Remove(path)

	w, h, fps := getMeta(path)
	if w != 320 || h != 240 || fps != 10.0 {
		t.Errorf("getMeta failed: %dx%d @ %f", w, h, fps)
	}
}

func TestIntegration(t *testing.T) {
	path := "test_int.mp4"
	generateTestVideo(t, path)
	defer os.Remove(path)
	
	// Clean output dir for test
	os.RemoveAll(outDir)
	initStorage()
	defer os.RemoveAll(outDir)

	w, h, fps := getMeta(path)
	th := calcH(w, h)
	cmd, out := startFF(path)
	loop(out, th, fps)
	cmd.Wait()

	files, _ := os.ReadDir(outDir)
	if len(files) == 0 {
		t.Error("no frames extracted")
	}
}
