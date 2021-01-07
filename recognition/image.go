package recognition

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"os"
	"runtime"
)

type resamplingFilter struct {
	Support float64
	Kernel  func(float64) float64
}

func ResizeForMatrix(i interface{}, width int, height int) (imgMatrix [][][]uint8, err error) {
	img, err1 := DecodeImage(i)

	if err1 != nil {
		err = err1
		return
	}

	nrgba := convertToNRGBA(img)
	src := Resize(nrgba, width, height)

	imgMatrix = NewRGBAMatrix(height, width)

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			c := src.At(j, i)
			r, g, b, a := c.RGBA()
			imgMatrix[i][j][0] = uint8(r)
			imgMatrix[i][j][1] = uint8(g)
			imgMatrix[i][j][2] = uint8(b)
			imgMatrix[i][j][3] = uint8(a)

		}
	}

	return
}

// resize size of image
func Resize(src *image.NRGBA, width int, height int) *image.NRGBA {
	dstW, dstH := width, height

	if dstW < 0 || dstH < 0 {
		return src
	}
	if dstW == 0 && dstH == 0 {
		return src
	}

	srcW := src.Rect.Max.X
	srcH := src.Rect.Max.Y

	if srcW <= 0 || srcH <= 0 {
		return src
	}

	// if new width or height is 0 then preserve aspect ratio, minimum 1px
	if dstW == 0 {
		tmpW := float64(dstH) * float64(srcW) / float64(srcH)
		dstW = int(math.Max(1.0, math.Floor(tmpW+0.5)))
	}
	if dstH == 0 {
		tmpH := float64(dstW) * float64(srcH) / float64(srcW)
		dstH = int(math.Max(1.0, math.Floor(tmpH+0.5)))
	}

	var dst *image.NRGBA

	var sinc = func(x float64) float64 {
		if x == 0 {
			return 1
		}
		return math.Sin(math.Pi*x) / (math.Pi * x)
	}

	var filter resamplingFilter = resamplingFilter{
		Support: 3.0,
		Kernel: func(x float64) float64 {
			x = math.Abs(x)
			if x < 3.0 {
				return sinc(x) * sinc(x/3.0)
			}
			return 0
		},
	}

	if filter.Support <= 0.0 {
		// nearest-neighbor special case
		dst = resizeNearest(src, dstW, dstH)

	} else {
		// two-pass resize
		if srcW != dstW {
			dst = resizeHorizontal(src, dstW, filter)
		} else {
			dst = src
		}

		if srcH != dstH {
			dst = resizeVertical(dst, dstH, filter)
		}
	}

	return dst
}

func resizeHorizontal(src *image.NRGBA, width int, filter resamplingFilter) *image.NRGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y
	srcMaxX := srcBounds.Max.X

	dstW := width
	dstH := srcH

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	dX := float64(srcW) / float64(dstW)
	scaleX := math.Max(dX, 1.0)
	rX := math.Ceil(scaleX * filter.Support)

	// divide image to parts for parallel processing
	numGoroutines := runtime.NumCPU()
	goMaxProcs := runtime.GOMAXPROCS(0)
	if numGoroutines > goMaxProcs {
		numGoroutines = goMaxProcs
	}
	if numGoroutines > dstW {
		numGoroutines = dstW
	}
	partSize := dstW / numGoroutines

	doneChan := make(chan bool, numGoroutines)

	for part := 0; part < numGoroutines; part++ {
		partStart := part * partSize
		partEnd := (part + 1) * partSize
		if part == numGoroutines-1 {
			partEnd = dstW
		}

		go func(partStart, partEnd int) {

			for dstX := partStart; dstX < partEnd; dstX++ {
				fX := float64(srcMinX) + (float64(dstX)+0.5)*dX - 0.5

				startX := int(math.Ceil(fX - rX))
				if startX < srcMinX {
					startX = srcMinX
				}
				endX := int(math.Floor(fX + rX))
				if endX > srcMaxX-1 {
					endX = srcMaxX - 1
				}

				// cache weights
				weightSum := 0.0
				weights := make([]float64, int(rX+2)*2)
				for x := startX; x <= endX; x++ {
					w := filter.Kernel((float64(x) - fX) / scaleX)
					weightSum += w
					weights[x-startX] = w
				}

				for dstY := 0; dstY < dstH; dstY++ {
					srcY := srcMinY + dstY

					r, g, b, a := 0.0, 0.0, 0.0, 0.0
					for x := startX; x <= endX; x++ {
						weight := weights[x-startX]
						i := src.PixOffset(x, srcY)
						r += float64(src.Pix[i+0]) * weight
						g += float64(src.Pix[i+1]) * weight
						b += float64(src.Pix[i+2]) * weight
						a += float64(src.Pix[i+3]) * weight
					}

					r = math.Min(math.Max(r/weightSum, 0.0), 255.0)
					g = math.Min(math.Max(g/weightSum, 0.0), 255.0)
					b = math.Min(math.Max(b/weightSum, 0.0), 255.0)
					a = math.Min(math.Max(a/weightSum, 0.0), 255.0)

					j := dst.PixOffset(dstX, dstY)
					dst.Pix[j+0] = uint8(r + 0.5)
					dst.Pix[j+1] = uint8(g + 0.5)
					dst.Pix[j+2] = uint8(b + 0.5)
					dst.Pix[j+3] = uint8(a + 0.5)
				}
			}

			doneChan <- true
		}(partStart, partEnd)

	}

	// wait for goroutines to finish
	for part := 0; part < numGoroutines; part++ {
		<-doneChan
	}

	return dst
}

