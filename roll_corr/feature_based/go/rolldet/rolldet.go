package rolldet

import (
	"image"
	"math"

	"gocv.io/x/gocv"
)

type TransformationParams struct {
	Pitch, Yaw, Zoom, Roll float64
}

func GetCentralArea(img gocv.Mat, pArea float64) gocv.Mat {
	h, w := img.Rows(), img.Cols()
	ratio := math.Sqrt(pArea / 100.0)
	newW, newH := int(float64(w)*ratio), int(float64(h)*ratio)
	rect := image.Rect((w-newW)/2, (h-newH)/2, (w+newW)/2, (h+newH)/2)
	return img.Region(rect)
}

func DetectAndDescribe(img gocv.Mat) ([]gocv.KeyPoint, gocv.Mat) {
	orb := gocv.NewORB()
	defer orb.Close()
	return orb.DetectAndCompute(img, gocv.NewMat())
}

func MatchFeatures(desc1, desc2 gocv.Mat) []gocv.DMatch {
	matcher := gocv.NewBFMatcherWithParams(gocv.NormHamming, false)
	defer matcher.Close()

	knnMatches := matcher.KnnMatch(desc1, desc2, 2)

	var goodMatches []gocv.DMatch
	for _, match := range knnMatches {
		if len(match) >= 2 && match[0].Distance < 0.75*match[1].Distance {
			goodMatches = append(goodMatches, match[0])
		}
	}
	return goodMatches
}

func SolveGeometricMapping(kp1, kp2 []gocv.KeyPoint, matches []gocv.DMatch) gocv.Mat {
	if len(matches) < 4 {
		return gocv.NewMat()
	}

	srcPts := make([]image.Point, len(matches))
	dstPts := make([]image.Point, len(matches))
	for i, m := range matches {
		srcPts[i] = image.Pt(int(kp1[m.QueryIdx].X), int(kp1[m.QueryIdx].Y))
		dstPts[i] = image.Pt(int(kp2[m.TrainIdx].X), int(kp2[m.TrainIdx].Y))
	}
	
	srcVec := gocv.NewPointVectorFromPoints(srcPts)
	defer srcVec.Close()
	dstVec := gocv.NewPointVectorFromPoints(dstPts)
	defer dstVec.Close()

	srcMat := gocv.NewMatFromPointVector(srcVec, false)
	defer srcMat.Close()
	dstMat := gocv.NewMatFromPointVector(dstVec, false)
	defer dstMat.Close()

	mask := gocv.NewMat()
	defer mask.Close()

	return gocv.FindHomography(srcMat, dstMat, gocv.HomographyMethodRANSAC, 3.0, &mask, 2000, 0.995)
}

func DecomposeTransformation(mapping gocv.Mat) TransformationParams {
	if mapping.Empty() {
		return TransformationParams{0, 0, 1, 0}
	}

	h11 := mapping.GetDoubleAt(0, 0)
	h21 := mapping.GetDoubleAt(1, 0)
	
	roll := math.Atan2(h21, h11) * (180.0 / math.Pi)

	return TransformationParams{
		Pitch: 0,
		Yaw:   0,
		Zoom:  math.Sqrt(h11*h11 + h21*h21), 
		Roll:  roll,
	}
}
