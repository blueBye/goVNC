package main

import (
	"os/exec"
)

func generateVideo(input string) error {
	cmd := exec.Command("./vnc2video", input)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
