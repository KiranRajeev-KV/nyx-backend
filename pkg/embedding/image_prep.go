package embedding

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"net/http"
	"runtime"
)

var (
	mean = []float32{0.48145466, 0.4578275, 0.40821073}
	std  = []float32{0.26862954, 0.26130258, 0.27577711}
	dim  = 224
)

type ImagePreprocessor struct{}

func NewImagePreprocessor() *ImagePreprocessor {
	return &ImagePreprocessor{}
}

func (p *ImagePreprocessor) PreprocessFromURL(imageURL string) ([]float32, error) {
	img, err := p.DownloadAndDecodeImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	return p.Preprocess(img), nil
}

func (p *ImagePreprocessor) PreprocessFromBase64(base64Data string) ([]float32, error) {
	img, err := p.DecodeBase64Image(base64Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return p.Preprocess(img), nil
}

func (p *ImagePreprocessor) DownloadAndDecodeImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	img, err := jpeg.Decode(resp.Body)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (p *ImagePreprocessor) DecodeBase64Image(base64Data string) (image.Image, error) {
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}

	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (p *ImagePreprocessor) Preprocess(img image.Image) []float32 {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	var resized image.Image
	if w < h {
		newH := (h * dim) / w
		resized = resizeImage(img, dim, newH)
	} else {
		newW := (w * dim) / h
		resized = resizeImage(img, newW, dim)
	}

	cropped := centerCrop(resized, dim, dim)

	return normalizeImage(cropped)
}

func resizeImage(img image.Image, w, h int) image.Image {
	result := image.NewRGBA(image.Rect(0, 0, w, h))

	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	runtime.GOMAXPROCS(runtime.NumCPU())

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcX := x * srcW / w
			srcY := y * srcH / h
			result.Set(x, y, img.At(srcX+srcBounds.Min.X, srcY+srcBounds.Min.Y))
		}
	}

	return result
}

func centerCrop(img image.Image, w, h int) image.Image {
	bounds := img.Bounds()
	imgW := bounds.Dx()
	imgH := bounds.Dy()

	x0 := (imgW - w) / 2
	y0 := (imgH - h) / 2

	if x0 < 0 {
		x0 = 0
	}
	if y0 < 0 {
		y0 = 0
	}

	cropped := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cropped.Set(x, y, img.At(x+x0+bounds.Min.X, y+y0+bounds.Min.Y))
		}
	}

	return cropped
}

func normalizeImage(img image.Image) []float32 {
	result := make([]float32, 3*dim*dim)

	bounds := img.Bounds()

	for y := 0; y < dim; y++ {
		for x := 0; x < dim; x++ {
			r, g, b, _ := img.At(x+bounds.Min.X, y+bounds.Min.Y).RGBA()

			r8 := float32(r>>8) / 255.0
			g8 := float32(g>>8) / 255.0
			b8 := float32(b>>8) / 255.0

			idx := y*dim + x
			result[0*dim*dim+idx] = (r8 - mean[0]) / std[0]
			result[1*dim*dim+idx] = (g8 - mean[1]) / std[1]
			result[2*dim*dim+idx] = (b8 - mean[2]) / std[2]
		}
	}

	return hwcToChw(result)
}

func hwcToChw(input []float32) []float32 {
	result := make([]float32, len(input))
	size := dim * dim

	for i := 0; i < size; i++ {
		result[i] = input[i]
		result[size+i] = input[size+i]
		result[2*size+i] = input[2*size+i]
	}

	return result
}
