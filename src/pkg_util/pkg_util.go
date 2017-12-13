package pkg_util

import (
	"fmt"
	"os"
	"zip_util"
	"time"
	"strings"
	"os/exec"
	"log"
	"io"
	"path/filepath"
	"errors"
)

var (

	LOCAL_PATH1 = os.TempDir() + "\\test_file.zip"			// 下载到本地的路径1
	LOCAL_PATH2 = "d:\\test_file.zip"						// 下载到本地的路径2
	UNZIP_PATH1 = os.TempDir() + "\\hbn_pkg\\"				// 更新包解压路径1
	UNZIP_PATH2 = "d:\\hbn_pkg\\"							// 更新包解压路径1

	//TOMCAT_PROCESS_PREFIX   = "tomcat" 						// tomcat进程前缀
	//TOMCAT_W_PROCESS_SUFFIX = "w.exe"  						// tomcatx进程后缀，形式：w.exe

	PKG_CFGFILE_PATH_ARR = []string{// 项目配置文件地址，
		"webapps\\agent\\WEB-INF\\log4j.properties",
		"webapps\\agent\\WEB-INF\\classes\\config.properties",
		"webapps\\agent\\WEB-INF\\classes\\openoffice.properties"}
)

type TomcatInfo struct {
	ProcessName         string // tomcat6.ext
	ProcessPath         string // d:\xx\tomcat6\bin\tomcat6.exe
	ProcessHome         string // d:\xx\tomcat6\
	PackageBackPath     string // 更新前项目备份路径
	PackageDir          string // 项目存放路径 d:\xx\tomcat6\webapps\agent
	PackageFileName     string // 更新包名称	agent_1114
	PackageBackFileName string // 备份包名称	201712081536更新前备份.zip
	ConfigFileBackupDir string // 配置文件临时目录d:\xx\tomcat6\temp_config
	NewPackageDir       string // 新包地址
	Update              bool   // 是否需要更新
	Complete            bool   // 更新完成
}

/**
备份tomcat目录下项目
 */
func BackupCurrentPackage(tomcatArr []*TomcatInfo) error  {

	var tempArr []*TomcatInfo;
	tempArr = tomcatArr;

	if(len(tempArr) == 0) {
		fmt.Println("当前系统未运行tomcat实例，无法更新");
		//os.Exit(1);
		var e = errors.New("当前系统未运行tomcat实例，无法更新")
		return e;
	}

	for i := range tempArr {

		var tomcatInfo = tempArr[i];
		if !tomcatInfo.Update {
			continue;
		}

		tomcatInfo.ConfigFileBackupDir = tomcatInfo.ProcessHome+ "pkg_cfg\\"

		_, stat_err := os.Stat(tomcatInfo.ProcessHome+ "pkg_cfg\\");

		if stat_err != nil && os.IsNotExist(stat_err) {
			var mkdir_err = os.MkdirAll(tomcatInfo.ProcessHome+ "pkg_cfg\\", 0777);
			if mkdir_err != nil {
				return stat_err;
			}
		} else if stat_err != nil {
			return stat_err;
		}

		tomcatWebappDirFile, err := os.Open(tomcatInfo.PackageDir);

		var tomcatWebappPath = []*os.File{tomcatWebappDirFile };

		if err != nil {
			fmt.Println(tomcatInfo.ProcessName + "备份失败")
			fmt.Println(err)
			os.Exit(1);
		}

		var tomcatBckupPath = tomcatInfo.PackageBackPath;
		_, stateErr := os.Stat(tomcatBckupPath)
		if stateErr != nil {
			direrr := os.Mkdir("" + tomcatBckupPath, 0777);
			//direrr := os.MkdirAll("D:\\Program Files\\Apache Software Foundation\\apache-tomcat-8.0.39\\backup", 0777)
			if direrr != nil {
				fmt.Println(direrr)
				fmt.Println("创建" + tomcatInfo.ProcessName + "备份目录失败");
				//os.Exit(1);
				return direrr;
			} else {
				fmt.Println("创建" + tomcatInfo.ProcessName + "备份目录成功");
			}
		}

		// 备份tomcat目录下当前项目
		// time.Now().Format("200601021504")
		var now = time.Now();
		var backupFileName = now.Format("200601021504") + "更新前备份";
		ziperr := zip_util.Zip(tomcatWebappPath, tomcatBckupPath + backupFileName + ".zip");
		if ziperr == nil {
			tomcatInfo.PackageBackFileName = backupFileName + ".zip";
			fmt.Println("创建" + tomcatInfo.ProcessName + "备份文件成功：" + tomcatInfo.PackageBackFileName);
		}

		// 备份tomcat目录下项目配置文件，如config.properties、log4j.xml、openoffice.properties
		for i := range PKG_CFGFILE_PATH_ARR {

			var cfgFilePath = tomcatInfo.ProcessHome + PKG_CFGFILE_PATH_ARR[i];
			var cfgFileName = strings.Split(PKG_CFGFILE_PATH_ARR[i], "\\")[len(strings.Split(PKG_CFGFILE_PATH_ARR[i], "\\")) - 1];	//配置文件名

			// 备份目录不存在则创建备份目录
			_, stateErr := os.Stat(tomcatInfo.ProcessHome + "pkg_cfg\\");
			if stateErr != nil {
				direrr := os.Mkdir(tomcatInfo.ProcessHome+ "pkg_cfg\\", 0777);
				if direrr != nil {
					fmt.Println(direrr)
					fmt.Println("创建" + tomcatInfo.ProcessName + "项目配置文件备份目录失败");
					os.Exit(1);
				} else {
					fmt.Println("创建" + tomcatInfo.ProcessName + "项目配置文件备份目录成功");
				}
			}

			written, copyErr := copyFile(tomcatInfo.ProcessHome+ "pkg_cfg\\" + cfgFileName, cfgFilePath);
			if copyErr != nil {
				fmt.Println(copyErr)
				fmt.Println("备份" + tomcatInfo.ProcessName + "配置文件失败");
				os.Exit(1);
			}

			fmt.Println("复制" + tomcatInfo.ProcessName+ "配置文件成功，文件：" + cfgFileName + "，大小：", written, "byte");
		}

	}

	//fmt.Println(tomcatArr);
	return nil;
}

