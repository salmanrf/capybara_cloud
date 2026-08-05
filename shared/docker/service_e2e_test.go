package docker

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/salmanrf/capybara-cloud/shared/deployment"
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
		Namespace: os.Getenv("DOCKER_NAMESPACE"),
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

func dockercontainerrm(docker_path string, name string) {
	dockerrm := exec.Command(docker_path, "rm", "-f", name)
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

	docker_service, err := New(cfg, &deployment.StubPortAllocatorService{})
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
	flag.Set("test.timeout", "10m")
	
	cfg, err := load_dockercfg()
	if err != nil {
		t.Fatal("got unexpected error loading docker config from env", err)
	}

	_, err = exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	docker_service, err := New(cfg, &deployment.StubPortAllocatorService{})
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

func TestPullImageE2E(t *testing.T) {
	flag.Set("test.timeout", "10m")

	cfg, err := load_dockercfg()
	if err != nil {
		t.Fatal("got unexpected error loading docker config from env", err)
	}

	docker, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	tests := []struct{
		name string
	}{
		{"redis:8.10.0-alpine3.23"},
		{"apache/kafka:latest"},
		{fmt.Sprintf("%s/mrserafino:2026-08-04-08-21-36", cfg.Namespace)},
		{fmt.Sprintf("%s/mrbovino:123", cfg.Namespace)},
	}

	for _, tt := range tests {
		t.Run("should be able to pull image from public (docker) & private repository (capy cloud)", func (t *testing.T) {
			defer dockerimagerm(docker, tt.name)

			docker_service, err := New(cfg, &deployment.StubPortAllocatorService{})
			if err != nil {
				t.Fatal("got unexpected error instantiating sut", err)
			}

			err = docker_service.Pull(tt.name)
			if err != nil {
				t.Fatalf("got unexpected error %v, want nil", err)
			}

			img, err := docker_service.FindOneImageByName(tt.name)
			if err != nil {
				t.Fatalf("got unexpected error %v, want nil", err)
			}
			if img == nil {
				t.Errorf("got image '%s' nil, want found", tt.name)
			}
		})
	}
} 

func TestRunContainerE2E(t *testing.T) {
	flag.Set("test.timeout", "10m")

	port_allocator := &deployment.StubPortAllocatorService{}

	cfg, err := load_dockercfg()
	if err != nil {
		t.Fatal("got unexpected error loading docker config from env", err)
	}

	docker, err := exec.LookPath("docker")
	if err != nil {
		t.Fatal("got unexpected error when searching docker executable", err)
	}

	tests := []struct{
		ImageName 		string
		ContainerName string
		ContainerPort int
		HostPort int
		EnvVars map[string]any
	}{
		{
			"redis:8.10.0-alpine3.23", 
			"myredis", 
			3000, 
			3111,
			map[string]any{
				"a": "b",
				"c": "d",
				"e": "f",
			},
		},
		{
			"apache/kafka:latest", 
			"mykafka", 
			3000, 
			3222,
			map[string]any{
				"CLIENT_ID": "abcd",
				"CLIENT_SECRET": "zxcvbnm",
			},
		},
	}

	var wg sync.WaitGroup
	
	defer func () {
		wg.Wait()
		for _, tt := range tests {
			dockerimagerm(docker, tt.ImageName)
			fmt.Println("Images cleaned up")
		}
	} ()

	t.Run("should run container from image", func (t *testing.T) {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("should start '%s' from '%s'", tt.ContainerName, tt.ImageName), func (t *testing.T) {
				wg.Add(1)
				
				dockercontainerrm(docker, tt.ContainerName)
				defer func () {
					dockercontainerrm(docker, tt.ContainerName)
					port_allocator.Clear()
					wg.Done()
				}()
	
				docker_service, err := New(cfg, port_allocator)
				if err != nil {
					t.Fatal("got unexpected error instantiating sut", err)
				}
	
				dto := DockerRunDto{
					ImageName: tt.ImageName,
					ContainerName: tt.ContainerName,
					ContainerPort: tt.ContainerPort,
				}

				port_allocator.Get_free_port_return = tt.HostPort
	
				res, err := docker_service.Run(dto)
				if err != nil {
					t.Fatalf("got unexpected error %v, want nil", err)
				}
	
				inspect_buf := bytes.NewBuffer([]byte{})
				docker_inspect := exec.Command(docker, "inspect", dto.ContainerName)
				docker_inspect.Stdout = inspect_buf
				if err := docker_inspect.Run(); err != nil {
					t.Fatalf("got unexpected error inspecting container '%s': %v", dto.ContainerName, err)
				}
	
				var inspect_output []ContainerInspect
				inspectDecoder := json.NewDecoder(inspect_buf)
	
				err = inspectDecoder.Decode(&inspect_output)
				if err != nil {
					t.Fatal("got unexpected error decoding docker inspect output", err)
				}
				if len(inspect_output) == 0 {
					t.Fatal("got 0 result from docker inspect")
				}
				container := inspect_output[0]
	
				portKey := fmt.Sprintf("%d/tcp", tt.ContainerPort)
				gotPortMappings, gotPortOk := container.NetworkSettings.Ports[portKey]
				if !gotPortOk {
					t.Fatalf("got container port '%s' unassigned, want having '%s'", portKey, portKey)
				}
				if len(gotPortMappings) == 0 {
					t.Fatalf("got empty port mappings for '%s', want at least one binding", portKey)
				}

				// * First check if the host port was the same as one assigned by allocator
				wantAllocatorHostPort := port_allocator.Get_free_port_return
				gotAllocatorHostPort := res.HostPort
				if gotAllocatorHostPort != wantAllocatorHostPort {
					t.Errorf("got host port '%v', want '%v'", gotAllocatorHostPort, wantAllocatorHostPort)
				}
	
				// * Then check if it matches the inspect too
				wantInspectHostPort := fmt.Sprintf("%d", res.HostPort)
				gotInspectHostPort := gotPortMappings[0].HostPort
				if gotInspectHostPort != wantInspectHostPort {
					t.Errorf("got host port '%v', want '%v'", gotInspectHostPort, wantInspectHostPort)
				}
			})
		}
	})

	t.Run("should run container with the correct environment variables", func (t *testing.T) {
		for _, tt := range tests {
			t.Run(fmt.Sprintf("should start '%s' from '%s'", tt.ContainerName, tt.ImageName), func (t *testing.T) {
				wg.Add(1)
				
				dockercontainerrm(docker, tt.ContainerName)
				defer func () {
					dockercontainerrm(docker, tt.ContainerName)
					port_allocator.Clear()
					wg.Done()
				}()
	
				docker_service, err := New(cfg, port_allocator)
				if err != nil {
					t.Fatal("got unexpected error instantiating sut", err)
				}
	
				dto := DockerRunDto{
					ImageName: tt.ImageName,
					ContainerName: tt.ContainerName,
					ContainerPort: tt.ContainerPort,
					EnvVars: tt.EnvVars,
				}

				port_allocator.Get_free_port_return = tt.HostPort
	
				res, err := docker_service.Run(dto)
				_ = res
				if err != nil {
					t.Fatalf("got unexpected error %v, want nil", err)
				}
	
				inspect_buf := bytes.NewBuffer([]byte{})
				docker_inspect := exec.Command(docker, "inspect", dto.ContainerName)
				docker_inspect.Stdout = inspect_buf
				if err := docker_inspect.Run(); err != nil {
					t.Fatalf("got unexpected error inspecting container '%s': %v", dto.ContainerName, err)
				}
	
				var inspect_output []ContainerInspect
				inspectDecoder := json.NewDecoder(inspect_buf)
	
				err = inspectDecoder.Decode(&inspect_output)
				if err != nil {
					t.Fatal("got unexpected error decoding docker inspect output", err)
				}
				if len(inspect_output) == 0 {
					t.Fatal("got 0 result from docker inspect")
				}

				container := inspect_output[0]
				env_vars := container.Config.Env
				want_env_vars := tt.EnvVars
				got_env_vars := utils.MapFromEnvVarStrings(env_vars)

				for want_k, want_v := range want_env_vars {
					v, _ := got_env_vars[want_k];
					got_v, _ := v.(string)

					if got_v != want_v {
						t.Errorf("got %s == '%v', want %v", want_k, got_v, want_v)
					}
				}
			})
		}
	})
}