package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func generateVideo(filename string, duration float64) error {
	cmd := exec.Command("./vnc2video", filename)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

func getVideoDuration(filename string) (float64, error) {
	// Run ffprobe command
	cmd := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=duration", "-of", "default=noprint_wrappers=1:nokey=1", filename)

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out

	// Run the command and check for errors
	err := cmd.Run()
	if err != nil {
		return 0, err
	}

	// Convert the output to a float64 (duration in seconds)
	durationStr := strings.TrimSpace(out.String())
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

func fixSpeed(filename string, currentDuration float64, targetDuration float64) error {
	ratio := currentDuration / targetDuration
	outputFilename := fmt.Sprintf("%s-fixed.mp4", filename)
	filename += ".mp4"

	// Build the ffmpeg command with individual arguments
	cmd := exec.Command("ffmpeg", "-i", filename, "-filter:v", fmt.Sprintf("setpts=PTS/%.6f", ratio), outputFilename)

	// Run the command
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run ffmpeg command: %v", err)
	}

	return nil
}
