package albumart

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"strings"

	"github.com/nfnt/resize"
)


type ASCIIArt struct {
	Width  int
	Height int
	Art    string
}


type Converter struct {
	width  int
	height int
	chars  string
}


func NewConverter(width, height int) *Converter {
	return &Converter{
		width:  width,
		height: height,
		chars:  " .:-=+*#%@", 
	}
}


func (c *Converter) ConvertImageToASCII(imageData []byte) (*ASCIIArt, error) {
	if len(imageData) == 0 {
		return nil, fmt.Errorf("no image data provided")
	}

	
	img, err := c.decodeImage(imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	
	resized := resize.Resize(uint(c.width), uint(c.height), img, resize.Lanczos3)

	
	art := c.imageToASCII(resized)

	return &ASCIIArt{
		Width:  c.width,
		Height: c.height,
		Art:    art,
	}, nil
}


func (c *Converter) decodeImage(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	
	reader.Seek(0, 0)
	if img, err := png.Decode(reader); err == nil {
		return img, nil
	}

	
	reader.Seek(0, 0)
	if img, err := jpeg.Decode(reader); err == nil {
		return img, nil
	}

	
	reader.Seek(0, 0)
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("unsupported image format: %v", err)
	}

	return img, nil
}


func (c *Converter) imageToASCII(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y

	var result strings.Builder
	result.Grow(width * height + height) 

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			
			r, g, b, _ := img.At(x, y).RGBA()
			
			
			gray := (0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)) / 65535.0
			
			
			charIndex := int(gray * float64(len(c.chars)-1))
			if charIndex >= len(c.chars) {
				charIndex = len(c.chars) - 1
			}
			
			result.WriteByte(c.chars[charIndex])
		}
		if y < height-1 {
			result.WriteByte('\n')
		}
	}

	return result.String()
}


func CreatePlaceholderArt(width, height int) *ASCIIArt {
	var result strings.Builder
	
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if y == 0 || y == height-1 || x == 0 || x == width-1 {
				result.WriteByte('#')
			} else if y == height/2 && x == width/2-1 {
				result.WriteString("♪")
			} else if y == height/2 && x == width/2 {
				result.WriteString("♫")
			} else {
				result.WriteByte(' ')
			}
		}
		if y < height-1 {
			result.WriteByte('\n')
		}
	}

	return &ASCIIArt{
		Width:  width,
		Height: height,
		Art:    result.String(),
	}
}


func (a *ASCIIArt) GetColorizedASCII() string {
	lines := strings.Split(a.Art, "\n")
	var result strings.Builder

	for _, line := range lines {
		for _, char := range line {
			switch char {
			case ' ', '.', ':':
				result.WriteString(fmt.Sprintf("[gray]%c[white]", char))
			case '-', '=', '+':
				result.WriteString(fmt.Sprintf("[darkgray]%c[white]", char))
			case '*', '#', '%':
				result.WriteString(fmt.Sprintf("[lightgray]%c[white]", char))
			case '@':
				result.WriteString(fmt.Sprintf("[white]%c[white]", char))
			case '♪', '♫':
				result.WriteString(fmt.Sprintf("[yellow]%c[white]", char))
			default:
				result.WriteRune(char)
			}
		}
		result.WriteByte('\n')
	}

	return strings.TrimSuffix(result.String(), "\n")
}


func CreateSimpleArt(title, artist string, width, height int) *ASCIIArt {
	var result strings.Builder
	
	
	hash := simpleHash(title + artist)
	pattern := hash % 4
	
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			switch pattern {
			case 0: 
				if (x+y)%3 == 0 {
					result.WriteByte('/')
				} else if (x-y)%3 == 0 {
					result.WriteByte('\\')
				} else {
					result.WriteByte(' ')
				}
			case 1: 
				if (x*y)%5 == 0 && x > 0 && y > 0 {
					result.WriteString("•")
				} else {
					result.WriteByte(' ')
				}
			case 2: 
				if x%4 == (y%4) {
					result.WriteByte('~')
				} else {
					result.WriteByte(' ')
				}
			case 3: 
				if x == width/2 || y == height/2 {
					result.WriteByte('+')
				} else {
					result.WriteByte(' ')
				}
			}
		}
		if y < height-1 {
			result.WriteByte('\n')
		}
	}

	return &ASCIIArt{
		Width:  width,
		Height: height,
		Art:    result.String(),
	}
}


func simpleHash(s string) int {
	hash := 0
	for i, c := range s {
		hash += int(c) * (i + 1)
	}
	return hash
}