func resizeVertical(src *image.NRGBA, height int, filter resamplingFilter) *image.NRGBA {
	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y
	srcMaxY := srcBounds.Max.Y

	dstW := srcW
	dstH := height

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	dY := float64(srcH) / float64(dstH)
	scaleY := math.Max(dY, 1.0)
	rY := math.Ceil(scaleY * filter.Support)

	// divide image to parts for parallel processing
	numGoroutines := runtime.NumCPU()
	goMaxProcs := runtime.GOMAXPROCS(0)
	if numGoroutines > goMaxProcs {
		numGoroutines = goMaxProcs
	}
	if numGoroutines > dstH {
		numGoroutines = dstH
	}
	partSize := dstH / numGoroutines

	doneChan := make(chan bool, numGoroutines)

	for part := 0; part < numGoroutines; part++ {
		partStart := part * partSize
		partEnd := (part + 1) * partSize
		if part == numGoroutines-1 {
			partEnd = dstH
		}

		go func(partStart, partEnd int) {

			for dstY := partStart; dstY < partEnd; dstY++ {
				fY := float64(srcMinY) + (float64(dstY)+0.5)*dY - 0.5

				startY := int(math.Ceil(fY - rY))
				if startY < srcMinY {
					startY = srcMinY
				}
				endY := int(math.Floor(fY + rY))
				if endY > srcMaxY-1 {
					endY = srcMaxY - 1
				}

				// cache weights
				weightSum := 0.0
				weights := make([]float64, int(rY+2)*2)
				for y := startY; y <= endY; y++ {
					w := filter.Kernel((float64(y) - fY) / scaleY)
					weightSum += w
					weights[y-startY] = w
				}

				for dstX := 0; dstX < dstW; dstX++ {
					srcX := srcMinX + dstX

					r, g, b, a := 0.0, 0.0, 0.0, 0.0
					for y := startY; y <= endY; y++ {
						weight := weights[y-startY]
						i := src.PixOffset(srcX, y)
						r += float64(src.Pix[i+0]) * weight
						g += float64(src.Pix[i+1]) * weight
						b += float64(src.Pix[i+2]) * weight
						a += float64(src.Pix[i+3]) * weight
					}

					r = math.Min(math.Max(r/weightSum, 0.0), 255.0)
					g = math.Min(math.Max(g/weightSum, 0.0), 255.0)
					b = math.Min(math.Max(b/weightSum, 0.0), 255.0)
					a = math.Min(math.Max(a/weightSum, 0.0), 255.0)

					j := dst.PixOffset(dstX, dstY)
					dst.Pix[j+0] = uint8(r + 0.5)
					dst.Pix[j+1] = uint8(g + 0.5)
					dst.Pix[j+2] = uint8(b + 0.5)
					dst.Pix[j+3] = uint8(a + 0.5)
				}
			}

			doneChan <- true
		}(partStart, partEnd)

	}

	// wait for goroutines to finish
	for part := 0; part < numGoroutines; part++ {
		<-doneChan
	}

	return dst
}

