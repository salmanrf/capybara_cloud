package docker

import (
	"context"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
)

type DockerConfig struct {
	Registry string
	AccessToken string
}

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

func (d *docker) Push(img *image.Summary) error {
	ctx := context.Background()
	opts := client.ImagePushOptions{
		RegistryAuth: "abcd",
	}
	
	d.client.ImagePush(ctx, "123", opts)
	
	return nil
}