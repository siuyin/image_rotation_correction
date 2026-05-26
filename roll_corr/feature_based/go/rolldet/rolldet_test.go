package rolldet

import (
	"fmt"
	"math"
	"testing"

	"gocv.io/x/gocv"
)

func ExampleDecomposeTransformation() {
	// Construct a 90-degree rotation matrix
	mapping := gocv.NewMatWithSize(3, 3, gocv.MatTypeCV64F)
	defer mapping.Close()
	mapping.SetDoubleAt(0, 0, 0.0)
	mapping.SetDoubleAt(0, 1, -1.0)
	mapping.SetDoubleAt(1, 0, 1.0)
	mapping.SetDoubleAt(1, 1, 0.0)
	mapping.SetDoubleAt(2, 2, 1.0)

	params := DecomposeTransformation(mapping)
	fmt.Printf("Roll: %.1f deg, Zoom: %.1f\n", params.Roll, params.Zoom)
	// Output:
	// Roll: 90.0 deg, Zoom: 1.0
}

func TestMatchFeatures(t *testing.T) {
	// Create images with random content to ensure features are detectable
	img1 := gocv.NewMatWithSize(100, 100, gocv.MatTypeCV8U)
	gocv.RandU(&img1, gocv.NewScalar(0, 0, 0, 0), gocv.NewScalar(255, 255, 255, 255))
	defer img1.Close()
	img2 := img1.Clone()
	defer img2.Close()

	orb := gocv.NewORB()
	defer orb.Close()
	_, desc1 := orb.DetectAndCompute(img1, gocv.NewMat())
	_, desc2 := orb.DetectAndCompute(img2, gocv.NewMat())
	defer desc1.Close()
	defer desc2.Close()

	matches := MatchFeatures(desc1, desc2)
	// Since images are identical, we should get matches
	if len(matches) == 0 {
		t.Errorf("Expected matches, got 0")
	}
}

func TestDecomposeTransformation(t *testing.T) {
	// Identity matrix
	mapping := gocv.NewMatWithSize(3, 3, gocv.MatTypeCV64F)
	mapping.SetDoubleAt(0, 0, 1.0)
	mapping.SetDoubleAt(0, 1, 0.0)
	mapping.SetDoubleAt(1, 0, 0.0)
	mapping.SetDoubleAt(1, 1, 1.0)
	mapping.SetDoubleAt(2, 2, 1.0)
	defer mapping.Close()

	params := DecomposeTransformation(mapping)
	if params.Roll != 0 {
		t.Errorf("Expected roll 0, got %f", params.Roll)
	}
	if math.Abs(params.Zoom-1.0) > 0.001 {
		t.Errorf("Expected zoom 1, got %f", params.Zoom)
	}
}

func TestRollComputation(t *testing.T) {
	// Construct a 90-degree rotation matrix:
	// [ 0 -1  0 ]
	// [ 1  0  0 ]
	// [ 0  0  1 ]
	mapping := gocv.NewMatWithSize(3, 3, gocv.MatTypeCV64F)
	mapping.SetDoubleAt(0, 0, 0.0)
	mapping.SetDoubleAt(0, 1, -1.0)
	mapping.SetDoubleAt(1, 0, 1.0)
	mapping.SetDoubleAt(1, 1, 0.0)
	mapping.SetDoubleAt(2, 2, 1.0)
	defer mapping.Close()

	params := DecomposeTransformation(mapping)
	expectedRoll := 90.0
	if math.Abs(params.Roll-expectedRoll) > 0.01 {
		t.Errorf("Expected roll %f, got %f", expectedRoll, params.Roll)
	}
}
