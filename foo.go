package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

const (
	PARAM_NAME = "file"
)

func usage() {
	println("Usage: monitor [DIRECTORY] [POST_URL]")
	os.Exit(1)
}

func fileHeader(full_path string) (*bytes.Buffer, *bytes.Buffer, string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	_, err := writer.CreateFormFile("file", full_path)
	if err != nil {
		return nil, nil, "", err
	}
	header := &bytes.Buffer{}
	io.Copy(header, buf)
	buf.Reset()

	err = writer.Close()
	if err != nil {
		return nil, nil, "", err
	}
	footer := &bytes.Buffer{}
	io.Copy(footer, buf)

	return header, footer, writer.Boundary(), nil
}

func testMultipart(test bool, bound string) (*bytes.Buffer, error) {
	const full_path = "foo.bar"
	file, err := os.Open(full_path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.SetBoundary(bound)

	part, err := writer.CreateFormFile("file", full_path)
	if err != nil {
		return nil, err
	}

	if !test {
		_, err = io.Copy(part, file)
		if err != nil {
			return nil, err
		}

		err = writer.Close()
		if err != nil {
			return nil, err
		}
	}

	return body, nil
}

type FileBody struct {
	header *bytes.Buffer
	footer *bytes.Buffer
	file   *os.File
	Reader io.Reader
	Length int64
}

func multipartHeader(fullPath string) (*bytes.Buffer, string, error) {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	_, err := writer.CreateFormFile(PARAM_NAME, fullPath)
	if err != nil {
		return nil, "", err
	}
	return buf, writer.Boundary(), nil
}

func multipartFooter(bound string) *bytes.Buffer {
	buf := &bytes.Buffer{}
	writer := multipart.NewWriter(buf)
	writer.SetBoundary(bound)
	writer.Close()
	return buf
}

func NewBody(fullPath string) (*FileBody, error) {
	fb := &FileBody{}
	tmp, bound, err := multipartHeader(fullPath)
	if err != nil {
		return nil, err
	}
	fb.header = tmp
	fb.footer = multipartFooter(bound)
	fb.file, err = os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	fileStat, err := fb.file.Stat()
	if err != nil {
		return nil, err
	}
	fb.Reader = io.MultiReader(fb.header, fb.file, fb.footer)
	fb.Length = int64(fb.header.Len()) + int64(fb.footer.Len()) + fileStat.Size()
	return fb, nil
}

func (fb FileBody) Close() {
	fb.file.Close()
}

func compareBuffers(buf *bytes.Buffer, buf2 *bytes.Buffer) {
	for i := 0; i < buf.Len(); i++ {
		b1, _ := buf.ReadByte()
		b2, _ := buf2.ReadByte()
		if b1 != b2 {
			fmt.Println("Bytes not equal!")
		}
	}
}

func main() {
	if len(os.Args) < 1 {
		usage()
	}
	// buf, _ := testMultipart(false)
	// header, footer, bound, _ := fileHeader("foo.bar")
	header, bound, _ := multipartHeader("foo.bar")
	footer := multipartFooter(bound)

	fmt.Println("<Header>", header, "</Header>")
	fmt.Println("<Footer>", footer, "</Footer>")

	buf, _ := testMultipart(false, bound)
	const full_path = "foo.bar"
	file, _ := os.Open(full_path)
	defer file.Close()
	content := &bytes.Buffer{}
	io.Copy(content, file)
	fmt.Println("Content:", content)

	fmt.Println("<<>>")
	fmt.Println(buf)

	buf2 := &bytes.Buffer{}
	io.Copy(buf2, header)
	io.Copy(buf2, content)
	io.Copy(buf2, footer)
	fmt.Println(buf2)
	fmt.Println(buf2.Len())
	fmt.Println(buf.Len())
	compareBuffers(buf, buf2)

	buf3 := &bytes.Buffer{}
	fb, _ := NewBody(full_path)
	io.Copy(buf3, fb.Reader)
	fmt.Println("Buffer 3")
	fmt.Println(buf3)
}
