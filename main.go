package main

import (
	"flag"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/nfnt/resize"
)

func processFile(fileChan <-chan string, size uint) {
	for file := range fileChan {
		fmt.Println("Processing image file: ", file)

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

	fmt.Printf("Image Path: %s\n", *imagePath)

	var files []string
	err := filepath.Walk(*imagePath, func(path string, info os.FileInfo, err error) error {
		if path != *imagePath && strings.Contains(path, ".jpg") {
			files = append(files, path)
			fmt.Println("Found file: ", path)
		}
		return nil
	})

	if err != nil {
		panic(err)
	}

	fmt.Println("Total number of files: ", len(files))

	start := time.Now()

	jobs := make(chan string)

	for w := 1; w <= runtime.NumCPU(); w++ {
		fmt.Println("Spinning go routine ", w)
		go processFile(jobs, *size)
	}

	for _, file := range files {
		jobs <- file
	}
	close(jobs)

	elapsed := time.Since(start)
	log.Printf("Took %s", elapsed)
}
