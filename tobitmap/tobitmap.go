package main

import (
	"flag"
	"fmt"
	"image" // register the PNG format with the image package
	"image/color"
	"image/png" // register the PNG format with the image package
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
	flag.Parse()
	if flag.NArg() < 3 {
		fmt.Printf("Usage: %s <FileName> <maxWidth> <maxHeight>\n", path.Base(os.Args[0]))
		os.Exit(1)
	}

	pngfile := flag.Arg(0)
	maxW, _ := strconv.ParseUint(flag.Arg(1), 10, 64)
	maxH, _ := strconv.ParseUint(flag.Arg(2), 10, 64)

	infile, err := os.Open(pngfile)
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer infile.Close()

	// Decode will figure out what type of image is in the file on its own.
	// We just have to be sure all the image packages we want are imported.
	src, _, err := image.Decode(infile)
	if err != nil {
		// replace this with real error handling
		panic(err)
	}

	// resize image
	src = resize.Thumbnail(uint(maxW), uint(maxH), src, resize.Bicubic)

	// Create a new grayscale image
	bounds := src.Bounds()
	gray := image.NewGray16(bounds.Bounds())

	name := strings.TrimSuffix(path.Base(pngfile), filepath.Ext(pngfile))
	fmt.Printf("import framebuf\n")
	fmt.Printf("%s = framebuf.FrameBuffer(bytearray([", name)
	for y := 0; y < bounds.Dy(); y += 8 {
		for x := 0; x < bounds.Dx(); x++ {
			var b uint8 = 0
			for by := 0; by < 8; by++ {
				oldColor := src.At(x, y)
				_, _, _, a := oldColor.RGBA()
				if a > 32767 {
					b |= 1 << uint(by)
				}
			}
			fmt.Printf("0x%x, ", b)
		}
		fmt.Printf("\n")
	}
	fmt.Printf("]), %d, %d, framebuf.MONO_VLSB)\n", src.Bounds().Dx(), src.Bounds().Dy())

	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			oldColor := src.At(x, y)
			_, _, _, a := oldColor.RGBA()
			var pixel uint16 = 0
			if a > 32767 {
				pixel = 65535
			}
			gray.Set(x, y, color.Gray16{pixel})
		}
	}

	// Encode the grayscale image to the output file
	outfile, err := os.Create("out.png")
	if err != nil {
		// replace this with real error handling
		panic(err)
	}
	defer outfile.Close()
	png.Encode(outfile, gray)
}