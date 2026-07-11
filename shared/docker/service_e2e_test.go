package docker

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/shared/utils"
)

func load_dockercfg() (DockerConfig, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return DockerConfig{}, err
	}

	err = godotenv.Load(path.Join(pwd, ".env.test"))
	if err != nil {
		return DockerConfig{}, err
	}

	cfg := DockerConfig{
		Registry: os.Getenv("DOCKER_REGISTRY"),
		AccessToken: os.Getenv("DOCKER_ACCESS_TOKEN"),
		Username: os.Getenv("DOCKER_USER"),
	}

	return cfg, nil
}

func dockerimagerm(docker_path string, name string) {
	dockerrm := exec.Command(docker_path, "image", "rm", "-f", name)
	if err := dockerrm.Run(); err != nil {
		fmt.Println("got unexpected error during container cleanup", err)
	}
}

func TestDockerConfig(t *testing.T) {
	t.Run("should return error if Registry is invalid", func (t *testing.T) {
		cfg := DockerConfig{
			AccessToken: "abcd",
		}
		valid, err := cfg.Validate()

		want_error := errors.New("Registry not provided/invalid")
		
		if err == nil {
			t.Errorf("got error %v, want %v", err, want_error)
		}
		if err.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", err, want_error)
		}
		if valid {
			t.Errorf("got error %v, want %v", valid, false)
		}
	})

	t.Run("should return error if AccessToken is invalid", func (t *testing.T) {
		cfg := DockerConfig{
			Registry: "abcd",
		}
		valid, err := cfg.Validate()

		want_error := errors.New("AccessToken not provided/invalid")
		
		if err == nil {
			t.Errorf("got error %v, want %v", err, want_error)
		}
		if err.Error() != want_error.Error() {
			t.Errorf("got error %v, want %v", err, want_error)
		}
		if valid {
			t.Errorf("got error %v, want %v", valid, false)
		}
	})

	t.Run("should return true and error if all fields valid", func (t *testing.T) {
		cfg := DockerConfig{
			Registry: "abcd",
			AccessToken: "zsh",
		}
		valid, err := cfg.Validate()

		var want_error error = nil
		
		if err != nil {
			t.Errorf("got error %v, want %v", err, want_error)
		}
		if !valid {
			t.Errorf("got valid %v, want %v", valid, true)
		}
	})
}

func TestFindOneImageByNameE2E(t *testing.T) {
	cfg := DockerConfig{}
	
	_, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	docker_service, err := New(cfg)
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

	defer dockerimagerm(docker_cli, "hello-world")
	
	for _, tt := range tests {
		t.Run(fmt.Sprintf("should find image %s", tt.name), func (t *testing.T) {
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
		})
	}
}

func TestPushImageE2E(t *testing.T) {
	flag.Set("test.timeout", "5m")
	
	cfg, err := load_dockercfg()
	if err != nil {
		t.Fatal("got unexpected error loading docker config from env", err)
	}

	_, err = exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	docker_service, err := New(cfg)
	if err != nil {
		t.Fatal("got unexpected error instantiating sut", err)
	}

	docker_cli, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("go unexpected error searching for docker cli executable", err)
	}

	tests := []struct{
		tag string
	}{
		{"mrbovinocat:123"},
		{fmt.Sprintf("mrserafino:%s", utils.DockerSafeDateString(time.Now()))},
	}
	
	defer dockerimagerm(docker_cli, "hello-world")
	
	for _, tt := range tests {
		dockerpull := exec.Command(docker_cli, "pull", "hello-world")
		stdout, _ := dockerpull.StdoutPipe()
		stderr, _ := dockerpull.StderrPipe()
		logstream(stdout, stderr, "out: ", "err: ")
		if err := dockerpull.Run(); err != nil {
			t.Fatal("got unexpected error setting up test containers", err)
		}

		docker_full_path := fmt.Sprintf("%s/%s", cfg.Registry, tt.tag)
		
		dockertag := exec.Command(docker_cli, "tag", "hello-world", docker_full_path)
		if err := dockertag.Run(); err != nil {
			t.Fatal("got unexpected error setting up test containers", err)
		}

		defer dockerimagerm(docker_cli, docker_full_path)

		err = docker_service.Push(docker_full_path)
		if err != nil {
			t.Fatalf("got error %v, want nil", err)
		}

		dockerpull = exec.Command(docker_cli, "pull", docker_full_path)
		stdout, _ = dockerpull.StdoutPipe()
		stderr, _ = dockerpull.StderrPipe()
		logstream(stdout, stderr, "[pull after push out]: ", "[pull after push err]: ")
		if err := dockerpull.Run(); err != nil {
			t.Fatalf("got error pulling container '%s': '%v', want nil", docker_full_path, err)
		}
	}
}