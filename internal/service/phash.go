package service

import (
	"bytes"
	"image"
	"math"
	"sort"

	_ "image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"
)

func ComputePHash(data []byte) (uint64, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}

	gray := toGrayscale32x32(img)
	dct := compute2DDCT(gray)
	lowFreq := extractLowFreq(dct, 8)
	hash := dctHash(lowFreq)
	return hash, nil
}

func toGrayscale32x32(img image.Image) [][]float64 {
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()

	const targetSize = 32
	result := make([][]float64, targetSize)
	for i := range result {
		result[i] = make([]float64, targetSize)
	}

	for y := 0; y < targetSize; y++ {
		for x := 0; x < targetSize; x++ {
			srcX := x * srcW / targetSize
			srcY := y * srcH / targetSize
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			result[y][x] = float64(r>>8)*0.299 + float64(g>>8)*0.587 + float64(b>>8)*0.114
		}
	}

	return result
}

func compute2DDCT(matrix [][]float64) [][]float64 {
	n := len(matrix)
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	for u := 0; u < n; u++ {
		cu := 1.0
		if u == 0 {
			cu = 1.0 / math.Sqrt2
		}
		for v := 0; v < n; v++ {
			cv := 1.0
			if v == 0 {
				cv = 1.0 / math.Sqrt2
			}
			var sum float64
			for x := 0; x < n; x++ {
				for y := 0; y < n; y++ {
					sum += matrix[y][x] *
						math.Cos((2*float64(x)+1)*float64(u)*math.Pi/(2*float64(n))) *
						math.Cos((2*float64(y)+1)*float64(v)*math.Pi/(2*float64(n)))
				}
			}
			result[v][u] = cu * cv * sum * 2 / float64(n)
		}
	}

	return result
}

func extractLowFreq(dct [][]float64, size int) []float64 {
	result := make([]float64, 0, size*size)
	for y := range size {
		for x := range size {
			result = append(result, dct[y][x])
		}
	}
	return result
}

func dctHash(values []float64) uint64 {
	if len(values) == 0 {
		return 0
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	median := sorted[len(sorted)/2]

	var hash uint64
	for i, v := range values {
		if v > median {
			hash |= 1 << uint(i)
		}
	}
	return hash
}
