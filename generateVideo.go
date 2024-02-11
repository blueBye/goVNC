package main

import (
	"os/exec"
)

func generateVideo(input string) error {
	cmd := exec.Command("./vnc2video.exe", input)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
