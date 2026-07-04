package docker

import (
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
)

type Docker interface {
	FindOneImageByName(name string) (*image.Summary, error)
	Push(*image.Summary) error
}

type docker struct {
	client *client.Client
}

func New() (Docker, error) {
	docker_client, err := client.New(client.FromEnv)

	return &docker{docker_client}, err
}

func (d *docker) FindOneImageByName(name string) (*image.Summary, error) {
	return nil, nil
}

func (d *docker) Push(img *image.Summary) error {
	return nil
}