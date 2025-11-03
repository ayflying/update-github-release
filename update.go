// Package update：GitHub Release 自更新模块
// - 检查最新版本并匹配当前平台资产（zip/tar.gz）
// - 解压到运行目录并设置权限
// - 可选重启：跨平台使用 exec.Command 启动新进程并退出旧进程
// - 非 Windows 平台启用新会话隔离，减少信号/控制台耦合
package update

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/encoding/gcompress"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcmd"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
)

type sUpdate struct {
	ApiURL    string
	IsRestart bool // 是否重启服务
}

// 本地版本号（建议从编译参数注入，如 -ldflags "-X main.version=v0.1.3"）
const versionFile = "version.txt"

var localVersion = "v0.0.0"

func New(url string) *sUpdate {
	return &sUpdate{
		ApiURL: url,
	}
}

// CheckUpdate 检查是否有新版本
// @param IsRestart 是否需要立即重启服务
func (s *sUpdate) CheckUpdate(IsRestart bool) (err error) {
	s.IsRestart = IsRestart
	ctx := gctx.New()
	latestVersion, assets, err := s.getLatestVersion()
	if err != nil {
		fmt.Printf("检查更新失败：%v\n", err)
		return
	}

	localVersion = gfile.GetContents(versionFile)

	if s.isNewVersion(localVersion, latestVersion) {
		g.Log().Printf(ctx, "发现新版本：%s（当前版本：%s）", latestVersion, localVersion)
		//拼接操作系统和架构（格式：OS_ARCH）
		platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
		//name := fmt.Sprintf("p2p_%s_%s.tar.gz", latestVersion, platform)
		for _, asset := range assets {
			if strings.Contains(fmt.Sprintf("_%s.", asset.Name), platform) {
				g.Log().Debugf(ctx, "下载链接：%s", asset.BrowserDownloadUrl)

				// 下载更新文件
				fileDownload, err2 := g.Client().Get(ctx, asset.BrowserDownloadUrl)
				if err2 != nil {
					return
				}
				updateFile := path.Join("download", asset.Name)
				err = gfile.PutBytes(updateFile, fileDownload.ReadAll())

				err = s.Update(ctx, updateFile)
				if err != nil {
					return
				}
				// 保存最新版本号到文件
				gfile.PutContents(versionFile, latestVersion)
				break
			}
		}
	} else {
		g.Log().Debugf(ctx, "当前已是最新版本：%s", localVersion)
	}
	return
}

func (s *sUpdate) Update(ctx context.Context, gzFile string) (err error) {
	//拼接操作系统和架构（格式：OS_ARCH）
	platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)

	runFile := gcmd.GetArg(0).String()
	oldFile, err := s.RenameRunningFile(runFile)
	g.Log().Debugf(ctx, "执行文件改名为%v", oldFile)
	if gzFile == "" {
		gzFile = path.Join("download", platform+".gz")
	}
	//结束后删除压缩包
	defer gfile.RemoveFile(gzFile)

	ext := gfile.Ext(gzFile)
	if ext == ".zip" {
		g.Log().Debugf(ctx, "zip解压%v到%v", gzFile, gfile.Dir(runFile))
		err = gcompress.UnZipFile(gzFile, gfile.Dir(runFile))
	} else {
		g.Log().Debugf(ctx, "gzip解压%v到%v", gzFile, gfile.Dir(runFile))
		err = s.UnTarGz(gzFile, gfile.Dir(runFile))
	}
	if err != nil {
		return
	}
	//修改文件权限为755
	err = gfile.Chmod(runFile, 0755)

	// 如果需要重启服务，5秒后重启
	if s.IsRestart {
		go func() {
			log.Println("5秒后开始重启...")
			time.Sleep(5 * time.Second)

			if err = s.RestartSelf(); err != nil {
				log.Fatalf("重启失败：%v", err)
			}
		}()
	}

	return
}

// UnTarGz 解压tar.gz文件到指定目录
func (s *sUpdate) UnTarGz(tarGzFileName, targetDir string) (err error) {
	// 打开tar.gz文件
	file, err := os.Open(tarGzFileName)
	if err != nil {
		return err
	}
	defer file.Close()

	// 创建gzip reader
	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	// 创建tar reader
	tr := tar.NewReader(gzr)

	// 遍历tar中的每个文件
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			// 到达文件末尾，退出循环
			break
		}
		if err != nil {
			return err
		}

		// 构建解压后的文件路径
		targetPath := targetDir + string(os.PathSeparator) + hdr.Name

		// 如果是目录，创建目录
		if hdr.Typeflag == tar.TypeDir {
			err := os.MkdirAll(targetPath, 0755)
			if err != nil {
				return err
			}
			continue
		}

		// 如果是文件，创建文件并写入内容
		outFile, err := os.Create(targetPath)
		if err != nil {
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, tr)
		if err != nil {
			return err
		}
	}

	return
}

