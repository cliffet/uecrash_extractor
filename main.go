package main

import (
	"compress/zlib"
	"encoding/binary"
	"io"
	"os"
	"path"
)

var bi32 = make([]byte, 4)

type UecrashFileHeader struct {
	Version          []byte
	DirName          string
	FileName         string
	UncompressedSize int32
	FileCount        int32
}

func ReadStr(w io.ReadCloser) string {
	w.Read(bi32)
	strlen := int32(binary.LittleEndian.Uint32(bi32))

	bstr := make([]byte, strlen)

	w.Read(bstr)

	n := int(strlen)
	for i, v := range bstr {
		if v == '\x00' {
			n = i
			break
		}
	}
	//fmt.Printf("len:%d %x\n", strlen, bstr)

	return string(bstr[:n])
}

func (h *UecrashFileHeader) Read(w io.ReadCloser) {
	h.Version = make([]byte, 3)
	w.Read(h.Version)

	if h.Version[0] != 'C' ||
		h.Version[1] != 'R' ||
		h.Version[2] != '1' {
		//fmt.Printf("%c%c%c\n", h.Version[0], h.Version[1], h.Version[2])
		//panic("Version이 CR1이 아님")
	}

	h.DirName = ReadStr(w)
	h.FileName = ReadStr(w)

	w.Read(bi32)
	h.UncompressedSize = int32(binary.LittleEndian.Uint32(bi32))

	w.Read(bi32)
	h.FileCount = int32(binary.LittleEndian.Uint32(bi32))

	//fmt.Printf("Version:          %s\n", h.Version)
	//fmt.Printf("DirName:          %s\n", h.DirName)
	//fmt.Printf("FileName:         %s\n", h.FileName)
	//fmt.Printf("UncompressedSize: 0x%x (%d)\n", bi32, h.UncompressedSize)
	//fmt.Printf("FileCount:        0x%x (%d)\n", bi32, h.FileCount)
}

type UecrashFile struct {
	Index    int
	FileName string
	Data     []byte
}

func (f *UecrashFile) Read(w io.ReadCloser, index int, destpath string) {
	w.Read(bi32)
	f.Index = int(binary.LittleEndian.Uint32(bi32))

	//fmt.Printf("- Index: 0x%x (%d)  --------\n", bi32, f.Index)
	if f.Index != index {
		return
	}

	f.FileName = ReadStr(w)
	//fmt.Printf(" FileName: %s\n", f.FileName)

	w.Read(bi32)
	filelen := int32(binary.LittleEndian.Uint32(bi32))

	filePath := path.Join(destpath, f.FileName)
	//fmt.Printf(" dirname: %q\n", dirname)
	//fmt.Printf(" filePath: %q\n", filePath)
	err := os.MkdirAll(destpath, 0755)
	if err != nil {
		//fmt.Printf("Error: %s\n", err)
		return
	}

	writer, err := os.Create(filePath)
	if err != nil {
		//fmt.Printf("Error: %s\n", err)
		return
	}

	defer writer.Close()
	fb := make([]byte, filelen)
	io.ReadFull(w, fb)
	writer.Write(fb)

	//fmt.Printf(" FileLen: %d\n", filelen)
}

type UecrashZipFile struct {
	Header UecrashFileHeader
	Files  []UecrashFile
}

func (zf *UecrashZipFile) Read(w io.ReadCloser, destpath string) {
	zf.Header.Read(w)

	//fmt.Printf("zf.Header.FileCount: %d\n", zf.Header.FileCount)
	for i := 0; i < int(zf.Header.FileCount); i++ {
		var f UecrashFile

		f.Read(w, i, destpath)
		zf.Files = append(zf.Files, f)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func unzip(f *os.File, destpath string) {
	w, _ := zlib.NewReader(f)

	{
		var ZipFile UecrashZipFile
		//	var strlen int32

		ZipFile.Read(w, destpath)
	}

	w.Close()
}

func decompress_uecrash(filename string, destpath string) {
	f, err := os.Open(filename)
	check(err)

	unzip(f, destpath)
}

func main() {
	decompress_uecrash("uecrash", "path")
}
