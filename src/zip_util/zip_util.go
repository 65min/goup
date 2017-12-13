package zip_util

import (
	"os"
	"archive/zip"
	"io"
	"fmt"
	"path/filepath"
)

/**
压缩文件
files 文件数组，可以是不同dir下的文件或者文件夹
dest 压缩文件存放地址
 */
func Zip(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := doCompress(file, "", w)
		if err != nil {
			return err
		}
	}
	return nil
}

func doCompress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				return err
			}
			err = doCompress(f, prefix, zw)
			if err != nil {
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}


/*
解压
 */
func Unzip(zipFile, dest string) error {

	fmt.Println("开始解压更新包(" + zipFile + ")...")

	// 打开压缩包
	unzip_file, err := zip.OpenReader(zipFile)
	if err!=nil {
		fmt.Println("打开压缩包(" + zipFile + ")失败")
		return err;
	}

	os.RemoveAll(dest)
	//removeErr := os.RemoveAll(dest)
	//if removeErr != nil {
	//	fmt.Println("删除目标目录失败")
	//	return 1;
	//}

	makeErr := os.MkdirAll(dest, 0755);
	if makeErr != nil {
		fmt.Println("创建目标目录失败")
		return err;
	}

	// 循环解压zip文件
	for _,f := range unzip_file.File {
		rc,err := f.Open()
		if err!=nil {
			fmt.Println("压缩包(" + zipFile + ")损坏")
			return err;
		}
		path := filepath.Join(dest, f.Name)
		// 判断解压出的是文件还是目录
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			// 创建解压文件
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				fmt.Println("创建本地文件失败")
				return err;
			}
			// 写入本地
			_,err = io.Copy(f, rc)
			if err!=nil {
				if err!=io.EOF {
					fmt.Println("写入本地失败")
					return err;
				}
			}
			f.Close()
		}
	}
	fmt.Println("更新包(" + zipFile + ")解压完成")
	return nil
}
