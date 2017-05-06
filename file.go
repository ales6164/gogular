package gogular

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"io"
	"bytes"
)

type TmpFile struct {
	Dir    string
	TmpDir string

	Filename  string
	Prefix    string
	Extension string

	TmpPath string
}

func (a *App) NewTempFile(path string) *TmpFile {
	pathArray := strings.Split(path, "/")
	dir := strings.Join(pathArray[:len(pathArray)-1], "/")
	filename := pathArray[len(pathArray)-1]
	extension := filename[strings.LastIndex(filename, "."):]
	bareName := filename[:len(extension)]

	//tempPath := a.TmpDirectory + "/" + bareName + "-" + RandString(12) + extension

	return &TmpFile{dir, a.TmpDirectory, filename, bareName, extension, path}
}

func (f *TmpFile) Create() *os.File {
	nf, err := ioutil.TempFile(f.TmpDir, f.Prefix+"-")
	if err != nil {
		fmt.Println(err)
	}
	f.TmpPath = nf.Name()
	return nf
}

func (f *TmpFile) Open() *os.File {
	osF, err := os.Open(f.TmpPath)
	if err != nil {
		fmt.Print(err)
	}

	return osF
}

func (f *TmpFile) GetBuffer() *bytes.Buffer {
	osF := f.Open()
	defer osF.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(osF)
	return buf
}

func (f *TmpFile) Copy(dest string) {
	osF := f.Open()
	newF, err := os.Create(dest + "/" + f.Filename)
	defer osF.Close()
	defer newF.Close()

	if err != nil {
		fmt.Println(err)
	}

	io.Copy(newF, osF)
}
