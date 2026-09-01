package deployment

import (
	"github.com/moby/moby/api/types/image"
	"github.com/salmanrf/capybara-cloud/packages/shared-go/docker"
)


type StubDocker struct {
	find_one_image_by_name_return *image.Summary
	find_one_image_by_name_error error
	find_one_image_by_name_return_n_calls int
	find_one_image_by_name_return_call_args []string

	push_return error
	push_return_n_calls int
	push_return_call_args []string

	pull_return error
	pull_return_n_calls int
	pull_return_call_args []string

	run_return *docker.DockerRunResult
	run_err error
	run_return_n_calls int
	run_return_call_args []docker.DockerRunDto
}

func (s *StubDocker) Clear() {
	s.find_one_image_by_name_return = nil
	s.find_one_image_by_name_error = nil
	s.find_one_image_by_name_return_n_calls = 0
	s.find_one_image_by_name_return_call_args = nil

	s.push_return = nil
	s.push_return_n_calls = 0
	s.push_return_call_args = nil

	s.pull_return = nil
	s.pull_return_n_calls = 0
	s.pull_return_call_args = nil

	s.run_return = nil
	s.run_return_n_calls = 0
	s.run_return_call_args = nil
}

func (s *StubDocker) FindOneImageByName(name string) (*image.Summary, error) {
	s.find_one_image_by_name_return_n_calls += 1
	s.find_one_image_by_name_return_call_args = append(s.find_one_image_by_name_return_call_args, name)
	
	return s.find_one_image_by_name_return, s.find_one_image_by_name_error
}

func (s *StubDocker) Push(img string) error {
	s.push_return_n_calls += 1
	s.push_return_call_args = append(s.push_return_call_args, img)

	return s.push_return
}

func (s *StubDocker) Pull(img string) error {
	s.pull_return_n_calls += 1
	s.pull_return_call_args = append(s.pull_return_call_args, img)

	return s.pull_return
}

func (s *StubDocker) Run(dto docker.DockerRunDto) (*docker.DockerRunResult, error) {
	s.run_return_n_calls += 1
	s.run_return_call_args = append(s.run_return_call_args, dto)

	return s.run_return, s.run_err
}
