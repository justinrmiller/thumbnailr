package main

import (
	"flag"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

func processFile(file string, size uint) {
	fileOpen, err := os.Open(file)
	defer fileOpen.Close()

	if err != nil {
		log.Fatal(err)
	} else {
		originalImage, _, err := image.Decode(fileOpen)

		newImage := resize.Resize(size, 0, originalImage, resize.Lanczos3)

		outFile, err := os.Create(file + ".thumb")
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()
		err = jpeg.Encode(outFile, newImage, nil)
		log.Println("Processed image file:", file)
	}
}

func processWorker(wg *sync.WaitGroup, fileChan <-chan string, size uint) {
	defer wg.Done()

	for file := range fileChan {
		if file == "end_thumbnail" {
			break
		} else {
			processFile(file, size)
		}
	}
}

func main() {
	imagePath := flag.String("path", "", "Path to image files. (Required)")
	size := flag.Uint("size", 160, "Resize image to this size. (Optional, Default: 160)")
	flag.Parse()

	if *imagePath == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var files []string
	err := filepath.Walk(*imagePath, func(path string, info os.FileInfo, err error) error {
		if path != *imagePath && strings.Contains(path, ".jpg") && !strings.Contains(path, ".thumb") {
			files = append(files, path)
			log.Println("Found file:", path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	log.Println("Image Path:", *imagePath, "contains the following number of jpg files:", len(files))

	start := time.Now()

	jobs := make(chan string)

	var wg sync.WaitGroup

	wg.Add(runtime.NumCPU())
	for w := 1; w <= runtime.NumCPU(); w++ {
		log.Println("Spinning up image processor", w)
		go processWorker(&wg, jobs, *size)
	}

	for _, file := range files {
		jobs <- file
	}

	for w := 1; w <= runtime.NumCPU(); w++ {
		jobs <- "end_thumbnail"
	}

	wg.Wait()

	log.Println("Images left to process:", len(jobs))

	close(jobs)

	elapsed := time.Since(start)
	log.Println("Took", elapsed)
}
