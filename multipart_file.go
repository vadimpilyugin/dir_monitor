package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
	"path"
)

type FileBody struct {
	header      *bytes.Buffer
	footer      *bytes.Buffer
	file        *os.File
	Reader      io.Reader
	ContentType string
	Length      int64
}

func multipartHeader(fn string) (*bytes.Buffer, string, string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	_, err := writer.CreateFormFile(PARAM_NAME, fn)
	if err != nil {
		return nil, "", "", err
	}
	return buf, writer.Boundary(), writer.FormDataContentType(), nil
}

func multipartFooter(bound string) *bytes.Buffer {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	writer.SetBoundary(bound)
	writer.Close()
	return buf
}

func (fb *FileBody) Read(p []byte) (int, error) {
	n, err := fb.file.Read(p)
	if err == io.EOF {
		fb.file.Close()
	}
	return n, err
}

func NewBody(dirPath, fn string) (*FileBody, error) {
	fb := &FileBody{}
	header, bound, contentType, err := multipartHeader(fn)
	if err != nil {
		return nil, err
	}
	fb.header = header
	fb.ContentType = contentType
	fb.footer = multipartFooter(bound)
	fb.file, err = os.Open(path.Join(dirPath, fn))
	if err != nil {
		return nil, err
	}
	fileStat, err := fb.file.Stat()
	if err != nil {
		return nil, err
	}
	fb.Reader = io.MultiReader(fb.header, fb, fb.footer)
	fb.Length = int64(fb.header.Len()) + int64(fb.footer.Len()) + fileStat.Size()
	return fb, nil
}
