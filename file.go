package gogular

import (
	"os"
	"bytes"
	"fmt"
	"io"
	"bufio"
)

type FileType int

const (
	HTML FileType = iota
	CSS
)

type File struct {
	FileType
	FileName    string
	FileDir     string
	TmpFilePath string
	TmpDir      string
	Version     int
}

func NewFile(fileType FileType, name string, dir string, tmpDir string) *File {
	return &File{fileType, name, dir, dir + "/" + name, tmpDir, 0}
}

func (f *File) UpdateFilePath(destDir string, originalName bool) (string, string) {
	oldPath := f.TmpFilePath
	if len(oldPath) == 0 {
		oldPath = f.FileName
	}

	if originalName {
		f.TmpFilePath = destDir + "/" + f.FileName
	} else {
		f.TmpFilePath = destDir + "/" + RandString(12) + ".txt"
	}

	f.Version++

	return oldPath, f.TmpFilePath
}

func (f *File) ReadFile(pw *io.PipeWriter) {
	osF := f.GetFile()
	writer := bufio.NewWriter(pw)
	writer.ReadFrom(osF)
	defer osF.Close()
}

func (f *File) WriteFile(destDir string, originalName bool, pr *io.PipeReader) {
	file := f.GetNewFile(destDir, originalName)
	writer := bufio.NewWriter(file)
	writer.ReadFrom(pr)
	defer file.Close()
}

func (f *File) GetNewFile(destDir string, originalName bool) *os.File {
	_, newPath := f.UpdateFilePath(destDir, originalName)
	osF, err := os.Create(newPath)
	if err != nil {
		fmt.Print(err)
	}
	return osF
}

func (f *File) GetFile() *os.File {
	osF, err := os.Open(f.TmpFilePath)
	if err != nil {
		fmt.Print(err)
	}
	return osF
}

func (f *File) OpenFileBuffer() *bytes.Buffer {
	osF := f.GetFile()
	buf := new(bytes.Buffer)
	buf.ReadFrom(osF)
	defer osF.Close()
	return buf
}

func (f *File) String() string {
	return f.OpenFileBuffer().String()
}