// RestartSelf 实现 Windows 平台下的程序自重启
func (s *sUpdate) RestartSelf() error {
	// 跨平台统一使用 exec.Command 启动新进程并退出当前进程

	// 1. 获取当前程序的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	// 处理路径中的符号链接（确保路径正确）
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return err
	}

	// 2. 获取命令行参数（os.Args[0] 是程序名，实际参数从 os.Args[1:] 开始）
	args := os.Args[1:]

	// 3. 构建新进程命令（路径为当前程序，参数为原参数）
	cmd := exec.Command(exePath, args...)

	// 设置新进程的工作目录与当前进程一致
	if wd, err := os.Getwd(); err == nil {
		cmd.Dir = wd
	}

	// 新进程的输出继承当前进程的标准输出（按需保留或重定向）
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// 在非 Windows 平台上，将子进程置于新会话，减少与父进程信号/控制台耦合
	setCmdSysProcAttr(cmd)

	// 4. 启动新进程（非阻塞，Start() 后立即返回）
	if err := cmd.Start(); err != nil {
		return err
	}

	// 5. 新进程启动成功后，退出当前进程
	os.Exit(0)
	return nil // 理论上不会执行到这里
}

// RenameRunningFile 重命名正在运行的程序文件（如 message.exe → message.exe~）
func (s *sUpdate) RenameRunningFile(exePath string) (string, error) {
	// 目标备份文件名（message.exe → message.exe~）
	backupPath := exePath + "~"

	// 先删除已存在的备份文件（若有）
	if _, err := os.Stat(backupPath); err == nil {
		if err := os.Remove(backupPath); err != nil {
			return "", fmt.Errorf("删除旧备份文件失败: %v", err)
		}
	}

	// 重命名正在运行的 exe 文件
	// 关键：Windows 允许对锁定的文件执行重命名操作
	if err := os.Rename(exePath, backupPath); err != nil {
		return "", fmt.Errorf("重命名运行中文件失败: %v", err)
	}
	return backupPath, nil
}

// isNewVersion 判断最新版本是否比本地版本新（仅适用于 vX.Y.Z 格式）
// 返回 true 表示有新版本，false 表示无新版本或版本格式错误
func (s *sUpdate) isNewVersion(local, latest string) bool {
	// 1. 移除前缀 "v"，分割为字符串切片
	localParts := strings.Split(strings.TrimPrefix(local, "v"), ".")
	latestParts := strings.Split(strings.TrimPrefix(latest, "v"), ".")

	// 2. 取最长的长度作为循环次数（避免遗漏更长的版本号，如 1.2 vs 1.2.1）
	maxLen := len(localParts)
	if len(latestParts) > maxLen {
		maxLen = len(latestParts)
	}

	// 3. 逐段转换为整数并比较
	for i := 0; i < maxLen; i++ {
		// 获取当前段的数值（超出长度的部分按 0 处理）
		localNum := 0
		if i < len(localParts) {
			num, err := strconv.Atoi(localParts[i])
			if err != nil {
				// 版本号包含非数字，视为无效版本（不更新）
				return false
			}
			localNum = num
		}

		latestNum := 0
		if i < len(latestParts) {
			num, err := strconv.Atoi(latestParts[i])
			if err != nil {
				// 最新版本格式错误，视为无更新
				return false
			}
			latestNum = num
		}

		// 比较当前段
		if latestNum > localNum {
			return true // 最新版本更新
		} else if latestNum < localNum {
			return false // 本地版本更新
		}
		// 相等则继续比较下一段
	}

	// 所有段都相等，无新版本
	return false
}

func (s *sUpdate) getLatestVersion() (string, []*Assets, error) {

	resp, err := http.Get(s.ApiURL)
	if err != nil {
		return "", nil, fmt.Errorf("请求失败：%v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("API 响应错误：%d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", nil, fmt.Errorf("解析响应失败：%v", err)
	}

	return release.TagName, release.Assets, nil
}
