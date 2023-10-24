package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/color"
	_ "image/jpeg"
	"image/png"
	_ "image/png"
	"math"
	"os"
	"strconv"
)

type vec3 struct {
	x, y, z float64
}

func (v vec3) length() float64 {
	return math.Sqrt(v.x*v.x + v.y*v.y + v.z*v.z)
}

func (v vec3) normalize() vec3 {
	l := v.length()
	return vec3{v.x / l, v.y / l, v.z / l}
}

func (v vec3) toColor() color.Color {
	return color.RGBA{
		uint8(v.x),
		uint8(v.y),
		uint8(v.z),
		255,
	}
}

func mul(v1, v2 vec3) vec3 {
	return vec3{v1.x * v2.x, v1.y * v2.y, v1.z * v2.z}
}

func add(v1, v2 vec3) vec3 {
	return vec3{v1.x + v2.x, v1.y + v2.y, v1.z + v2.z}
}

func addf(v vec3, f float64) vec3 {
	return vec3{v.x + f, v.y + f, v.z + f}
}

func clamp(x, min, max float64) float64 {
	return math.Min(math.Max(x, min), max)
}

func textureLookup(img image.Image, u, v float64) vec3 {
	xf := u * float64(img.Bounds().Dy())
	yf := v * float64(img.Bounds().Dy())

	x := int(math.Floor(xf))
	y := int(math.Floor(yf))

	x2 := x + 1
	y2 := y + 1

	diffx := xf - float64(x)
	diffy := yf - float64(y)

	// interpolation for fixing the edge lines
	A := img.At(x%img.Bounds().Dx(), int(clamp(float64(y), 0, float64(img.Bounds().Dy()-1))))
	B := img.At(int(clamp(float64(x2), 0, float64(img.Bounds().Dx()-1))), int(clamp(float64(y), 0, float64(img.Bounds().Dy()-1))))
	C := img.At(x%img.Bounds().Dx(), int(clamp(float64(y2), 0, float64(img.Bounds().Dy()-1))))
	D := img.At(int(clamp(float64(x2), 0, float64(img.Bounds().Dx()-1))), int(clamp(float64(y2), 0, float64(img.Bounds().Dy()-1))))

	r := colorToVec3(A).x*(1-diffx)*(1-diffy) + colorToVec3(B).x*diffx*(1-diffy) + colorToVec3(C).x*(1-diffx)*diffy + colorToVec3(D).x*diffx*diffy
	g := colorToVec3(A).y*(1-diffx)*(1-diffy) + colorToVec3(B).y*diffx*(1-diffy) + colorToVec3(C).y*(1-diffx)*diffy + colorToVec3(D).y*diffx*diffy
	b := colorToVec3(A).z*(1-diffx)*(1-diffy) + colorToVec3(B).z*diffx*(1-diffy) + colorToVec3(C).z*(1-diffx)*diffy + colorToVec3(D).z*diffx*diffy

	return vec3{clamp(r, 0, 255), clamp(g, 0, 255), clamp(b, 0, 255)}
}

func colorToVec3(c color.Color) vec3 {
	r, g, b, _ := c.RGBA()
	return vec3{float64(r) / 255, float64(g) / 255, float64(b) / 255}
}

func loadImage(filepath string) (image.Image, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}

	defer file.Close()
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

func saveImage(img image.Image, foldername, filename string) error {

	if _, err := os.Stat(foldername); os.IsNotExist(err) {
		err := os.Mkdir(foldername, 0755)
		if err != nil {
			return err
		}
	}

	file, err := os.Create(foldername + "/" + filename)

	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func sampleSphere(p vec3) vec3 {
	theta := math.Atan2(p.y, p.x)
	r := math.Hypot(p.x, p.y)
	phi := math.Atan2(p.z, r)

	u := (theta + math.Pi) / math.Pi
	v := (math.Pi/2 - phi) / math.Pi
	return vec3{u, v, 0}
}

func outImgToXYZ(i, j, faceIdx, faceSize int) vec3 {
	a := 2.0 * float64(i) / float64(faceSize)
	b := 2.0 * float64(j) / float64(faceSize)

	switch faceIdx {
	case 0:
		return vec3{1.0 - a, 1.0, 1.0 - b}
	case 1:
		return vec3{a - 1.0, -1.0, 1.0 - b}
	case 2:
		return vec3{b - 1.0, a - 1.0, 1.0}
	case 3:
		return vec3{1.0 - b, a - 1.0, -1.0}
	case 4:
		return vec3{1.0, a - 1.0, 1.0 - b}
	case 5:
		return vec3{-1.0, 1.0 - a, 1.0 - b}
	default:
		return vec3{0, 0, 0}
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Printf("Usage: %s <output_size> <input_image> <output_folder>\n", os.Args[0])
		fmt.Printf("Example: %s 512 skybox.png output\n", os.Args[0])
		return
	}

	inputImage := os.Args[2]

	img, err := loadImage(inputImage)
	if err != nil {
		fmt.Println(err)
		return
	}

	size, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}

	drawable := image.NewRGBA(image.Rect(0, 0, size, size))

	for faceIdx := 0; faceIdx < 6; faceIdx++ {
		fmt.Printf("Converting face %d\n", faceIdx)

		for x := 0; x < size; x++ {
			for y := 0; y < size; y++ {
				p := outImgToXYZ(x, y, faceIdx, size)
				uv := sampleSphere(p)
				drawable.Set(x, y, textureLookup(img, uv.x, uv.y).toColor())
			}
		}

		err := saveImage(drawable, os.Args[3], fmt.Sprintf("face%d.png", faceIdx))
		if err != nil {
			fmt.Println(err)
			return
		}

	}

}