// fast nearest-neighbor resize, no filtering
func resizeNearest(src *image.NRGBA, width, height int) *image.NRGBA {
	dstW, dstH := width, height

	srcBounds := src.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()
	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y
	srcMaxX := srcBounds.Max.X
	srcMaxY := srcBounds.Max.Y

	dst := image.NewNRGBA(image.Rect(0, 0, dstW, dstH))

	dx := float64(srcW) / float64(dstW)
	dy := float64(srcH) / float64(dstH)

	// divide image to parts for parallel processing
	numGoroutines := runtime.NumCPU()
	goMaxProcs := runtime.GOMAXPROCS(0)
	if numGoroutines > goMaxProcs {
		numGoroutines = goMaxProcs
	}
	if numGoroutines > dstH {
		numGoroutines = dstH
	}
	partSize := dstH / numGoroutines

	doneChan := make(chan bool, numGoroutines)

	for part := 0; part < numGoroutines; part++ {
		partStart := part * partSize
		partEnd := (part + 1) * partSize
		if part == numGoroutines-1 {
			partEnd = dstH
		}

		go func(partStart, partEnd int) {

			for dstY := partStart; dstY < partEnd; dstY++ {
				fy := float64(srcMinY) + (float64(dstY)+0.5)*dy - 0.5

				for dstX := 0; dstX < dstW; dstX++ {
					fx := float64(srcMinX) + (float64(dstX)+0.5)*dx - 0.5

					srcX := int(math.Min(math.Max(math.Floor(fx+0.5), float64(srcMinX)), float64(srcMaxX)))
					srcY := int(math.Min(math.Max(math.Floor(fy+0.5), float64(srcMinY)), float64(srcMaxY)))

					srcOffset := src.PixOffset(srcX, srcY)
					dstOffset := dst.PixOffset(dstX, dstY)

					dst.Pix[dstOffset+0] = src.Pix[srcOffset+0]
					dst.Pix[dstOffset+1] = src.Pix[srcOffset+1]
					dst.Pix[dstOffset+2] = src.Pix[srcOffset+2]
					dst.Pix[dstOffset+3] = src.Pix[srcOffset+3]
				}
			}

			doneChan <- true
		}(partStart, partEnd)
	}

	// wait for goroutines to finish
	for part := 0; part < numGoroutines; part++ {
		<-doneChan
	}

	return dst
}

// create a three dimenson slice
func New3DSlice(x int, y int, z int) (theSlice [][][]uint8) {
	theSlice = make([][][]uint8, x, x)
	for i := 0; i < x; i++ {
		s2 := make([][]uint8, y, y)
		for j := 0; j < y; j++ {
			s3 := make([]uint8, z, z)
			s2[j] = s3
		}
		theSlice[i] = s2
	}
	return
}

// create a new rgba matrix
func NewRGBAMatrix(height int, width int) (rgbaMatrix [][][]uint8) {
	rgbaMatrix = New3DSlice(height, width, 4)
	return
}

func Matrix2Vector(imgMatrix [][][]uint8) (vector []uint8) {
	h := len(imgMatrix)
	w := len(imgMatrix[0])
	r := len(imgMatrix[0][0])
	length := h * w * r

	vector = make([]uint8, length)

	for i := 0; i < h; i++ {
		for j := 0; j < w; j++ {
			for k := 0; k < r-1; k++ {
				vector = append(vector, imgMatrix[i][j][k])
			}
		}
	}
	return
}

func Dot(x []uint8, y []uint8) float64 {
	xlen := len(x)

	var sum float64 = 0
	for i := 0; i < xlen; i++ {
		sum = sum + float64(x[i])*float64(y[i])

	}
	return sum
}

type IterFunc func(i int, j int, k int, src [][][]uint8) [][][]uint8

func Iterator(filepath string, iter IterFunc) (imgMatrix [][][]uint8, err error) {
	imgMatrix, err = Read(filepath)
	if err != nil {
		return
	}

	height := len(imgMatrix)
	width := len(imgMatrix[0])
	pix := len(imgMatrix[0][0])
	if height == 0 || width == 0 {
		err = errors.New("The input of matrix is illegal!")
		return
	}

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			for k := 0; k < pix; k++ {
				imgMatrix = iter(i, j, k, imgMatrix)
			}
		}
	}
	return
}