/**
获取tomcat信息
 */
func GetTomcatArray(
	tomcatPrefix string,
	tomcatSuffix string) [] TomcatInfo {

	//out, err := exec.Command("cmd", "/C", "tasklist ").Output()
	out, err := exec.Command("cmd", "/C", "tasklist").Output()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf(string(out))

	var processStrList[] string = strings.Split(string(out), "\r\n");
	var tomcatArr []TomcatInfo;

	for i := range processStrList {

		if(strings.HasPrefix(strings.ToLower(processStrList[i]), tomcatPrefix)) {
			//fmt.Println(i)
			//fmt.Println(processStrList[i])

			var processName = strings.Split(processStrList[i], " ")[0];
			if ! strings.HasSuffix(processName, tomcatSuffix) {
				out2, err2 := exec.Command("cmd", "/C", "wmic process where name='" + processName + "' get ExecutablePath").Output()
				if err2 == nil {

					// TODO
					var fileDirectoryArr[] string  = strings.Split(strings.Split(string(out2), "\r\n", )[1], "\\");
					if(len(fileDirectoryArr) < 2) {
						continue;
					}

					var parentDirectoryArr = fileDirectoryArr[0: len(fileDirectoryArr) - 2];

					var tomcatInfo TomcatInfo;
					tomcatInfo.ProcessName = processName;
					tomcatInfo.ProcessHome = strings.Join(parentDirectoryArr, "\\") + "\\";
					tomcatInfo.ProcessPath = tomcatInfo.ProcessHome + "bin\\" + processName;
					tomcatInfo.PackageBackPath = tomcatInfo.ProcessHome + "backup\\";
					tomcatInfo.PackageDir = tomcatInfo.ProcessHome + "webapps\\agent\\";

					tomcatArr = append(tomcatArr, tomcatInfo);

				} else {
					fmt.Println(err2)
				}
				//fmt.Println("------------------------------------------------------")
			}

		}
	}

	return tomcatArr;
	//fmt.Println(TOMCAT_PROCESS_MAP)
}

/**
确认tomcat
 */
func ConfirmTomcat(tomcatArr []* TomcatInfo) {

	var tempArr []*TomcatInfo;
	tempArr = tomcatArr;

	for i := range tempArr {

		var tomcatInfo = tempArr[i];
		if tomcatInfo == nil {
			continue;
		}

		// 当前tomcat是否需要更新
		for true {

			var update string
			fmt.Print("是否需要更新 " + tomcatInfo.ProcessName + "(0:否；1:是)： ");
			fmt.Scanln(&update);
			if(update == "1" || update == "0") {
				tomcatInfo.Update = (update == "1");
				break;
			}
			fmt.Print("输入有误，请重新输入0或者1！ ");
		}

		// 当前tomcat需要的更新包
		for tomcatInfo.Update {
			var pkg string
			fmt.Print(tomcatInfo.ProcessName + "需要哪个包进行更新(0:通用版最新包；1:安装包1114)： ");
			fmt.Scanln(&pkg);
			if(pkg == "0") {
				tomcatInfo.PackageFileName = "tyb";
				break;
			} else if(pkg == "1") {
				tomcatInfo.PackageFileName = "agent_1114";
				break;
			}
			fmt.Print("输入有误，请重新输入0或者1！ ");
		}
	}

	//fmt.Println("confirm complete.")
}

