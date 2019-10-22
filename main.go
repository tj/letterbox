package main

import (
	"flag"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/tj/go-sync/semaphore"
)

func main() {
	dir := flag.String("output", "processed", "Image output directory")
	white := flag.Bool("white", false, "Output a white letterbox")
	aspect := flag.String("aspect", "16:9", "Output aspect ratio")
	concurrency := flag.Int("concurrency", runtime.NumCPU(), "Concurrency of image processing")
	flag.Parse()

	// create destination directory
	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		log.Fatalf("error creating output directory: %s\n", err)
	}

	// parse aspect
	ratio, err := parseAspect(*aspect)
	if err != nil {
		log.Fatalf("error parsing aspect ratio: %s", err)
	}

	// images explicitly passed, or inferred
	images := flag.Args()
	if len(images) == 0 {
		images, err = listImages(".")
		if err != nil {
			log.Fatalf("error listing images: %s", err)
		}
	}

	// process
	sem := make(semaphore.Semaphore, *concurrency)
	start := time.Now()

	log.Printf("Processing %d images\n", len(images))
	for _, path := range images {
		path := path
		sem.Run(func() {
			log.Printf("Cropping %s\n", path)
			err := convert(path, *dir, *white, ratio)
			if err != nil {
				log.Fatalf("error converting %q: %s\n", path, err)
			}
		})
	}

	sem.Wait()
	log.Printf("Processed %d images in %s\n", len(images), time.Since(start).Round(time.Second))
}

// convert an image.
func convert(path, dir string, white bool, ratio float64) error {
	// open
	f, err := os.Open(path)
	if err != nil {
		return errors.Wrap(err, "opening")
	}
	defer f.Close()

	// decode
	src, _, err := image.Decode(f)
	if err != nil {
		return errors.Wrap(err, "decoding")
	}

	// dimensions
	sb := src.Bounds()
	sw := sb.Max.X
	sh := sb.Max.Y
	dw := sw
	dh := int(float64(dw) * ratio)

	// new image
	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	db := dst.Bounds()

	// dst rect
	dr := image.Rect(
		dw/2-(sw/2),
		dh/2-(sh/2),
		dw/2+dw,
		dh/2+sh)

	// color
	bg := color.Black
	if white {
		bg = color.White
	}

	// draw
	draw.Draw(dst, db, &image.Uniform{bg}, image.ZP, draw.Src)
	draw.Draw(dst, dr, src, src.Bounds().Min, draw.Src)

	// write
	return write(dst, filepath.Join(dir, path))
}

// write image to path.
func write(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating")
	}

	err = jpeg.Encode(f, img, &jpeg.Options{
		Quality: 90,
	})

	if err != nil {
		return errors.Wrap(err, "encoding")
	}

	return nil
}

// parseAspect returns a parsed aspect ratio.
func parseAspect(s string) (float64, error) {
	parts := strings.Split(s, ":")

	a, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, err
	}

	b, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, err
	}

	if b > a {
		a, b = b, a
	}

	return b / a, nil
}

// listImages returns the images in the given directory.
func listImages(dir string) (images []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".jpg" || ext == ".jpeg" {
			images = append(images, filepath.Join(dir, f.Name()))
		}
	}

	return
}
