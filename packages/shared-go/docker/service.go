package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/deployment"
)

type DockerConfig struct {
	Registry string
	Namespace string
	AccessToken string
	Username string
}

type DockerRunDto struct {
	ImageName string
	ContainerName string
	ContainerPort int
	EnvVars map[string]any
}

type DockerRunResult struct {
	HostPort int
}

func (cfg *DockerConfig) Validate() (bool, error) {
	if cfg.AccessToken == "" {
		return false, errors.New("AccessToken not provided/invalid")
	}
	
	if cfg.Registry == "" {
		return false, errors.New("Registry not provided/invalid")
	}
	
	return true, nil
}

type Docker interface {
	FindOneImageByName(string) (*image.Summary, error)
	Push(string) error
	Pull(string) error
	Run(DockerRunDto) (*DockerRunResult, error)
}

type docker struct {
	client *client.Client
	config *DockerConfig
	port_allocator deployment.PortAllocatorService
}

func New(config DockerConfig, port_allocator deployment.PortAllocatorService) (Docker, error) {
	docker_client, err := client.New(client.FromEnv)

	return &docker{docker_client, &config, port_allocator}, err
}

func (d *docker) FindOneImageByName(name string) (*image.Summary, error) {
	ctx := context.Background()
	image_list, err := d.client.ImageList(
		ctx, 
		client.ImageListOptions{}, 
	)
	if err != nil {
		return nil, err
	}

	for _, ct := range image_list.Items {
		for _, n := range ct.RepoTags {
			if n != "" && name == n {
				return &ct, nil
			}
		}
	}
	
	return nil, nil
}

func (d *docker) login() error {
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Push error: unable to find docker executable", err)
		return nil
	}
	
	logincmd := exec.Command(docker, "login", "-u", d.config.Username, "-p", d.config.AccessToken)
	if err := logincmd.Run(); err != nil {
		fmt.Println("Login error: failed to authenticate", err)
		return err
	}

	return nil
}

func (d *docker) Push(image_name string) error {
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Push error: unable to find docker executable", err)
		return nil
	}

	if err := d.login(); err != nil {
		return err
	}

	pushcmd := exec.Command(docker, "push", image_name)
	stdout, _ := pushcmd.StdoutPipe()
	stderr, _ := pushcmd.StderrPipe()
	logstream(stdout, stderr, "push out: ", "push err: ")
	if err := pushcmd.Run(); err != nil {
		fmt.Println("Push error: ", err)
		return err
	}
	
	return nil
}

func (d *docker) Pull(image_name string) error {
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Push error: unable to find docker executable", err)
		return nil
	}

	if err := d.login(); err != nil {
		return err
	}

	pushcmd := exec.Command(docker, "pull", image_name)
	stdout, _ := pushcmd.StdoutPipe()
	stderr, _ := pushcmd.StderrPipe()
	logstream(stdout, stderr, "pull out: ", "pull err: ")
	if err := pushcmd.Run(); err != nil {
		fmt.Println("Push error: ", err)
		return err
	}
	
	return nil
} 

func (d *docker) Run(dto DockerRunDto) (res *DockerRunResult, err error) {
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Push error: unable to find docker executable", err)
		return res, err
	}

	if err := d.login(); err != nil {
		return res, err
	}

	host_port, _ := d.port_allocator.GetFreePort()

	args := []string{
		"run",
		"-d",
		"-p",
		fmt.Sprintf("%d:%d", host_port, dto.ContainerPort), 
		"--name", dto.ContainerName, 
	}
	
	for k, v := range dto.EnvVars {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}
	args = append(args, dto.ImageName)

	run_cmd := exec.Command(
		docker,
		args...,
	)

	stdout, _ := run_cmd.StdoutPipe()
	stderr, _ := run_cmd.StderrPipe()
	logstream(stdout, stderr, "run out: ", "run err: ")
	if err := run_cmd.Run(); err != nil {
		fmt.Println("Push error: ", err)
		return res, err
	}

	res = &DockerRunResult{
		HostPort: host_port,
	}
	
	return res, err
} 

func logstream(stdout, stderr io.Reader, outprefix, errprefix string) {
	errbuf := make([]byte, 1024)
	outbuf := make([]byte, 1024)

	go func () {
		for n, err := stderr.Read(errbuf); err == nil; n, err = stderr.Read(errbuf) {
			fmt.Printf("%s%s", errprefix, string(errbuf[:n]))
		}

		fmt.Printf("%sclosed\n", errprefix)
	}()

	go func () {
		for n, err := stdout.Read(outbuf); err == nil; n, err = stdout.Read(outbuf) {
			fmt.Printf("%s%s", outprefix, string(outbuf[:n]))
		}

		fmt.Printf("%sclosed\n", outprefix)
	}()
}