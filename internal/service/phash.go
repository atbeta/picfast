package service

import (
	"bytes"
	"image"
	"math"
	"sort"

	_ "image/jpeg"
	_ "image/png"
	_ "golang.org/x/image/webp"

	"golang.org/x/image/draw"
)

func ComputePHash(data []byte) (uint64, error) {
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return 0, err
	}

	gray := resampleGrayscale32x32(src)
	dct := compute2DDCT(gray)
	lowFreq := extractLowFreq(dct, 8)
	hash := dctHash(lowFreq)
	return hash, nil
}

func resampleGrayscale32x32(src image.Image) [][]float64 {
	const targetSize = 32
	dst := image.NewNRGBA(image.Rect(0, 0, targetSize, targetSize))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	result := make([][]float64, targetSize)
	for i := range result {
		result[i] = make([]float64, targetSize)
	}

	stride := dst.Stride
	pix := dst.Pix
	for y := 0; y < targetSize; y++ {
		row := result[y]
		offset := y * stride
		for x := 0; x < targetSize; x++ {
			i := offset + x*4
			row[x] = float64(pix[i])*0.299 + float64(pix[i+1])*0.587 + float64(pix[i+2])*0.114
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
