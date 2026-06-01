# geektime-downloader

`geektime-downloader` 是一个面向桌面端的极客时间下载工具，当前提供 GUI 图形界面，支持课程下载、参数配置、日志查看和 TS 视频转 MP4。

## 功能概览

- 图形界面首页已拆分为 4 个页面：
  - `下载操作`
  - `参数配置`
  - `日志查看`
  - `视频转换`
- 支持课程信息加载、下载全部、下载单篇
- 支持下载进度展示
- 支持下载任务暂停、继续、取消
- 支持运行日志查看、刷新、删除
- 支持扫描目录下所有 `.ts` 文件并批量转换为 `.mp4`

**极客时间**
- [x] 专栏(PDF/Markdown/音频)
- [x] 视频课
- [x] 每日一课
- [x] 大厂案例
- [x] 训练营视频
- [ ] 线下大会

**企业版极客时间**
- [ ] 体系课
- [ ] 每日一课
- [ ] 大厂案例
- [ ] 生态课
- [x] 训练营视频

部分资源暂未支持下载，欢迎PR。

[![go report card](https://goreportcard.com/badge/github.com/nicoxiang/geektime-downloader "go report card")](https://goreportcard.com/report/github.com/nicoxiang/geektime-downloader)
[![MIT license](https://img.shields.io/badge/license-MIT-brightgreen.svg)](https://opensource.org/licenses/MIT)

## 使用方式

### 运行前提

- Chrome installed
- FFmpeg installed

说明：

- 下载视频课程后，程序会使用 FFmpeg 将 `.ts` 自动转换为 `.mp4`
- `视频转换` 页面也依赖 FFmpeg 完成批量转换

#### 安装 FFmpeg

<details>
<summary>Windows</summary>

1. 前往 [FFmpeg 官网下载页](https://ffmpeg.org/download.html)，点击 Windows 图标，选择 `gyan.dev` 链接
2. 下载 `ffmpeg-release-essentials.zip`
3. 解压到任意目录，例如 `C:\ffmpeg`
4. 将 FFmpeg 的 `bin` 目录添加到系统环境变量 `PATH` 中：
   - 右键"此电脑" → 属性 → 高级系统设置 → 环境变量
   - 在"系统变量"中找到 `Path`，点击编辑 → 新建
   - 添加 `C:\ffmpeg\bin`（根据实际解压路径修改）
   - 点击确定保存
5. 打开新的命令行窗口，验证安装：
   ```bash
   ffmpeg -version
   ```

**或者使用包管理器安装：**

```bash
# 使用 winget
winget install Gyan.FFmpeg

# 使用 Chocolatey
choco install ffmpeg

# 使用 Scoop
scoop install ffmpeg
```

</details>

<details>
<summary>macOS</summary>

```bash
# 使用 Homebrew（推荐）
brew install ffmpeg

# 验证安装
ffmpeg -version
```

</details>

<details>
<summary>Linux</summary>

**Ubuntu / Debian：**

```bash
sudo apt update
sudo apt install ffmpeg

# 验证安装
ffmpeg -version
```

**CentOS / RHEL / Fedora：**

```bash
# Fedora
sudo dnf install ffmpeg

# CentOS / RHEL（需要 RPM Fusion 仓库）
sudo yum install epel-release
sudo yum install https://mirrors.rpmfusion.org/free/el/rpmfusion-free-release-$(rpm -E %rhel).noarch.rpm
sudo yum install ffmpeg

# 验证安装
ffmpeg -version
```

**Arch Linux：**

```bash
sudo pacman -S ffmpeg

# 验证安装
ffmpeg -version
```

</details>

## 编译

当前项目已经切换为 GUI 应用入口：

- 入口文件：`main.go`
- GUI 代码目录：`internal/uiapp`

Windows 下可以直接使用项目内置脚本编译：

```powershell
.\build.ps1
```

或：

```bat
build.bat
```

编译产物默认输出到：

```text
bin\geektime-downloader.exe
```

如果编译时提示找不到 C 编译器，请先安装并配置可用的 `gcc`。

## 页面说明

### 下载操作

- 选择产品类型和课程 ID
- 加载课程信息
- 下载全部内容
- 下载选中文章
- 查看当前课程下载进度
- 暂停、继续、取消下载任务

### 参数配置

- 配置 `GCID`、`GCESS`
- 配置下载目录
- 配置视频清晰度
- 配置评论模式
- 配置输出格式：`PDF`、`Markdown`、`音频`
- 配置 PDF 等待时间和超时时间
- 配置下载间隔
- 配置日志级别
- 切换企业版模式
- 保存配置

### 日志查看

- 按 `widget.List` 方式逐行显示日志
- 手动刷新日志
- 删除日志文件
- 仅展示日志尾部内容，避免日志过大拖慢界面

### 视频转换

- 选择一个目录
- 递归扫描目录下全部 `.ts` 文件
- 展示待处理视频列表
- 批量转换为 `.mp4`
- 如果存在同名 `.mp4`，直接删除对应 `.ts`
- 如果某个文件转换失败，会跳过当前文件继续处理后续文件
- 转换过程会写入日志，可在 `日志查看` 页面查看

### 文件下载目标位置

默认情况下：

- Windows 位于 `%USERPROFILE%/geektime-downloader`
- macOS / Linux 位于 `$HOME/geektime-downloader`

### 如何查看课程 ID?

**普通课程：**

打开极客时间[课程列表页](https://time.geekbang.org/resource)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL 最后的数字，例如下面的链接中 100056701 就是课程 ID：

```
https://time.geekbang.org/column/intro/100056701
```

**训练营课程：**

打开极客时间[训练营课程列表页](https://u.geekbang.org/schedule)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```lesson/```后的数字，例如下面的链接中 419 就是课程 ID：

```
https://u.geekbang.org/lesson/419?article=535616
```

**每日一课课程：**

选择你想要下载的视频，查看 URL ```dailylesson/detail/```后的数字，例如下面的链接中 100122405 就是课程 ID：

```
https://time.geekbang.org/dailylesson/detail/100122405
```

**大厂案例课程：**

选择你想要下载的视频，查看 URL ```qconplus/detail/```后的数字，例如下面的链接中 100110494 就是课程 ID：

```
https://time.geekbang.org/qconplus/detail/100110494
```

**公开课课程：**

选择你想要下载的视频，查看 URL ```opencourse/intro/``` 或 ```opencourse/videointro/```后的数字，例如下面的链接中 100546701 就是课程 ID：

```
https://time.geekbang.org/opencourse/videointro/100546701
```

**其他：**

打开极客时间[我的课程-其他](https://time.geekbang.org/dashboard/course)，选择你想要查看的课程，在新打开的课程详情 Tab 页，查看 URL ```course/intro/``` 最后的数字，例如下面的链接中 100551201 就是课程 ID：

```
https://time.geekbang.org/course/intro/100551201
```

**企业版训练营：**

选择你想要查看的课程，查看 URL ```mall/product/```后的数字，例如下面的链接中 100618109 就是课程 ID：

```
https://b.geekbang.org/mall/product/100618109
```

## 常见问题

### 为什么我下载的PDF是空白页?
首先下载课程请保证VPN已关闭。在此前提下如果仍然出现空白页情况，说明后台Chrome网页加载速度较慢，可以尝试加大--print-pdf-wait参数，保证页面完全加载完成后再开始生成PDF。

### 为什么我下载PDF一直提示超时?
首先下载课程请保证VPN已关闭。在此前提下如果下载持续出现超时，有可能是因为课程章节图片等内容较多，生成速度慢，比如课程《AI 绘画核心技术与实战》中的部分章节，可以尝试加大--print-pdf-timeout参数，并耐心等待。

### 如何下载专栏的 Markdown 格式和文章音频?

默认情况下载专栏的输出内容只有 PDF，可以通过 --output 参数按需选择是否需要下载 Markdown 格式和文章音频。比如 --output 3 就是下载 PDF 和 Markdown；--output 6 就是下载 Markdown 和音频；--output 7 就是下载所有。

Markdown 格式虽然显示效果上不及 PDF，但优势为可以显示完整的代码块（PDF 代码块在水平方向太长时会有缺失）并保留了原文中的超链接。

现在部分新课程的专栏文章中会包含视频，如课程《Kubernetes 入门实战课》等，目前程序会自动下载文章所包含的视频，视频目录在文章所在目录的子目录 videos 下，此类文章PDF的下载会耗费更多时间，请耐心等待。

### 退出程序和继续下载

GUI 中可以直接关闭窗口。

如果下载过程被暂停、取消或中断，重新进入程序后可重新加载课程并继续处理未完成内容。
