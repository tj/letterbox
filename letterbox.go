package letterbox

import (
	"context"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Option function.
type Option func(*Processor) error

// Processor is a batch image processor for automating
// cropping and letterboxes.
type Processor struct {
	dir         string
	white       bool
	aspect      float64
	quality     int
	concurrency int
	force       bool
}

// New processor outputting to dir with the given options.
func New(dir string, options ...Option) (*Processor, error) {
	var v Processor
	v.concurrency = 1
	v.quality = 90
	v.dir = dir
	for _, o := range options {
		if err := o(&v); err != nil {
			return nil, err
		}
	}
	return &v, nil
}

// WithWhiteBackground changes the background color to white.
func WithWhiteBackground(v bool) Option {
	return func(p *Processor) error {
		p.white = v
		return nil
	}
}

// WithForce changes whether or not to force re-processing of existing images.
func WithForce(v bool) Option {
	return func(p *Processor) error {
		p.white = v
		return nil
	}
}

// WithAspect changes the aspect ratio which defaults to "16:9".
func WithAspect(ratio string) Option {
	return func(p *Processor) error {
		n, err := parseAspect(ratio)
		p.aspect = n
		return err
	}
}

// WithQuality changes the jpeg output quality, from 0-100.
func WithQuality(n int) Option {
	return func(p *Processor) error {
		p.quality = n
		return nil
	}
}

// WithConcurrency changes the processing concurrency.
func WithConcurrency(n int) Option {
	return func(p *Processor) error {
		p.concurrency = n
		return nil
	}
}

// Process the given images.
func (p *Processor) Process(ctx context.Context, images []string) error {
	sem := semaphore.NewWeighted(int64(p.concurrency))
	var errg errgroup.Group

	for _, path := range images {
		err := sem.Acquire(ctx, 1)
		if err != nil {
			return err
		}

		path := path
		errg.Go(func() error {
			defer sem.Release(1)
			return p.process(path)
		})
	}

	return errg.Wait()
}

// process implementation.
func (p *Processor) process(path string) error {
	dstpath := filepath.Join(p.dir, path)

	// open
	log.Printf("Processing %s\n", path)
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
	dh := int(float64(dw) * p.aspect)

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
	if p.white {
		bg = color.White
	}

	// draw
	draw.Draw(dst, db, &image.Uniform{bg}, image.ZP, draw.Src)
	draw.Draw(dst, dr, src, src.Bounds().Min, draw.Src)

	// write
	return writeImage(dst, dstpath, p.quality)
}

// writeImage writes a jpeg image to the given path.
func writeImage(img image.Image, path string, quality int) error {
	f, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "creating")
	}

	err = jpeg.Encode(f, img, &jpeg.Options{
		Quality: quality,
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

// unmodified returns true if the output image already exists,
// and is newer than the source image. Errors are treated
// as falsey.
func unmodified(src, dst string) bool {
	di, err := os.Stat(dst)

	// doesn't exist
	if os.IsNotExist(err) {
		return false
	}

	// exists, compare modified times
	si, err := os.Stat(src)
	if err != nil {
		return false
	}

	if di.ModTime().After(si.ModTime()) {
		return true
	}

	return false
}