/**
拷贝文件
 */
func copyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer dst.Close()
	fmt.Println("拷贝文件:" + src.Name() + " > " + dst.Name());
	return io.Copy(dst, src)
}

func copyDir(src string, dest string) error {

	src_original := src;

	err := filepath.Walk(src, func(src string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {

			//fmt.Println(f.Name())
			//copyDir(f.Name(), dest+"/"+f.Name())

			if(src != src_original) {

				var temp_str = strings.Replace(src, src_original, dest, 1);
				os.MkdirAll(temp_str, 0777);

			}


		} else {
			//fmt.Println(src);
			//fmt.Println(src_original);
			//fmt.Println(dest);

			//fmt.Println("--------------------------------------------------------------------------------")

			dest_new := strings.Replace(src, src_original, dest, -1);
			//fmt.Println(dest_new);
			//fmt.Println("拷贝文件:" + src + " > " + dest_new);

			os.Create(dest_new);

			copyFile(dest_new, src);
		}
		//println(path)
		return nil
	})

	if err != nil {
		fmt.Printf("filepath.Walk() returned %v\n", err);
		return err;
	}

	return nil;
}

/**
替换包
 */
func ReplacePkg(tomcatInfo TomcatInfo) error {

	var stopErr = stopTomcat(tomcatInfo);
	if stopErr != nil {
		fmt.Println("停止" + tomcatInfo.ProcessName + "出错！");
	} else {
		fmt.Println("停止" + tomcatInfo.ProcessName + "成功。");
	}

	//var destDir = tomcatInfo.ProcessHome + "webapps\\agent";
	var destDir = tomcatInfo.ProcessHome + "webapps";

	// 删除webapps\agent
	rem_err := os.RemoveAll(destDir + "\\agent");
	if rem_err != nil {
		fmt.Println("移除目录出错：" + destDir + "\\agent");
		fmt.Println(rem_err)
		return rem_err;
	}

	copy_err := copyDir(tomcatInfo.NewPackageDir, destDir);
	if copy_err != nil {
		fmt.Println("拷贝目录出错：" + destDir);
		return copy_err;
	}

	// 还原配置文件
	for i := range PKG_CFGFILE_PATH_ARR {

		var cfgFilePath = tomcatInfo.ProcessHome + PKG_CFGFILE_PATH_ARR[i];
		var cfgFileName = strings.Split(PKG_CFGFILE_PATH_ARR[i], "\\")[len(strings.Split(PKG_CFGFILE_PATH_ARR[i], "\\")) - 1];	//配置文件名

		write_len, copy_err2 := copyFile(cfgFilePath, tomcatInfo.ConfigFileBackupDir + cfgFileName);
		if copy_err2 != nil || write_len == 0 {
			fmt.Println("还原配置文件出错：" + tomcatInfo.ConfigFileBackupDir + cfgFileName);
			return copy_err2;
		}
	}

	var startErr = startTomcat(tomcatInfo);
	if startErr != nil {
		fmt.Println("启动" + tomcatInfo.ProcessName + "出错！");
	} else {
		fmt.Println("启动" + tomcatInfo.ProcessName + "成功。");
	}
	return nil;
}

/**
停止tomcat
 */
func stopTomcat(tomcatInfo TomcatInfo) error {

	var processName = strings.Split(tomcatInfo.ProcessName, ".")[0]

	_, err := exec.Command("cmd", "/C", "net stop " + processName + " && taskkill /f /im " + tomcatInfo.ProcessHome).Output();
	if err != nil {
		return err;
	}
	return nil;
}

/**
启动tomcat
 */
func startTomcat(tomcatInfo TomcatInfo) error {

	var processName = strings.Split(tomcatInfo.ProcessName, ".")[0]
	_, err := exec.Command("cmd", "/C", "net start " + processName).Output();
	if err != nil {
		return err;
	}
	return nil;
}
