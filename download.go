package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

func main() {

	url := "<URL>"             // URL of the video files to download
	nFiles := 10               // Number of video files to download
	nConcurrentDownloads := 20 // Number of concurrent downloads allowed

	var wg sync.WaitGroup
	downloadSemaphore := make(chan struct{}, nConcurrentDownloads) // Semaphore to control the number of concurrent downloads

	for i := 1; i <= nFiles; i++ {

		fileURL := fmt.Sprintf("%s/video%d.ts", url, i) // Construct the URL of the video file
		fileName := fmt.Sprintf("video%d.ts", i)        // Construct the name of the video file

		wg.Add(1)
		downloadSemaphore <- struct{}{} // Acquire semaphore to limit the number of concurrent downloads

		go func(url, fileName string) {
			defer func() {
				<-downloadSemaphore // Release semaphore after download is complete
				wg.Done()
			}()

			e := downloadFile(url, fileName) // Download the video file
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

// Function to download a file from a given URL and save it with a given name
func downloadFile(url, fileName string) error {

	response, e := http.Get(url) // Send HTTP GET request to the URL
	if e != nil {
		return e
	}
	defer response.Body.Close()

	out, e := os.Create(fileName) // Create a new file with the given name
	if e != nil {
		return e
	}
	defer out.Close()

	_, e = io.Copy(out, response.Body) // Copy the response body to the file
	if e != nil {
		return e
	}

	return nil
}
