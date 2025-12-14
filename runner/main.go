package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	path, err := exec.LookPath("tar")
	if err != nil {
		fmt.Println("Unable to find tar executable", err)
		os.Exit(1)
	}

	cmd := exec.Command(path, "-cvf", "samples.tar", "./samples")
	cmd.Stderr = os.Stdout

	if err := cmd.Start(); err != nil {
		fmt.Println("Unable to start deployment bundling", err)
		os.Exit(1)
	}
	
	if err := cmd.Wait(); err != nil {
		fmt.Println("Failed to execute deployment bundling", err)
		os.Exit(1)
	}

	fmt.Println("Deployment bundled successfully")
}