package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func main() {

	url := "<URL>"
	nFiles := 10
	nConcurrentDownloads := 20

	var wg sync.WaitGroup
	downloadSemaphore := make(chan struct{}, nConcurrentDownloads)

	for i := 1; i <= nFiles; i++ {

		fileURL := fmt.Sprintf("%s/video%d.ts", url, i)
		fileName := fmt.Sprintf("video%d.ts", i)

		wg.Add(1)
		downloadSemaphore <- struct{}{} // Acquire semaphore

		go func(url, fileName string) {
			defer func() {
				<-downloadSemaphore // Release semaphore
				wg.Done()
			}()

			e := downloadFile(url, fileName)
			if e != nil {
				fmt.Printf("Failed to download %s: %s\n", fileName, e.Error())
				return
			}

			fmt.Printf("File %s downloaded successfully.\n", fileName)
		}(fileURL, fileName)
	}

	wg.Wait()
	fmt.Println("All files downloaded.")
}

func downloadFile(url, fileName string) error {

	response, e := http.Get(url)
	if e != nil {
		return e
	}
	defer response.Body.Close()

	out, e := os.Create(fileName)
	if e != nil {
		return e
	}
	defer out.Close()

	_, e = io.Copy(out, response.Body)
	if e != nil {
		return e
	}

	return nil
}
