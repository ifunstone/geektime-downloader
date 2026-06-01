package uiapp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/ffmpeg"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

type conversionFileEntry struct {
	TSPath  string
	MP4Path string
}

func (u *UI) buildConversionPage() fyne.CanvasObject {
	listSpacer := canvas.NewRectangle(nil)
	listSpacer.SetMinSize(fyne.NewSize(0, 320))
	return container.NewVBox(
		widget.NewLabel("TS 转 MP4"),
		widget.NewForm(
			widget.NewFormItem("视频目录", u.wrapConversionFolderPicker()),
		),
		container.NewHBox(u.scanTSButton, u.convertTSButton),
		u.conversionStatusLabel,
		u.conversionProgressBar,
		widget.NewLabel("待处理视频列表"),
		container.NewStack(listSpacer, u.conversionList),
	)
}

func (u *UI) wrapConversionFolderPicker() fyne.CanvasObject {
	chooseBtn := widget.NewButton("选择目录", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				u.showError(err)
				return
			}
			if uri == nil {
				return
			}
			u.conversionFolderEntry.SetText(filepath.Clean(uri.Path()))
		}, u.window)
	})
	return container.NewBorder(nil, nil, nil, chooseBtn, u.conversionFolderEntry)
}

func (u *UI) scanTSFiles() {
	root := strings.TrimSpace(u.conversionFolderEntry.Text)
	if root == "" {
		u.showError(fmt.Errorf("请选择视频目录"))
		return
	}

	info, err := os.Stat(root)
	if err != nil {
		u.showError(err)
		return
	}
	if !info.IsDir() {
		u.showError(fmt.Errorf("请选择有效目录"))
		return
	}

	files, err := findTSFiles(root)
	if err != nil {
		u.showError(err)
		return
	}

	u.conversionFiles = files
	u.conversionList.Refresh()
	u.conversionProgressBar.SetValue(0)
	if len(files) == 0 {
		u.conversionStatusLabel.SetText("当前目录下没有 ts 视频文件")
		u.statusLabel.SetText("未发现可转换视频")
		u.convertTSButton.Disable()
		return
	}

	u.conversionStatusLabel.SetText(fmt.Sprintf("已找到 %d 个 ts 视频文件", len(files)))
	u.statusLabel.SetText(fmt.Sprintf("待转换视频 %d 个", len(files)))
	u.convertTSButton.Enable()
}

func (u *UI) convertTSFiles(ctx context.Context) {
	if u.conversionRunning {
		return
	}
	if len(u.conversionFiles) == 0 {
		u.showError(fmt.Errorf("请先扫描 ts 视频文件"))
		return
	}

	u.conversionRunning = true
	logger.Init(u.cfg.LogLevel)
	logger.Infof("开始转换 ts 视频, 目录: %s, 文件数: %d", strings.TrimSpace(u.conversionFolderEntry.Text), len(u.conversionFiles))
	u.scanTSButton.Disable()
	u.convertTSButton.Disable()
	u.conversionProgressBar.SetValue(0)
	u.conversionStatusLabel.SetText("开始转换...")
	u.statusLabel.SetText("正在转换视频...")

	files := append([]conversionFileEntry(nil), u.conversionFiles...)
	go func() {
		total := len(files)
		successCount := 0
		failedCount := 0
		for i, file := range files {
			current := i + 1
			fyne.Do(func() {
				u.conversionStatusLabel.SetText(fmt.Sprintf("正在处理 %d/%d: %s", current, total, filepath.Base(file.TSPath)))
			})

			if err := processTSFile(file); err != nil {
				failedCount++
				logger.Errorf(err, "视频转换失败，跳过当前文件: %s", file.TSPath)
				progress := float64(current) / float64(total)
				fyne.Do(func() {
					u.conversionProgressBar.SetValue(progress)
					u.conversionStatusLabel.SetText(fmt.Sprintf("跳过失败文件 %d/%d: %s", current, total, filepath.Base(file.TSPath)))
					u.statusLabel.SetText(fmt.Sprintf("已跳过失败文件: %s", filepath.Base(file.TSPath)))
				})
				continue
			}
			successCount++
			logger.Infof("视频处理完成: %s", file.TSPath)

			progress := float64(current) / float64(total)
			fyne.Do(func() {
				u.conversionProgressBar.SetValue(progress)
				u.conversionStatusLabel.SetText(fmt.Sprintf("已完成 %d/%d: %s", current, total, filepath.Base(file.TSPath)))
			})
		}

		fyne.Do(func() {
			u.conversionRunning = false
			u.scanTSButton.Enable()
			u.convertTSButton.Enable()
			u.statusLabel.SetText(fmt.Sprintf("视频转换完成，成功 %d，失败 %d", successCount, failedCount))
			u.scanTSFiles()
			if len(u.conversionFiles) == 0 && failedCount == 0 {
				u.conversionStatusLabel.SetText("视频转换完成，已清理全部 ts 文件")
			} else {
				u.conversionStatusLabel.SetText(fmt.Sprintf("视频转换完成，成功 %d，失败 %d", successCount, failedCount))
			}
			u.conversionProgressBar.SetValue(1)
			logger.Infof("ts 视频转换结束, 成功: %d, 失败: %d", successCount, failedCount)
			u.refreshLogView()
		})
	}()
}

func findTSFiles(root string) ([]conversionFileEntry, error) {
	files := make([]conversionFileEntry, 0)
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".ts") {
			return nil
		}
		files = append(files, conversionFileEntry{
			TSPath:  path,
			MP4Path: strings.TrimSuffix(path, filepath.Ext(path)) + ffmpeg.MP4Extension,
		})
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func processTSFile(file conversionFileEntry) error {
	if _, err := os.Stat(file.MP4Path); err == nil {
		return os.Remove(file.TSPath)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	if _, err := ffmpeg.ConvertToMP4(file.TSPath, filepath.Dir(file.TSPath)); err != nil {
		return err
	}
	return os.Remove(file.TSPath)
}
