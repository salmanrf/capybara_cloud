package masbro_worker

import (
	"github.com/moby/moby/api/types/image"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/docker"
)

type StubDocker struct {
	pull_error error
	pull_n_calls int
	pull_call_args []string

	find_one_return *image.Summary
	find_one_error error
	find_one_n_calls int
	find_one_call_args []string

	push_error error
	push_n_calls int
	push_call_args []string

	run_return *docker.DockerRunResult
	run_error error
	run_n_calls int
	run_call_args []docker.DockerRunDto
}

func (sd *StubDocker) Clear() {
	sd.pull_error = nil
	sd.pull_n_calls = 0
	sd.pull_call_args = nil

	sd.find_one_return = nil
	sd.find_one_error = nil
	sd.find_one_n_calls = 0
	sd.find_one_call_args = nil

	sd.push_error = nil
	sd.push_n_calls = 0
	sd.push_call_args = nil

	sd.run_return = nil
	sd.run_error = nil
	sd.run_n_calls = 0
	sd.run_call_args = nil
}

func (sd *StubDocker) FindOneImageByName(name string) (*image.Summary, error) {
	sd.find_one_n_calls += 1
	sd.find_one_call_args = append(sd.find_one_call_args, name)

	return sd.find_one_return, sd.find_one_error
}

func (sd *StubDocker) Push(image_name string) error {
	sd.push_n_calls += 1
	sd.push_call_args = append(sd.push_call_args, image_name)

	return sd.push_error
}

func (sd *StubDocker) Pull(image_name string) error {
	sd.pull_n_calls += 1
	sd.pull_call_args = append(sd.pull_call_args, image_name)

	return sd.pull_error
}

func (sd *StubDocker) Run(dto docker.DockerRunDto) (*docker.DockerRunResult, error) {
	sd.run_n_calls += 1
	sd.run_call_args = append(sd.run_call_args, dto)

	return sd.run_return, sd.run_error
}
