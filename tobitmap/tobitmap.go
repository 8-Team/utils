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

var (
	noResize = flag.Bool("noresize", false, "Do not resize.")
	level    = flag.Uint64("level", 32767, "Gray level to set monocrome")
	channel  = flag.String("channel", "a", "Select color [r,g,b,a]")
)

func main() {
	var maxH uint64
	var maxW uint64
	flag.Parse()
	if *noResize {
		if flag.NArg() < 1 {
			fmt.Printf("Usage: %s resize <FileName>\n", path.Base(os.Args[0]))
			os.Exit(1)
		}
	} else {
		if flag.NArg() < 3 {
			fmt.Printf("Usage: %s <FileName> <maxWidth> <maxHeight>\n", path.Base(os.Args[0]))
			os.Exit(1)
		}
		maxW, _ = strconv.ParseUint(flag.Arg(1), 10, 64)
		maxH, _ = strconv.ParseUint(flag.Arg(2), 10, 64)
	}

	pngfile := flag.Arg(0)

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
	if !*noResize {
		src = resize.Thumbnail(uint(maxW), uint(maxH), src, resize.MitchellNetravali)
	}

	// Create a new grayscale image
	bounds := src.Bounds()
	gray := image.NewGray16(bounds.Bounds())
	for y := 0; y < bounds.Dy(); y++ {
		for x := 0; x < bounds.Dx(); x++ {
			oldColor := src.At(x, y)
			pixel := color.Gray16Model.Convert(oldColor)
			var ch uint32
			r, g, b, a := pixel.RGBA()
			if r == 0 && b == 0 && g == 0 {
				r, g, b, a = oldColor.RGBA()
			}
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
		// replace this with real error handling
		panic(err)
	}
	defer outfile.Close()
	png.Encode(outfile, gray)

	name := strings.TrimSuffix(path.Base(pngfile), filepath.Ext(pngfile))
	fmt.Printf("%s = framebuf.FrameBuffer(bytearray([", name)
	for y := 0; y < bounds.Dy(); y += 8 {
		for x := 0; x < bounds.Dx(); x++ {
			var b uint8 = 0
			for by := 0; by < 8; by++ {
				oldPixel := src.At(x, by+y)
				pixel := color.Gray16Model.Convert(oldPixel)
				var ch uint32
				cr, cg, cb, ca := pixel.RGBA()
				if cr == 0 && cb == 0 && cg == 0 {
					cr, cg, cb, ca = oldPixel.RGBA()
				}
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
