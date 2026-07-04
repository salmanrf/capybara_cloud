package deployment

import "github.com/moby/moby/api/types/image"


type StubDocker struct {
	find_one_image_by_name_return *image.Summary
	find_one_image_by_name_error error
	find_one_image_by_name_return_n_calls int
	find_one_image_by_name_return_call_args []string

	push_return error
	push_return_n_calls int
	push_return_call_args []*image.Summary
}

func (s *StubDocker) Clear() {
	s.find_one_image_by_name_return = nil
	s.find_one_image_by_name_error = nil
	s.find_one_image_by_name_return_n_calls = 0
	s.find_one_image_by_name_return_call_args = nil

	s.push_return = nil
	s.push_return_n_calls = 0
	s.push_return_call_args = nil
}

func (s *StubDocker) FindOneImageByName(name string) (*image.Summary, error) {
	s.find_one_image_by_name_return_n_calls += 1
	s.find_one_image_by_name_return_call_args = append(s.find_one_image_by_name_return_call_args, name)
	
	return s.find_one_image_by_name_return, s.find_one_image_by_name_error
}

func (s *StubDocker) Push(img *image.Summary) error {
	s.push_return_n_calls += 1
	s.push_return_call_args = append(s.push_return_call_args, img)

	return s.push_return
}
