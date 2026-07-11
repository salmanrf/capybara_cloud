package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
)

type DockerConfig struct {
	Registry string
	AccessToken string
	Username string
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
	FindOneImageByName(name string) (*image.Summary, error)
	Push(image_name string) error
}

type docker struct {
	client *client.Client
	config *DockerConfig
}

func New(config DockerConfig) (Docker, error) {
	docker_client, err := client.New(client.FromEnv)
	return &docker{docker_client, &config}, err
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

func (d *docker) Push(image_name string) error {
	cfg := d.config
	
	docker, err := exec.LookPath("docker")
	if err != nil {
		fmt.Println("Push error: unable to find docker executable", err)
		return nil
	}

	logincmd := exec.Command(docker, "login", "-u", cfg.Username, "-p", cfg.AccessToken)
	if err := logincmd.Run(); err != nil {
		fmt.Println("Push login error: failed to authenticate", err)
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