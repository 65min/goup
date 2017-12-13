package ftp_util

import (
	"fmt"
	"github.com/shenshouer/ftp4go"
)

/**
下载压缩包
 */
func DownloadFile(
	server string,
	username string,
	passwd string,
	filename string,
	tempdir string) error {

	ftpClient := ftp4go.NewFTP(0) // 1 for debugging

	//connect
	_, err := ftpClient.Connect(server, ftp4go.DefaultFtpPort, "")
	if err != nil {
		fmt.Println("FTP连接失败")
		//os.Exit(1)
		return err;
	}
	defer ftpClient.Quit()

	_, err = ftpClient.Login(username, passwd, "")
	if err != nil {
		fmt.Println("FTP登录失败")
		//os.Exit(1)
		return err;
	}

	size, err := ftpClient.Size(filename)
	if err != nil {
		fmt.Println("SIZE命令失败")
		//os.Exit(1)
		return err;
	}
	fmt.Println("开始下载更新包，大小： ", size/1024/1024, "MB")

	// start resume file download
	if err = ftpClient.DownloadResumeFile(filename, tempdir + filename, false); err != nil {
		fmt.Println("下载失败")
		//os.Exit(1)
		return err;
	}

	return nil;
}