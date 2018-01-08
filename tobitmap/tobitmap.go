package main

import (
	"flag"
	"fmt"
	"image" // register the PNG format with the image package
	"image/color"
	"image/png" // register the PNG format with the image package
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
)

var (
	level     = flag.Uint64("level", 32767, "Gray level to set monocrome")
	channel   = flag.String("channel", "a", "Select color [r,g,b,a]")
	maxWidth  = flag.Int("width", 0, "height of the final image, if 0 no resize is performed")
	maxHeight = flag.Int("height", 0, "height of the final image, if 0 no resize is performed")
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		log.Fatalf("Usage: %s resize <FileName>\n", path.Base(os.Args[0]))
	}

	pngfile := flag.Arg(0)

	infile, err := os.Open(pngfile)
	if err != nil {
		log.Fatalf("error while opening source %s", err)
	}
	defer infile.Close()

	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	if err != nil {
		log.Fatalf("error while decoding source %s", err)
	}

	// resize image
	if *maxWidth > 0 && *maxHeight > 0 {
		src = resize.Thumbnail(uint(*maxWidth), uint(*maxHeight), src, resize.MitchellNetravali)
	}

	// Create a new grayscale image
	bounds := src.Bounds()
	gray := image.NewGray16(bounds.Bounds())
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			oldColor := src.At(x, y)
			pixel := color.Gray16Model.Convert(oldColor)

			r, g, b, a := pixel.RGBA()
			if r == 0 && b == 0 && g == 0 {
				r, g, b, a = oldColor.RGBA()
			}

			var ch uint32
			switch *channel {
			case "r":
				ch = r
			case "g":
				ch = g
			case "b":
				ch = b
			case "a":
				ch = a
			}

			c := color.Gray16{Y: 0}
			if ch > uint32(*level) {
				c = color.Gray16{Y: 65535}
			}
			gray.Set(x, y, c)
		}
	}

	// Encode the grayscale image to the output file
	outfile, err := os.Create("out.png")
	if err != nil {
		log.Fatalf("error while creating destination file %s", err)
	}
	defer outfile.Close()

	if err := png.Encode(outfile, gray); err != nil {
		log.Fatalf("error while encoding image %s", err)
	}

	name := strings.TrimSuffix(path.Base(pngfile), filepath.Ext(pngfile))
	fmt.Printf("%s = framebuf.FrameBuffer(bytearray([", name)

	for y := 0; y < bounds.Dy(); y += 8 {
		for x := 0; x < bounds.Dx(); x++ {
			b := 0
			for by := 0; by < 8; by++ {
				oldPixel := src.At(x, by+y)
				pixel := color.Gray16Model.Convert(oldPixel)

				cr, cg, cb, ca := pixel.RGBA()
				if cr == 0 && cb == 0 && cg == 0 {
					cr, cg, cb, ca = oldPixel.RGBA()
				}

				var ch uint32
				switch *channel {
				case "r":
					ch = cr
				case "g":
					ch = cg
				case "b":
					ch = cb
				case "a":
					ch = ca
				}

				if ch > uint32(*level) {
					b |= 1 << uint(by)
				}
			}
			fmt.Printf("0x%x, ", b)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("]), %d, %d, framebuf.MONO_VLSB)\n", src.Bounds().Dx(), src.Bounds().Dy())
}
