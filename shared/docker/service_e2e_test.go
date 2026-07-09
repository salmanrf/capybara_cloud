package docker

import (
	"fmt"
	"os/exec"
	"testing"
	"time"

	"github.com/salmanrf/capybara-cloud/shared/utils"
)

func dockerimagerm(docker_path string, name string) {
	dockerrm := exec.Command(docker_path, "image", "rm", "-f", name)
	if err := dockerrm.Run(); err != nil {
		fmt.Println("got unexpected error during container cleanup", err)
	}
}

func TestFindOneImageByNameE2E(t *testing.T) {
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	docker_service, err := New()
	if err != nil {
		t.Fatal("got unexpected error instantiating sut", err)
	}

	docker_cli, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("go unexpected error searching for docker cli executable", err)
	}

	tests := []struct{
		name string
	}{
		{"mrfreshgallery:123"},
		{fmt.Sprintf("mrserafino:%s", utils.DockerSafeDateString(time.Now()))},
	}
	
	for _, tt := range tests {
		dockerpull := exec.Command(docker_cli, "pull", "hello-world")
		if err := dockerpull.Run(); err != nil {
			t.Fatal("got unexpected error setting up test containers", err)
		}
		dockertag := exec.Command(docker_cli, "tag", "hello-world", tt.name)
		if err := dockertag.Run(); err != nil {
			t.Fatal("got unexpected error setting up test containers", err)
		}

		defer dockerimagerm(docker_cli, tt.name)

		res, err := docker_service.FindOneImageByName(tt.name)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}
		if res == nil {
			t.Fatalf("got container image summary %v, want non-nil", err)
		}
	}
}

func TestPushImageE2E(t *testing.T) {}