func GetImageHeight(img image.Image) int {
	return img.Bounds().Max.Y
}

func GetImageWidth(img image.Image) int {
	return img.Bounds().Max.X
}

// decode a image and retrun golang image interface
func DecodeImage(i interface{}) (img image.Image, err error) {
	switch i.(type) {
	case string:
		reader, err := os.Open(i.(string))
		if err != nil {
			return nil, err
		}
		defer reader.Close()
		img, _, err = image.Decode(reader)
		return img, err
	case io.Reader:
		img, _, err = image.Decode(i.(io.Reader))
		return
	case image.Image:
		img = i.(image.Image)
		return
	default:
		err = errors.New("Incoming parameter error!")
		return
	}

}

//convert image to NRGBA
func convertToNRGBA(src image.Image) *image.NRGBA {
	srcBounds := src.Bounds()
	dstBounds := srcBounds.Sub(srcBounds.Min)

	dst := image.NewNRGBA(dstBounds)

	dstMinX := dstBounds.Min.X
	dstMinY := dstBounds.Min.Y

	srcMinX := srcBounds.Min.X
	srcMinY := srcBounds.Min.Y
	srcMaxX := srcBounds.Max.X
	srcMaxY := srcBounds.Max.Y

	switch src0 := src.(type) {

	case *image.NRGBA:
		rowSize := srcBounds.Dx() * 4
		numRows := srcBounds.Dy()

		i0 := dst.PixOffset(dstMinX, dstMinY)
		j0 := src0.PixOffset(srcMinX, srcMinY)

		di := dst.Stride
		dj := src0.Stride

		for row := 0; row < numRows; row++ {
			copy(dst.Pix[i0:i0+rowSize], src0.Pix[j0:j0+rowSize])
			i0 += di
			j0 += dj
		}

	case *image.NRGBA64:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				j := src0.PixOffset(x, y)

				dst.Pix[i+0] = src0.Pix[j+0]
				dst.Pix[i+1] = src0.Pix[j+2]
				dst.Pix[i+2] = src0.Pix[j+4]
				dst.Pix[i+3] = src0.Pix[j+6]

			}
		}

	case *image.RGBA:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				j := src0.PixOffset(x, y)
				a := src0.Pix[j+3]
				dst.Pix[i+3] = a

				switch a {
				case 0:
					dst.Pix[i+0] = 0
					dst.Pix[i+1] = 0
					dst.Pix[i+2] = 0
				case 0xff:
					dst.Pix[i+0] = src0.Pix[j+0]
					dst.Pix[i+1] = src0.Pix[j+1]
					dst.Pix[i+2] = src0.Pix[j+2]
				default:
					dst.Pix[i+0] = uint8(uint16(src0.Pix[j+0]) * 0xff / uint16(a))
					dst.Pix[i+1] = uint8(uint16(src0.Pix[j+1]) * 0xff / uint16(a))
					dst.Pix[i+2] = uint8(uint16(src0.Pix[j+2]) * 0xff / uint16(a))
				}
			}
		}

	case *image.RGBA64:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				j := src0.PixOffset(x, y)
				a := src0.Pix[j+6]
				dst.Pix[i+3] = a

				switch a {
				case 0:
					dst.Pix[i+0] = 0
					dst.Pix[i+1] = 0
					dst.Pix[i+2] = 0
				case 0xff:
					dst.Pix[i+0] = src0.Pix[j+0]
					dst.Pix[i+1] = src0.Pix[j+2]
					dst.Pix[i+2] = src0.Pix[j+4]
				default:
					dst.Pix[i+0] = uint8(uint16(src0.Pix[j+0]) * 0xff / uint16(a))
					dst.Pix[i+1] = uint8(uint16(src0.Pix[j+2]) * 0xff / uint16(a))
					dst.Pix[i+2] = uint8(uint16(src0.Pix[j+4]) * 0xff / uint16(a))
				}
			}
		}

	case *image.Gray:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				j := src0.PixOffset(x, y)
				c := src0.Pix[j]
				dst.Pix[i+0] = c
				dst.Pix[i+1] = c
				dst.Pix[i+2] = c
				dst.Pix[i+3] = 0xff

			}
		}

	case *image.Gray16:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				j := src0.PixOffset(x, y)
				c := src0.Pix[j]
				dst.Pix[i+0] = c
				dst.Pix[i+1] = c
				dst.Pix[i+2] = c
				dst.Pix[i+3] = 0xff

			}
		}

	case *image.YCbCr:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				yj := src0.YOffset(x, y)
				cj := src0.COffset(x, y)
				r, g, b := color.YCbCrToRGB(src0.Y[yj], src0.Cb[cj], src0.Cr[cj])

				dst.Pix[i+0] = r
				dst.Pix[i+1] = g
				dst.Pix[i+2] = b
				dst.Pix[i+3] = 0xff

			}
		}

	default:
		i0 := dst.PixOffset(dstMinX, dstMinY)
		for y := srcMinY; y < srcMaxY; y, i0 = y+1, i0+dst.Stride {
			for x, i := srcMinX, i0; x < srcMaxX; x, i = x+1, i+4 {

				c := color.NRGBAModel.Convert(src.At(x, y)).(color.NRGBA)

				dst.Pix[i+0] = c.R
				dst.Pix[i+1] = c.G
				dst.Pix[i+2] = c.B
				dst.Pix[i+3] = c.A

			}
		}
	}

	return dst
}

