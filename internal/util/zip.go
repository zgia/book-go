package util

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func ZipFile(src, dst string) (err error) {
	archive, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer archive.Close()

	zipWriter := zip.NewWriter(archive)

	defer func() {
		// 检测一下是否成功关闭
		if err := zipWriter.Close(); err != nil {
			log.Fatalln(err)
		}
	}()

	srcInfo, _ := os.Lstat(src)

	// 下面来将文件写入 zw ，因为有可能会有很多个目录及文件，所以递归处理
	return filepath.Walk(src, func(path string, fi os.FileInfo, errBack error) (err error) {
		if errBack != nil {
			return errBack
		}

		// 通过文件信息，创建 zip 的文件信息
		fh, err := zip.FileInfoHeader(fi)
		if err != nil {
			return
		}

		// 压缩比
		fh.Method = zip.Deflate

		// 去除最外层目录
		if srcInfo.IsDir() {
			fh.Name = strings.ReplaceAll(path, src, "")
		} else {
			fh.Name = filepath.Base(src)
		}

		// 替换文件信息中的文件名
		fh.Name = strings.TrimPrefix(fh.Name, string(filepath.Separator))

		// 这步开始没有加，会发现解压的时候说它不是个目录
		if fi.IsDir() {
			fh.Name += "/"
		}

		// 写入文件信息，并返回一个 Write 结构
		w, err := zipWriter.CreateHeader(fh)
		if err != nil {
			return
		}

		// 检测，如果不是标准文件就只写入头信息，不写入文件数据到 w
		// 如目录，也没有数据需要写
		if !fh.Mode().IsRegular() {
			return nil
		}

		// 打开要压缩的文件
		fr, err := os.Open(path)
		if err != nil {
			return
		}
		defer fr.Close()

		// 将打开的文件 Copy 到 w
		_, err = io.Copy(w, fr)
		if err != nil {
			return
		}

		return nil
	})
}
