package main

import (
	"fmt"
	"os"
	"time"
	"math/rand"
	"pkg_util"
	"ftp_util"
	"zip_util"
	"strconv"
	"path/filepath"
	"log"
)


var (
	FTP_SERVER 				= "192.168.128.206"										// ftp地址
	FTP_USER 				= "*"													// ftp用户名
	FTP_PASSWD 				= "*"													// ftp密码

	BASE_FTP_PATH    		= "/" 												// base data path in ftp server

	LOCAL_TEMP_DIR       	= os.TempDir() + "\\"
	LOCAL_TEMP_UNZIP_DIR 	= os.TempDir() + "\\hbn_pkg\\"

	TOMCAT_PROCESS_PREFIX   = "tomcat" // tomcat进程前缀
	TOMCAT_W_PROCESS_SUFFIX = "w.exe"  // tomcatx进程后缀，形式：w.exe
)


func main() {

	// 临时目录
	initCfg();

	// 安装模式 0:离线模式；1:在线模式
	var mode = mode();

	var tomcatArr = pkg_util.GetTomcatArray(TOMCAT_PROCESS_PREFIX, TOMCAT_W_PROCESS_SUFFIX);

	var tempArr = make([]* pkg_util.TomcatInfo, len(tomcatArr));

	for i := range tomcatArr {
		tempArr[i] = &tomcatArr[i]
	}

	pkg_util.ConfirmTomcat(tempArr);


	// 几分钟后更新
	var minute = confirm();
	time.Sleep(time.Minute * time.Duration(minute))

	// 备份当前安装包
	var back_err = pkg_util.BackupCurrentPackage(tempArr);
	if back_err != nil {
		fmt.Println(back_err);
		os.Exit(1);
	}

	{
		for i := range tomcatArr {

			var tomcatInfo = tomcatArr[i];
			var zipFilePath string;

			if(mode == "1") {

				zipFilePath = LOCAL_TEMP_DIR + tomcatInfo.PackageFileName + ".zip";

				// 下载安装包
				// 安装包是否已存在，不存在时，则下载
				_, stat_err := os.Stat(LOCAL_TEMP_DIR + tomcatInfo.PackageFileName + ".zip");
				if stat_err != nil && os.IsNotExist(stat_err){

					fmt.Println("下载" + tomcatInfo.ProcessName + " -> " + tomcatInfo.PackageFileName + "安装包...")
					err := ftp_util.DownloadFile(
						FTP_SERVER,
						FTP_USER,
						FTP_PASSWD,
						"\\" + tomcatInfo.PackageFileName + ".zip",
						LOCAL_TEMP_DIR);
					if err != nil {
						fmt.Println("下载" + tomcatInfo.ProcessName + " -> " + tomcatInfo.PackageFileName + "失败！");
						fmt.Println("更新程序已中止。")
						os.Exit(1);
					}

				}
			} else {
				zipFilePath = getCurrentDirectory() + tomcatInfo.PackageFileName + ".zip";
			}

			// 路径不存在时，则解压
			var unzipPath = LOCAL_TEMP_UNZIP_DIR + tomcatInfo.PackageFileName;
			//var unzipPath = zipFilePath;
			_, stat_err2 := os.Stat(unzipPath);
			if stat_err2 != nil && os.IsNotExist(stat_err2){

				// 解压安装包
				unzip_err := zip_util.Unzip(zipFilePath, unzipPath);
				if unzip_err != nil {
					os.Exit(1);
				}
			}

			// 新包路径
			//tomcatInfo.NewPackageDir = LOCAL_TEMP_UNZIP_DIR + tomcatInfo.PackageFileName + "\\agent";
			tomcatInfo.NewPackageDir = LOCAL_TEMP_UNZIP_DIR + tomcatInfo.PackageFileName;

			pkg_util.ReplacePkg(tomcatInfo);
		}
	}

	fmt.Print("输入回车键退出当前更新程序：");
	var dummy string;
	fmt.Scanln(&dummy);
	os.RemoveAll(LOCAL_TEMP_DIR);
}

/**
初始化配置信息，包括临时目录等...
 */
func initCfg() {

	rand.Seed(int64(time.Now().Nanosecond()));
	LOCAL_TEMP_DIR = LOCAL_TEMP_DIR + strconv.FormatInt(rand.Int63(), 16) + "\\";
	//LOCAL_TEMP_DIR = LOCAL_TEMP_DIR + "77be30a6bb4be42e\\";
	LOCAL_TEMP_UNZIP_DIR = LOCAL_TEMP_DIR + "hbn_pkg\\"

	os.MkdirAll(LOCAL_TEMP_DIR, 0777);
	os.MkdirAll(LOCAL_TEMP_UNZIP_DIR, 0777);
	fmt.Println("临时目录：", LOCAL_TEMP_DIR);
}

/**
确认
 */
func confirm() int64 {

	var minute int64;
	fmt.Print("几分钟后执行更新任务：");

	for true {

		fmt.Scanln(&minute);
		if(minute >= 0 && minute <= 300) {
			break;
		}

		fmt.Print("请输入0到300之间的数字：");
	}

	fmt.Println(minute, " 分钟后(" + (time.Now().Add(time.Minute * time.Duration(minute)).Format("2006-01-02 15:04:05")) + ")将更新ERP系统，在此期间内请关闭360等安全软件，并不要关闭计算机。");
	//fmt.Println(time.Now().Add(time.Minute * time.Duration(minute)).Format("2006-01-02 15:04:05"));
	return minute;
}

/**
更新模式
 */
func mode() string {

	fmt.Print("请选择更新模式（0:离线模式；1:在线模式）：");
	var mode string;
	for true {

		fmt.Scanln(&mode);
		if(mode == "0" || mode == "1") {
			break;
		}

		fmt.Print("输入有误，请输入0或者1（0:离线模式；1:在线模式）：");
	}
	return mode;
}

/**
当前目录
 */
func getCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	return dir + "\\";
}