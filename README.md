# update-github-release

一个用于从 GitHub Releases 检查并拉取最新版本、替换当前运行可执行文件并安全重启的 Go 自更新模块。支持 Windows 和 Linux，支持 `.zip` 与 `.tar.gz` 两种资产格式。

## 特性
- 自动对比 `version.txt` 中本地版本与 GitHub 最新 `tag_name`。
- 根据当前平台自动匹配下载资产（命名需包含 `_<os>_<arch>.`）。
- 支持解压 `.zip` 与 `.tar.gz`，并将新二进制替换到原路径。
- Windows/Linux 安全重启：统一使用 `exec.Command` 启动新进程并退出旧进程（非 Windows 平台启用新会话隔离）。

## 安装

- 运行：`go get github.com/ayflying/update-github-release`

## 快速开始

> 说明：本仓库现作为可引用库（包名 `update`），可在你的可执行项目中直接导入并调用。

```go
package main

import (
    "fmt"
    update "github.com/ayflying/update-github-release"
)

func main() {
    // 指向你的项目的 GitHub Releases 最新版本 API
    // 例如："https://api.github.com/repos/<owner>/<repo>/releases/latest"
    u := update.New("https://api.github.com/repos/<owner>/<repo>/releases/latest")

    if err := u.CheckUpdate(); err != nil {
        fmt.Println("自更新检查失败:", err)
    }

    // 你的业务逻辑 ...
}
```

### 本地版本文件
- 程序根目录需要一个 `version.txt`，记录当前版本号（例如：`v0.1.3`）。
- 代码会读取该文件内容作为本地版本进行对比；若不存在则默认为 `v0.0.0`。

### 发布资产命名约定
- 程序会根据当前平台拼接 `platform := <os>_<arch>`（如 `windows_amd64`、`linux_arm64`）。
- 检查资产名是否包含 `_<platform>.`，示例：
  - `myapp_v0.1.4_windows_amd64.zip`
  - `myapp_v0.1.4_linux_amd64.tar.gz`
- 这样就能针对当前平台自动选择正确的下载包。

### 压缩格式支持
- `.zip`：使用 `gcompress.UnZipFile` 解压到运行目录。
- `.tar.gz`：使用内置 `UnTarGz` 解压到运行目录。

### 自重启策略
- Windows：启动同一可执行的新进程（继承原参数），随后退出当前进程。
- Linux：与 Windows 相同，通过 `exec.Command` 启动新进程并退出旧进程（在非 Windows 平台使用新会话隔离，减少信号耦合）。

## 注意事项
- 更新过程会将当前正在运行的二进制文件重命名为同名加 `~` 的备份文件（例如：`message.exe` → `message.exe~`）。
- 更新完成后会为新二进制设置 `0755` 权限。
- 下载临时包位于 `download/` 目录，完成后会自动删除。
- 版本号格式建议遵循 `vX.Y.Z`，便于简化比较逻辑。

## 依赖
- Go 1.21+（`go.mod` 中为 1.24.8，按你的实际环境调整）
- `github.com/gogf/gf/v2`

## 许可
本项目采用 MIT 许可协议，详见 `LICENSE` 文件。