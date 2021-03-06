package pt

import (
	"fmt"
	"image"
	"math"
)

type Texture interface {
	Sample(u, v float64) Color
	NormalSample(u, v float64) Vector
	BumpSample(u, v float64) Vector
	Pow(a float64) Texture
	MulScalar(a float64) Texture
}

var textures map[string]Texture

func init() {
	textures = make(map[string]Texture)
}

func GetTexture(path string) Texture {
	if texture, ok := textures[path]; ok {
		return texture
	}
	if texture, err := LoadTexture(path); err == nil {
		textures[path] = texture
		return texture
	}
	return nil
}

func LoadTexture(path string) (Texture, error) {
	fmt.Printf("Loading IMG: %s\n", path)
	im, err := LoadImage(path)
	if err != nil {
		return nil, err
	}
	return NewTexture(im), nil
}

type ColorTexture struct {
	Width  int
	Height int
	Data   []Color
}

func NewTexture(im image.Image) Texture {
	size := im.Bounds().Max
	data := make([]Color, size.X*size.Y)
	for y := 0; y < size.Y; y++ {
		for x := 0; x < size.X; x++ {
			index := y*size.X + x
			data[index] = NewColor(im.At(x, y)).Pow(2.2)
		}
	}
	return &ColorTexture{size.X, size.Y, data}
}

func (t *ColorTexture) Pow(a float64) Texture {
	for i := range t.Data {
		t.Data[i] = t.Data[i].Pow(a)
	}
	return t
}

func (t *ColorTexture) MulScalar(a float64) Texture {
	for i := range t.Data {
		t.Data[i] = t.Data[i].MulScalar(a)
	}
	return t
}

func (t *ColorTexture) bilinearSample(u, v float64) Color {
	if u == 1 {
		u -= EPS
	}
	if v == 1 {
		v -= EPS
	}
	w := float64(t.Width) - 1
	h := float64(t.Height) - 1
	X, x := math.Modf(u * w)
	Y, y := math.Modf(v * h)
	x0 := int(X)
	y0 := int(Y)
	x1 := x0 + 1
	y1 := y0 + 1
	c00 := t.Data[y0*t.Width+x0]
	c01 := t.Data[y1*t.Width+x0]
	c10 := t.Data[y0*t.Width+x1]
	c11 := t.Data[y1*t.Width+x1]
	c := Black
	c = c.Add(c00.MulScalar((1 - x) * (1 - y)))
	c = c.Add(c10.MulScalar(x * (1 - y)))
	c = c.Add(c01.MulScalar((1 - x) * y))
	c = c.Add(c11.MulScalar(x * y))
	return c
}

func (t *ColorTexture) Sample(u, v float64) Color {
	u = Fract(Fract(u) + 1)
	v = Fract(Fract(v) + 1)
	return t.bilinearSample(u, 1-v)
}

func (t *ColorTexture) NormalSample(u, v float64) Vector {
	c := t.Sample(u, v)
	return Vector{c.R*2 - 1, c.G*2 - 1, c.B*2 - 1}.Normalize()
}

func (t *ColorTexture) BumpSample(u, v float64) Vector {
	u = Fract(Fract(u) + 1)
	v = Fract(Fract(v) + 1)
	v = 1 - v
	x := int(u * float64(t.Width))
	y := int(v * float64(t.Height))
	x1, x2 := ClampInt(x-1, 0, t.Width-1), ClampInt(x+1, 0, t.Width-1)
	y1, y2 := ClampInt(y-1, 0, t.Height-1), ClampInt(y+1, 0, t.Height-1)
	cx := t.Data[y*t.Width+x1].Sub(t.Data[y*t.Width+x2])
	cy := t.Data[y1*t.Width+x].Sub(t.Data[y2*t.Width+x])
	return Vector{cx.R, cy.R, 0}
}
