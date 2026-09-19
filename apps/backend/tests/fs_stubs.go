package tests

import (
	"bufio"
	"os"
)

type File_utils_args struct {
	name string
	flag int
	perm os.FileMode
}

type File_utils_stub struct {
	Open_call_args []File_utils_args
	Open_call 		 int
	Open_return 	 *os.File
	Open_error 		 error
	Open_fn				 func (name string, flag int, perm os.FileMode) (*os.File, error)
}

func (fs *File_utils_stub) OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	fs.Open_call += 1
	args := File_utils_args{
		name,
		flag,
		perm,
	}
	fs.Open_call_args = append(fs.Open_call_args, args)

	if fs.Open_fn != nil {
		return fs.Open_fn(name, flag, perm)
	}
	
	return fs.Open_return, fs.Open_error
} 

func (fs *File_utils_stub) Clear() {
	fs.Open_call = 0
	fs.Open_call_args = []File_utils_args{}
	fs.Open_return = nil
	fs.Open_error = nil
}

func ReadFileLines(file *os.File) ([]string, error) {
	_, err := file.Seek(0, 0)
	if err != nil {
		return []string{}, err
	}
	
	scanner := bufio.NewScanner(file)

	var got_lines []string
	for scanner.Scan() {
		l := scanner.Text()
		got_lines = append(got_lines, l)
	}

	if err := scanner.Err(); err != nil {
		return []string{}, err
	}

	return got_lines, nil
}