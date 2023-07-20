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

	mp4OutputFileName := "output.mp4" // Output file name for the final merged and converted video
	listFileName := "file_list.txt"   // File name for the temporary file list
	tsFilesPattern := "*.ts"          // Pattern to match the input .ts files

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

	// Merge audio streams from the .ts files into a single audio file
	ffmpegMergeAudioCommand := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", listFileName, "-map", "0:0", "-map", "0:1", "-c", "copy", "-strict", "-2", "-vn", "audio.ts")
	ffmpegMergeAudioCommand.Stdout = os.Stdout
	ffmpegMergeAudioCommand.Stderr = os.Stderr

	if e := ffmpegMergeAudioCommand.Run(); e != nil {
		log.Fatalf("Failed to merge audio files: %s", e)
	}

	// Merge video streams from the .ts files into a single video file
	ffmpegMergeVideoCommand := exec.Command("ffmpeg", "-f", "concat", "-safe", "0", "-i", listFileName, "-map", "0:0", "-map", "0:1", "-an", "-c", "copy", "video.ts")
	ffmpegMergeVideoCommand.Stdout = os.Stdout
	ffmpegMergeVideoCommand.Stderr = os.Stderr

	if e := ffmpegMergeVideoCommand.Run(); e != nil {
		log.Fatalf("Failed to merge video files: %s", e)
	}

	// Convert the merged audio file to FLAC format
	ffmpegFlacCommand := exec.Command("ffmpeg", "-i", "audio.ts", "-map", "0:0", "-codec:a", "flac", "-ac", "2", "-ar", "48000", "-sample_fmt", "s16", "audio.flac")
	ffmpegFlacCommand.Stdout = os.Stdout
	ffmpegFlacCommand.Stderr = os.Stderr

	if e := ffmpegFlacCommand.Run(); e != nil {
		log.Fatalf("Failed to convert audio file to FLAC: %s", e)
	}

	// Convert the merged video file to MP4 format
	ffmpegConvertCommand := exec.Command("ffmpeg", "-i", "video.ts", "-map", "0:0", "-c", "copy", "-an", "video.mp4")
	ffmpegConvertCommand.Stdout = os.Stdout
	ffmpegConvertCommand.Stderr = os.Stderr

	if e := ffmpegConvertCommand.Run(); e != nil {
		log.Fatalf("Failed to convert video file to MP4: %s", e)
	}

	// Merge the converted video and audio files into a final MP4 file
	ffmpegMergeCommand := exec.Command("ffmpeg", "-i", "video.mp4", "-i", "audio.flac", "-c", "copy", "-map", "0:0", "-map", "1:0", "-map_metadata", "-1", "-movflags", "+faststart", "-strict", "-3", "-f", "mp4", mp4OutputFileName)
	ffmpegMergeCommand.Stdout = os.Stdout
	ffmpegMergeCommand.Stderr = os.Stderr

	if e := ffmpegMergeCommand.Run(); e != nil {
		log.Fatalf("Failed to merge video and audio files: %s", e)
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
