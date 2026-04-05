package utils

import (
	"fmt"

	"github.com/jaevor/go-nanoid"
)

//! This utility is intended for tests only!!!
//! Do not use it for large multipart/form-data payload, or for uploading real files

type MultipartFile struct {
	Filename string
	Filesize int
	Mimetype string
	Content []byte
}

type MultiPartField struct {
	Type string
	Name string
	Value string
	File MultipartFile
}

type multipart struct {
	boundaryStr string
	contentLength int
	data map[string]*MultiPartField
	raw string
	encoded []byte
}

type MultipartForm interface {
	SetField(name string, value string)
	SetFile(fieldname, filename string, filesize int, mimetype string)
	GetContentLength() int
	GetRawString() string
	GetEncoded() []byte
	GetHeaderContentType() string
	GetHeaderContentLength() string
}

func NewMultipartForm() MultipartForm {
	canonicID, err := nanoid.Standard(8)
	if err != nil {
		panic(err)
	}
	
	boundary := canonicID()
	data := make(map[string]*MultiPartField)
	
	return &multipart{
		boundaryStr: boundary,
		contentLength: 0,
		data: data,
		encoded: []byte{},
	}
}

func (f *multipart) SetFile(fieldname, filename string, filesize int, mimetype string) {
	bts := make([]byte, filesize)

	field := &MultiPartField{
		Type: "string",
		Name: fieldname,
		Value: "",
		File: MultipartFile{
			Filename: filename,
			Filesize: filesize,
			Mimetype: mimetype,
			Content: bts,
		},
	}
	f.data[fieldname] = field

	newdatastr := fmt.Sprintf("--%s\r\n", f.boundaryStr)
	newdatastr += fmt.Sprintf(
		"Content-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n"+
		"Content-Type: %s\r\n" +
		"Content-Length: %d\r\n" +
		"\r\n" +
		"%x\r\n",
		fieldname,
		filename,
		mimetype,
		filesize,
		bts,
	)

	f.raw += newdatastr
	f.contentLength += len([]byte(newdatastr))
}

func (f *multipart) SetField(key string, value string) {
	field := &MultiPartField{
		Type: "string",
		Name: key,
		Value: value,
	}
	f.data[key] = field

	newdatastr := fmt.Sprintf("--%s\r\n", f.boundaryStr)
	newdatastr += fmt.Sprintf(
		"Content-Disposition: form-data; name=\"%s\"\r\n"+
		"\r\n"+
		"%s\r\n",
		key,
		value,
	)

	f.raw += newdatastr
	f.contentLength += len([]byte(newdatastr))
}

func (f *multipart) finalize() {
	closingBoundary := fmt.Sprintf("--%s--\r\n", f.boundaryStr)
	f.raw += closingBoundary
	f.contentLength += len(closingBoundary)
}

func (f *multipart) GetContentLength() int {
	if f.contentLength == 0 && len(f.data) > 0 {
		f.finalize()
	}
	return f.contentLength
}

func (f *multipart) GetRawString() string {
	return f.raw
}

func (f *multipart) GetHeaderContentType() string {
	return fmt.Sprintf("multipart/form-data; boundary=%s", f.boundaryStr)
}

func (f *multipart) GetHeaderContentLength() string {
	return fmt.Sprintf("%d", f.contentLength)
}

func (f *multipart) GetEncoded() []byte {
	if len(f.encoded) == 0 && len(f.data) > 0 {
		f.finalize()
	}
	return []byte(f.raw)
}