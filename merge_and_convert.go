package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
)

func main() {

	tsOutputFileName := "merged_video.ts"
	mp4OutputFileName := "output.mp4"
	listFileName := "file_list.txt"
	tsFilesPattern := "*.ts"

	// Get the current working directory
	wd, e := os.Getwd()
	if e != nil {
		log.Fatalf("Failed to get current working directory: %s", e)
	}

	// Find all .ts files in the current directory
	tsFiles, e := filepath.Glob(filepath.Join(wd, tsFilesPattern))
	if e != nil {
		log.Fatalf("Failed to find .ts files: %s", e)
	}

	// Sort the TS files in ascending order using a custom sorting function
	sort.Slice(tsFiles, func(i, j int) bool {
		return extractNumber(tsFiles[i]) < extractNumber(tsFiles[j])
	})

	// Create a file list containing the paths of the .ts files
	listFile, e := os.Create(listFileName)
	if e != nil {
		log.Fatalf("Failed to create file list: %s", e)
	}
	defer listFile.Close()

	writer := bufio.NewWriter(listFile)
	for _, file := range tsFiles {
		fmt.Fprintf(writer, "file '%s'\n", file)
	}
	writer.Flush()

	// // Prepare the FFMPEG command to merge the TS files
	ffmpegMergeCommand := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", listFileName, "-c", "copy", tsOutputFileName)
	ffmpegMergeCommand.Stdout = os.Stdout
	ffmpegMergeCommand.Stderr = os.Stderr

	// Run the FFMPEG command to merge the TS files
	if e := ffmpegMergeCommand.Run(); e != nil {
		log.Fatalf("Failed to merge video files: %s", e)
	}

	// Prepare the FFMPEG command to convert the merged TS file to MP4
	ffmpegConvertCommand := exec.Command("ffmpeg", "-i", tsOutputFileName, "-c:v", "h264_nvenc", "-c:a", "aac", mp4OutputFileName)
	ffmpegConvertCommand.Stdout = os.Stdout
	ffmpegConvertCommand.Stderr = os.Stderr

	// Run the FFMPEG command to convert the merged TS file to MP4
	if e := ffmpegConvertCommand.Run(); e != nil {
		log.Fatalf("Failed to convert merged video file to MP4: %s", e)
	}

	fmt.Printf("Video files merged and converted to MP4 successfully. Output file: %s\n", mp4OutputFileName)
}

// Extracts the numeric part of a file name and returns it as an integer
func extractNumber(fileName string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(fileName)
	if match == "" {
		return 0
	}
	n, e := strconv.Atoi(match)
	if e != nil {
		return 0
	}
	return n
}
