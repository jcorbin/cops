package sheet

import (
	"image"
)

type Uniform string

type Strings struct {
	Strings []string
	Stride  int
	Rect    image.Rectangle
}

type Sheet interface {
	At(x, y int) string
	Bounds() image.Rectangle
}

func Copy(dest, src *Strings) {
	area := dest.Bounds().Intersect(src.Bounds())
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			dest.Set(x, y, src.At(x, y))
		}
	}
}

func (s *Strings) Write(x, y int, str string) (int, int) {
	w := s.Rect.Dx()
	for _, c := range str {
		s.Set(x, y, string(c))
		x++
		if x >= w {
			x = 0
			y++
		}
	}
	return x, y
}

func (s *Strings) Fill(str string) {
	area := s.Rect
	for y := area.Min.Y; y < area.Max.Y; y++ {
		for x := area.Min.X; x < area.Max.X; x++ {
			s.Set(x, y, str)
		}
	}
}

func NewStrings(r image.Rectangle) *Strings {
	w, h := r.Dx(), r.Dy()
	count := w * h
	buf := make([]string, count)
	s := &Strings{
		Strings: buf,
		Stride:  w,
		Rect:    r,
	}
	return s
}

func (s Strings) At(x, y int) string {
	if !(image.Point{x, y}.In(s.Rect)) {
		return ""
	}
	i := s.StringsOffset(x, y)
	return s.Strings[i]
}

func (s Strings) Bounds() image.Rectangle {
	return s.Rect
}

func (s Strings) Set(x, y int, str string) {
	if !(image.Point{x, y}.In(s.Rect)) {
		return
	}
	i := s.StringsOffset(x, y)
	s.Strings[i] = str
}

func (s Strings) SubSheet(r image.Rectangle) *Strings {
	r = r.Intersect(s.Rect)
	if r.Empty() {
		return &Strings{}
	}
	i := s.StringsOffset(r.Min.X, r.Min.Y)
	return &Strings{
		Strings: s.Strings[i:],
		Stride:  s.Stride,
		Rect:    r,
	}
}

func (s Strings) StringsOffset(x, y int) int {
	return (y-s.Rect.Min.Y)*s.Stride + (x - s.Rect.Min.X)
}