// read a image return a image matrix by path or image
func Read(imgOrPath interface{}) (imgMatrix [][][]uint8, err error) {
	var img image.Image
	switch imgOrPath.(type) {
	case string:
		img, err = DecodeImage(imgOrPath.(string))
		if err != nil {
			return
		}
	case image.Image:
		img = imgOrPath.(image.Image)
	default:
		err = errors.New("Incoming parameter error!")
		return
	}

	bounds := img.Bounds()
	width := bounds.Max.X  //width
	height := bounds.Max.Y //height

	src := convertToNRGBA(img)
	imgMatrix = NewRGBAMatrix(height, width)

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			c := src.At(j, i)
			r, g, b, a := c.RGBA()
			imgMatrix[i][j][0] = uint8(r)
			imgMatrix[i][j][1] = uint8(g)
			imgMatrix[i][j][2] = uint8(b)
			imgMatrix[i][j][3] = uint8(a)

		}
	}
	return
}

// read a image return a image matrix , if appear an error it will panic
func MustRead(filepath string) (imgMatrix [][][]uint8) {
	img, decodeErr := DecodeImage(filepath)
	if decodeErr != nil {
		panic(decodeErr)
	}

	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	src := convertToNRGBA(img)
	imgMatrix = NewRGBAMatrix(height, width)

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			c := src.At(j, i)
			r, g, b, a := c.RGBA()
			imgMatrix[i][j][0] = uint8(r)
			imgMatrix[i][j][1] = uint8(g)
			imgMatrix[i][j][2] = uint8(b)
			imgMatrix[i][j][3] = uint8(a)

		}
	}
	return
}

// save a image matrix as a png , if unsuccessful it will return a error
func SaveAsPNG(filepath string, imgMatrix [][][]uint8) error {
	height := len(imgMatrix)
	width := len(imgMatrix[0])

	if height == 0 || width == 0 {
		return errors.New("The input of matrix is illegal!")
	}

	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			nrgba.SetNRGBA(j, i, color.NRGBA{imgMatrix[i][j][0], imgMatrix[i][j][1], imgMatrix[i][j][2], imgMatrix[i][j][3]})
		}
	}
	outfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	png.Encode(outfile, nrgba)

	return nil
}

// save a image matrix as a jpeg,if unsuccessful it will return a error,quality must be 1 to 100
func SaveAsJPEG(filepath string, imgMatrix [][][]uint8, quality int) error {
	height := len(imgMatrix)
	width := len(imgMatrix[0])

	if height == 0 || width == 0 {
		return errors.New("The input of matrix is illegal!")
	}

	if quality < 1 {
		quality = 1
	} else if quality > 100 {
		quality = 100
	}

	nrgba := image.NewNRGBA(image.Rect(0, 0, width, height))

	for i := 0; i < height; i++ {
		for j := 0; j < width; j++ {
			nrgba.SetNRGBA(j, i, color.NRGBA{imgMatrix[i][j][0], imgMatrix[i][j][1], imgMatrix[i][j][2], imgMatrix[i][j][3]})
		}
	}
	outfile, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer outfile.Close()

	jpeg.Encode(outfile, nrgba, &jpeg.Options{Quality: quality})

	return nil
}
