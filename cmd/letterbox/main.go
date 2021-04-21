package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/tj/letterbox"
)

func main() {
	dir := flag.String("output", "processed", "Image output directory")
	white := flag.Bool("white", false, "Output a white letterbox")
	aspect := flag.String("aspect", "16:9", "Output aspect ratio")
	quality := flag.Int("quality", 90, "Output jpeg quality")
	padding := flag.Int("padding", 0, "Output image padding in percentage")
	concurrency := flag.Int("concurrency", runtime.NumCPU(), "Concurrency of image processing")
	force := flag.Bool("force", false, "Force image reprocess when it exists")
	flag.Parse()

	// create destination directory
	err := os.MkdirAll(*dir, 0755)
	if err != nil {
		log.Fatalf("error creating output directory: %s\n", err)
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
	start := time.Now()
	log.Printf("Processing %d images\n", len(images))

	processor, err := letterbox.New(*dir,
		letterbox.WithWhiteBackground(*white),
		letterbox.WithConcurrency(*concurrency),
		letterbox.WithQuality(*quality),
		letterbox.WithForce(*force),
		letterbox.WithAspect(*aspect),
		letterbox.WithPadding(*padding),
	)

	if err != nil {
		log.Fatalf("error creating proessor: %s", err)
	}

	ctx := context.Background()
	err = processor.Process(ctx, images)
	if err != nil {
		log.Fatalf("error processing: %s", err)
	}

	log.Printf("Processed in %s\n", time.Since(start).Round(time.Second))
}

// listImages returns the images in the given directory.
func listImages(dir string) (images []string, err error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		ext := strings.ToLower(filepath.Ext(f.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".tif" {
			images = append(images, filepath.Join(dir, f.Name()))
		}
	}

	return
}
