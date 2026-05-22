package main

import (
	"os"
	"os/exec"
	"testing"
)

func generateTestVideo(t *testing.T, path string) {
	cmd := exec.Command("ffmpeg", "-y", "-f", "lavfi", "-i", "testsrc=size=320x240:rate=10", "-t", "1", "-vcodec", "libx264", path)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate test video: %v", err)
	}
}

func TestCalcH(t *testing.T) {
	tests := []struct {
		w, h, tgtW, exp int
	}{
		{1920, 1080, 640, 360},
		{1280, 720, 640, 360},
		{0, 720, 640, 0},
		{100, 105, 640, 672},
	}
	for _, tt := range tests {
		res := calcH(tt.w, tt.h, tt.tgtW)
		if res != tt.exp {
			t.Errorf("calcH(%d, %d, %d) = %d; want %d", tt.w, tt.h, tt.tgtW, res, tt.exp)
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
	if m.Pix[0] != 0 || m.Pix[1] != 1 || m.Pix[2] != 2 || m.Pix[3] != 255 {
		t.Errorf("wrong pixel values: %v", m.Pix[0:4])
	}
}

func TestGetMeta(t *testing.T) {
	path := "test_meta.mp4"
	generateTestVideo(t, path)
	defer os.Remove(path)

	w, h := getMeta(path)
	if w != 320 || h != 240 {
		t.Errorf("getMeta failed: %dx%d", w, h)
	}
}

func TestIntegration(t *testing.T) {
	path := "test_int.mp4"
	generateTestVideo(t, path)
	defer os.Remove(path)
	
	os.RemoveAll(outDir)
	initStorage()
	defer os.RemoveAll(outDir)

	w, h := getMeta(path)
	tgtW, tFPS := 160, 2.0
	th := calcH(w, h, tgtW)
	cmd, out := startFF(path, tgtW, tFPS)
	loop(out, tgtW, th, tFPS)
	cmd.Wait()

	files, _ := os.ReadDir(outDir)
	if len(files) == 0 {
		t.Error("no frames extracted")
	}
}
