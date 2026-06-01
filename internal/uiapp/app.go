package uiapp

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/nicoxiang/geektime-downloader/internal/app"
	"github.com/nicoxiang/geektime-downloader/internal/config"
	"github.com/nicoxiang/geektime-downloader/internal/progress"
)

type UI struct {
	fyneApp fyne.App
	window  fyne.Window

	cfg appConfigState

	productOptions []app.ProductTypeOption
	selectedType   *app.ProductTypeOption
	selection      *app.SelectionResult

	productTypeSelect *ScrollableSelect
	articleSelect     *ScrollableSelect
	statusLabel       *widget.Label
	courseLabel       *widget.Label
	progressBar       *widget.ProgressBar
	progressDetailLabel *widget.Label
	logList           *widget.List
	tabs              *container.AppTabs
	conversionList    *widget.List
	loadButton        *widget.Button
	downloadAllButton *widget.Button
	downloadOneButton *widget.Button
	pauseButton       *widget.Button
	cancelButton      *widget.Button
	refreshLogButton  *widget.Button
	deleteLogButton   *widget.Button
	scanTSButton      *widget.Button
	convertTSButton   *widget.Button
	saveConfigButton  *widget.Button

	gcidEntry            *widget.Entry
	gcessEntry           *widget.Entry
	downloadFolderEntry  *widget.Entry
	qualitySelect        *ScrollableSelect
	commentsSelect       *ScrollableSelect
	outputPDFCheck       *widget.Check
	outputMarkdownCheck  *widget.Check
	outputAudioCheck     *widget.Check
	printPDFWaitEntry    *widget.Entry
	printPDFTimeoutEntry *widget.Entry
	intervalEntry        *widget.Entry
	enterpriseCheck      *widget.Check
	logLevelSelect       *ScrollableSelect
	productIDEntry       *widget.Entry
	conversionFolderEntry *widget.Entry
	conversionProgressBar *widget.ProgressBar
	conversionStatusLabel *widget.Label

	taskMu       sync.Mutex
	activeTask   *downloadTaskState
	activeCancel context.CancelFunc
	lastLogText  string
	logLines     []string
	logTabActive bool
	conversionFiles   []conversionFileEntry
	conversionRunning bool
}

type appConfigState struct {
	Gcid                   string
	Gcess                  string
	DownloadFolder         string
	Quality                string
	DownloadComments       string
	PrintPDFWaitSeconds    string
	PrintPDFTimeoutSeconds string
	Interval               string
	IsEnterprise           bool
	LogLevel               string
	OutputPDF              bool
	OutputMarkdown         bool
	OutputAudio            bool
	ProductID              string
	ProductType            string
}

type downloadTaskKind int

const (
	taskNone downloadTaskKind = iota
	taskDownloadAll
	taskDownloadOne
)

type downloadTaskState struct {
	kind         downloadTaskKind
	selection    *app.SelectionResult
	articleIndex int
	paused       bool
	running      bool
}

func Run(ctx context.Context) {
	ui := newUI()
	ui.build(ctx)
	ui.window.ShowAndRun()
}

func newUI() *UI {
	defaultCfg := app.DefaultConfig()
	a := fyneapp.NewWithID("github.com.nicoxiang.geektime-downloader")
	w := a.NewWindow("Geektime Downloader")
	w.Resize(fyne.NewSize(1080, 820))

	ui := &UI{
		fyneApp: a,
		window:  w,
		cfg: appConfigState{
			DownloadFolder:         defaultCfg.DownloadFolder,
			Quality:                defaultCfg.Quality,
			DownloadComments:       strconv.Itoa(defaultCfg.DownloadComments),
			PrintPDFWaitSeconds:    strconv.Itoa(defaultCfg.PrintPDFWaitSeconds),
			PrintPDFTimeoutSeconds: strconv.Itoa(defaultCfg.PrintPDFTimeoutSeconds),
			Interval:               strconv.Itoa(defaultCfg.Interval),
			LogLevel:               defaultCfg.LogLevel,
			OutputPDF:              true,
		},
	}

	if settings, err := loadSettings(); err == nil {
		ui.applySettings(settings)
	}
	return ui
}

func (u *UI) build(ctx context.Context) {
	u.statusLabel = widget.NewLabel("等待操作")
	u.statusLabel.Wrapping = fyne.TextWrapWord
	u.courseLabel = widget.NewLabel("尚未加载课程")
	u.courseLabel.Wrapping = fyne.TextWrapWord
	u.progressBar = widget.NewProgressBar()
	u.progressBar.Min = 0
	u.progressBar.Max = 1
	u.progressBar.SetValue(0)
	u.progressBar.Hide()
	u.progressDetailLabel = widget.NewLabel("当前无下载任务")
	u.progressDetailLabel.Wrapping = fyne.TextWrapWord
	u.logList = widget.NewList(
		func() int {
			return len(u.logLines)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Importance = widget.WarningImportance
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.logLines) {
				obj.(*widget.Label).SetText("")
				return
			}
			label := obj.(*widget.Label)
			label.TextStyle = fyne.TextStyle{Monospace: true}
			label.Importance = widget.WarningImportance
			label.SetText(u.logLines[id])
		},
	)
	u.conversionFolderEntry = widget.NewEntry()
	u.conversionFolderEntry.SetPlaceHolder("选择包含 ts 视频文件的目录")
	u.conversionProgressBar = widget.NewProgressBar()
	u.conversionProgressBar.Min = 0
	u.conversionProgressBar.Max = 1
	u.conversionStatusLabel = widget.NewLabel("请选择目录并扫描 ts 文件")
	u.conversionStatusLabel.Wrapping = fyne.TextWrapWord
	u.conversionList = widget.NewList(
		func() int {
			return len(u.conversionFiles)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Wrapping = fyne.TextWrapWord
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < 0 || id >= len(u.conversionFiles) {
				obj.(*widget.Label).SetText("")
				return
			}
			file := u.conversionFiles[id]
			obj.(*widget.Label).SetText(fmt.Sprintf("%s\n-> %s", file.TSPath, file.MP4Path))
		},
	)

	u.gcidEntry = widget.NewPasswordEntry()
	u.gcidEntry.SetPlaceHolder("输入 GCID")
	u.gcidEntry.SetText(u.cfg.Gcid)
	u.gcessEntry = widget.NewPasswordEntry()
	u.gcessEntry.SetPlaceHolder("输入 GCESS")
	u.gcessEntry.SetText(u.cfg.Gcess)
	u.downloadFolderEntry = widget.NewEntry()
	u.downloadFolderEntry.SetText(u.cfg.DownloadFolder)
	u.productIDEntry = widget.NewEntry()
	u.productIDEntry.SetPlaceHolder("输入课程 ID")
	u.productIDEntry.SetText(u.cfg.ProductID)

	u.qualitySelect = NewScrollableSelect([]string{"ld", "sd", "hd"}, func(value string) {
		u.cfg.Quality = value
	})
	u.qualitySelect.SetSelected(u.cfg.Quality)

	u.commentsSelect = NewScrollableSelect([]string{"0", "1", "2"}, func(value string) {
		u.cfg.DownloadComments = value
	})
	u.commentsSelect.SetSelected(u.cfg.DownloadComments)

	u.logLevelSelect = NewScrollableSelect([]string{"debug", "info", "warn", "error", "none"}, func(value string) {
		u.cfg.LogLevel = value
	})
	u.logLevelSelect.SetSelected(u.cfg.LogLevel)

	u.outputPDFCheck = widget.NewCheck("PDF", func(value bool) {
		u.cfg.OutputPDF = value
	})
	u.outputPDFCheck.SetChecked(u.cfg.OutputPDF)
	u.outputMarkdownCheck = widget.NewCheck("Markdown", func(value bool) {
		u.cfg.OutputMarkdown = value
	})
	u.outputMarkdownCheck.SetChecked(u.cfg.OutputMarkdown)
	u.outputAudioCheck = widget.NewCheck("音频", func(value bool) {
		u.cfg.OutputAudio = value
	})
	u.outputAudioCheck.SetChecked(u.cfg.OutputAudio)

	u.printPDFWaitEntry = widget.NewEntry()
	u.printPDFWaitEntry.SetText(u.cfg.PrintPDFWaitSeconds)
	u.printPDFTimeoutEntry = widget.NewEntry()
	u.printPDFTimeoutEntry.SetText(u.cfg.PrintPDFTimeoutSeconds)
	u.intervalEntry = widget.NewEntry()
	u.intervalEntry.SetText(u.cfg.Interval)

	u.enterpriseCheck = widget.NewCheck("企业版", func(value bool) {
		u.cfg.IsEnterprise = value
		u.refreshProductOptions()
	})
	u.enterpriseCheck.SetChecked(u.cfg.IsEnterprise)

	u.productTypeSelect = NewScrollableSelect(nil, func(value string) {
		for i := range u.productOptions {
			if u.productOptions[i].Text == value {
				u.selectedType = &u.productOptions[i]
				break
			}
		}
	})

	u.articleSelect = NewScrollableSelect(nil, nil)
	u.articleSelect.PlaceHolder = "课程文章列表"

	u.loadButton = widget.NewButton("加载课程", func() {
		u.loadCourse(ctx)
	})
	u.downloadAllButton = widget.NewButton("下载全部", func() {
		u.downloadAll(ctx)
	})
	u.downloadOneButton = widget.NewButton("下载选中文章", func() {
		u.downloadSelectedArticle(ctx)
	})
	u.pauseButton = widget.NewButton("暂停下载", func() {
		u.togglePauseDownload(ctx)
	})
	u.cancelButton = widget.NewButton("取消下载", func() {
		u.cancelDownload()
	})
	u.refreshLogButton = widget.NewButton("刷新日志", func() {
		u.refreshLogView()
	})
	u.deleteLogButton = widget.NewButton("删除日志", func() {
		u.confirmDeleteLog()
	})
	u.scanTSButton = widget.NewButton("扫描 ts 文件", func() {
		u.scanTSFiles()
	})
	u.convertTSButton = widget.NewButton("转换视频", func() {
		u.convertTSFiles(ctx)
	})
	u.saveConfigButton = widget.NewButton("保存配置", func() {
		if err := u.persistSettings(); err != nil {
			u.showError(err)
			return
		}
		u.statusLabel.SetText("配置已保存")
	})
	u.downloadAllButton.Disable()
	u.downloadOneButton.Disable()
	u.pauseButton.Disable()
	u.cancelButton.Disable()
	u.convertTSButton.Disable()

	u.refreshProductOptions()
	u.restoreSelectedProductType()
	u.refreshLogView()

	downloadPage := container.NewPadded(u.buildDownloadPage())
	settingsPage := container.NewPadded(u.buildSettingsPage())
	logPage := container.NewPadded(u.buildLogPage())
	conversionPage := container.NewPadded(u.buildConversionPage())

	u.tabs = container.NewAppTabs(
		container.NewTabItem("下载操作", downloadPage),
		container.NewTabItem("参数配置", settingsPage),
		container.NewTabItem("日志查看", logPage),
		container.NewTabItem("视频转换", conversionPage),
	)
	u.tabs.SetTabLocation(container.TabLocationTop)
	u.tabs.OnSelected = func(item *container.TabItem) {
		u.logTabActive = item != nil && item.Text == "日志查看"
		if u.logTabActive {
			u.refreshLogView()
		}
	}

	content := container.NewBorder(
		nil,
		container.NewVBox(widget.NewSeparator(), u.statusLabel),
		nil,
		nil,
		u.tabs,
	)

	u.window.SetContent(content)
	if current := u.tabs.Selected(); current != nil && current.Text == "日志查看" {
		u.logTabActive = true
	}
	u.window.SetCloseIntercept(func() {
		_ = u.persistSettings()
		u.window.Close()
	})
	u.startLogAutoRefresh(ctx)
}

func (u *UI) buildDownloadPage() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("课程下载"),
		widget.NewForm(
			widget.NewFormItem("产品类型", u.productTypeSelect),
			widget.NewFormItem("课程 ID", u.productIDEntry),
		),
		container.NewHBox(u.loadButton, u.downloadAllButton, u.downloadOneButton),
		container.NewHBox(u.pauseButton, u.cancelButton),
		u.courseLabel,
		u.progressBar,
		u.progressDetailLabel,
		widget.NewLabel("文章选择"),
		u.articleSelect,
	)
}

func (u *UI) buildSettingsPage() fyne.CanvasObject {
	return container.NewVBox(
		widget.NewLabel("下载参数配置"),
		widget.NewForm(
			widget.NewFormItem("GCID", u.gcidEntry),
			widget.NewFormItem("GCESS", u.gcessEntry),
			widget.NewFormItem("下载目录", u.wrapFolderPicker()),
			widget.NewFormItem("视频清晰度", u.qualitySelect),
			widget.NewFormItem("评论模式", u.commentsSelect),
			widget.NewFormItem("输出格式", container.NewHBox(u.outputPDFCheck, u.outputMarkdownCheck, u.outputAudioCheck)),
			widget.NewFormItem("PDF等待秒数", u.printPDFWaitEntry),
			widget.NewFormItem("PDF超时秒数", u.printPDFTimeoutEntry),
			widget.NewFormItem("下载间隔秒数", u.intervalEntry),
			widget.NewFormItem("日志级别", u.logLevelSelect),
			widget.NewFormItem("", u.enterpriseCheck),
		),
		container.NewHBox(u.saveConfigButton),
	)
}

func (u *UI) buildLogPage() fyne.CanvasObject {
	return container.NewBorder(
		widget.NewLabel("运行日志"),
		container.NewHBox(u.refreshLogButton, u.deleteLogButton),
		nil,
		nil,
		u.logList,
	)
}

func (u *UI) wrapFolderPicker() fyne.CanvasObject {
	chooseBtn := widget.NewButton("选择目录", func() {
		dialog.ShowFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				u.showError(err)
				return
			}
			if uri == nil {
				return
			}
			u.downloadFolderEntry.SetText(filepath.Clean(uri.Path()))
		}, u.window)
	})
	return container.NewBorder(nil, nil, nil, chooseBtn, u.downloadFolderEntry)
}

func (u *UI) refreshProductOptions() {
	u.productOptions = app.ProductTypeOptions(u.cfg.IsEnterprise)
	names := make([]string, 0, len(u.productOptions))
	for i := range u.productOptions {
		names = append(names, u.productOptions[i].Text)
	}
	u.productTypeSelect.SetOptions(names)
	if len(names) > 0 {
		u.productTypeSelect.SetSelected(names[0])
	}
	u.selection = nil
	u.articleSelect.ClearSelected()
	u.articleSelect.SetOptions(nil)
	u.downloadAllButton.Disable()
	u.downloadOneButton.Disable()
	u.courseLabel.SetText("尚未加载课程")
}

func (u *UI) restoreSelectedProductType() {
	if u.cfg.ProductType == "" {
		return
	}
	for _, option := range u.productOptions {
		if option.Text == u.cfg.ProductType {
			u.productTypeSelect.SetSelected(option.Text)
			return
		}
	}
}

func (u *UI) loadCourse(ctx context.Context) {
	productID, err := strconv.Atoi(strings.TrimSpace(u.productIDEntry.Text))
	if err != nil {
		u.showError(fmt.Errorf("课程 ID 格式不合法"))
		return
	}
	if u.selectedType == nil {
		u.showError(fmt.Errorf("请选择产品类型"))
		return
	}

	svc, err := u.newService(ctx)
	if err != nil {
		u.showError(err)
		return
	}

	u.setBusy(true, "正在加载课程信息...")
	go func() {
		result, err := svc.ResolveProduct(*u.selectedType, productID)
		fyne.Do(func() {
			u.setBusy(false, "")
			if err != nil {
				u.showError(err)
				return
			}
			u.selection = result
			_ = u.persistSettings()
			u.courseLabel.SetText(fmt.Sprintf("课程：%s，共 %d 篇", result.Course.Title, len(result.Course.Articles)))
			u.downloadAllButton.Enable()
			if result.IsDirectMode {
				u.articleSelect.SetOptions(nil)
				u.downloadOneButton.Disable()
				return
			}
			articleTitles := make([]string, 0, len(result.Course.Articles))
			for _, article := range result.Course.Articles {
				articleTitles = append(articleTitles, article.Title)
			}
			u.articleSelect.SetOptions(articleTitles)
			if len(articleTitles) > 0 {
				u.articleSelect.SetSelected(articleTitles[0])
				u.downloadOneButton.Enable()
			}
		})
	}()
}

func (u *UI) downloadAll(ctx context.Context) {
	if u.selection == nil {
		u.showError(fmt.Errorf("请先加载课程"))
		return
	}
	selection := u.selection
	u.setBusy(true, "正在下载全部内容...")
	u.runTask(ctx, downloadTaskState{kind: taskDownloadAll, selection: selection, articleIndex: -1})
}

func (u *UI) downloadSelectedArticle(ctx context.Context) {
	if u.selection == nil {
		u.showError(fmt.Errorf("请先加载课程"))
		return
	}
	selectedTitle := u.articleSelect.Selected
	if selectedTitle == "" {
		u.showError(fmt.Errorf("请选择文章"))
		return
	}
	index := -1
	for i, article := range u.selection.Course.Articles {
		if article.Title == selectedTitle {
			index = i
			break
		}
	}
	if index < 0 {
		u.showError(fmt.Errorf("请选择有效的文章"))
		return
	}

	selection := u.selection
	u.setBusy(true, "正在下载选中文章...")
	u.runTask(ctx, downloadTaskState{kind: taskDownloadOne, selection: selection, articleIndex: index})
}

func (u *UI) newService(ctx context.Context) (*app.Service, error) {
	cfg, err := u.buildConfig()
	if err != nil {
		return nil, err
	}
	return app.NewService(ctx, &cfg, u.handleDownloadProgress)
}

func (u *UI) buildConfig() (config.AppConfig, error) {
	waitSeconds, err := strconv.Atoi(strings.TrimSpace(u.printPDFWaitEntry.Text))
	if err != nil {
		return config.AppConfig{}, fmt.Errorf("PDF等待秒数格式不合法")
	}
	timeoutSeconds, err := strconv.Atoi(strings.TrimSpace(u.printPDFTimeoutEntry.Text))
	if err != nil {
		return config.AppConfig{}, fmt.Errorf("PDF超时秒数格式不合法")
	}
	interval, err := strconv.Atoi(strings.TrimSpace(u.intervalEntry.Text))
	if err != nil {
		return config.AppConfig{}, fmt.Errorf("下载间隔秒数格式不合法")
	}
	comments, err := strconv.Atoi(u.commentsSelect.Selected)
	if err != nil {
		return config.AppConfig{}, fmt.Errorf("评论模式格式不合法")
	}

	outputType := 0
	if u.outputPDFCheck.Checked {
		outputType |= 1
	}
	if u.outputMarkdownCheck.Checked {
		outputType |= 2
	}
	if u.outputAudioCheck.Checked {
		outputType |= 4
	}

	return config.AppConfig{
		Gcid:                   strings.TrimSpace(u.gcidEntry.Text),
		Gcess:                  strings.TrimSpace(u.gcessEntry.Text),
		DownloadFolder:         strings.TrimSpace(u.downloadFolderEntry.Text),
		Quality:                u.qualitySelect.Selected,
		DownloadComments:       comments,
		ColumnOutputType:       outputType,
		PrintPDFWaitSeconds:    waitSeconds,
		PrintPDFTimeoutSeconds: timeoutSeconds,
		Interval:               interval,
		IsEnterprise:           u.enterpriseCheck.Checked,
		LogLevel:               u.logLevelSelect.Selected,
	}, nil
}

func (u *UI) setBusy(busy bool, status string) {
	if status == "" {
		status = "等待操作"
	}
	u.statusLabel.SetText(status)
	if busy {
		u.progressBar.Show()
		u.progressBar.SetValue(0)
		u.progressDetailLabel.SetText("准备开始...")
		u.loadButton.Disable()
		u.downloadAllButton.Disable()
		u.downloadOneButton.Disable()
		u.pauseButton.Enable()
		u.pauseButton.SetText("暂停下载")
		u.cancelButton.Enable()
		u.saveConfigButton.Disable()
		u.refreshLogButton.Disable()
		return
	}

	u.progressBar.Hide()
	u.progressBar.SetValue(0)
	u.progressDetailLabel.SetText("当前无下载任务")
	u.loadButton.Enable()
	u.pauseButton.Disable()
	u.pauseButton.SetText("暂停下载")
	u.cancelButton.Disable()
	u.saveConfigButton.Enable()
	u.refreshLogButton.Enable()
	if u.selection != nil {
		u.downloadAllButton.Enable()
		if !u.selection.IsDirectMode && len(u.selection.Course.Articles) > 0 {
			u.downloadOneButton.Enable()
		}
	}
}

func (u *UI) handleDownloadProgress(download progress.Download) {
	fyne.Do(func() {
		u.progressBar.Show()
		if download.TotalBytes > 0 {
			value := float64(download.DownloadedBytes) / float64(download.TotalBytes)
			if value < 0 {
				value = 0
			}
			if value > 1 {
				value = 1
			}
			u.progressBar.SetValue(value)
			u.progressDetailLabel.SetText(fmt.Sprintf(
				"当前课程进度：%d/%d | 当前条目：%s | 视频下载 %.1f%%",
				download.CurrentItem,
				max(download.TotalItems, 1),
				download.ItemTitle,
				value*100,
			))
			return
		}

		u.progressBar.SetValue(0)
		u.progressDetailLabel.SetText(fmt.Sprintf(
			"当前课程进度：%d/%d | 当前条目：%s | 状态：%s",
			download.CurrentItem,
			max(download.TotalItems, 1),
			download.ItemTitle,
			download.Stage,
		))
	})
}

func (u *UI) finishTask(err error, fromResume bool) bool {
	u.taskMu.Lock()
	defer u.taskMu.Unlock()
	if u.activeTask == nil {
		return false
	}
	if err == context.Canceled {
		if u.activeTask.paused {
			u.activeTask.running = false
			u.activeCancel = nil
			u.progressBar.Hide()
			u.pauseButton.SetText("继续下载")
			u.cancelButton.Enable()
			u.loadButton.Enable()
			u.saveConfigButton.Enable()
			u.refreshLogButton.Enable()
			u.progressDetailLabel.SetText("下载已暂停，可继续")
			u.statusLabel.SetText("下载已暂停")
			return true
		}
		u.activeTask = nil
		u.activeCancel = nil
		u.setBusy(false, "")
		u.statusLabel.SetText("下载已取消")
		return true
	}
	u.activeTask = nil
	u.activeCancel = nil
	return false
}

func (u *UI) togglePauseDownload(baseCtx context.Context) {
	u.taskMu.Lock()
	task := u.activeTask
	cancel := u.activeCancel
	if task == nil {
		u.taskMu.Unlock()
		return
	}
	if task.running {
		task.paused = true
		u.taskMu.Unlock()
		if cancel != nil {
			cancel()
		}
		return
	}
	if !task.paused {
		u.taskMu.Unlock()
		return
	}
	resumeTask := *task
	resumeTask.paused = false
	resumeTask.running = true
	u.taskMu.Unlock()

	u.pauseButton.SetText("暂停下载")
	u.cancelButton.Enable()
	u.setBusy(true, "正在继续下载...")
	u.runTask(baseCtx, resumeTask)
}

func (u *UI) cancelDownload() {
	u.taskMu.Lock()
	cancel := u.activeCancel
	u.taskMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (u *UI) runTask(baseCtx context.Context, task downloadTaskState) {
	runCtx, cancel := context.WithCancel(baseCtx)
	u.taskMu.Lock()
	u.activeCancel = cancel
	u.activeTask = &task
	u.taskMu.Unlock()

	svc, err := u.newService(runCtx)
	if err != nil {
		u.taskMu.Lock()
		u.activeCancel = nil
		u.activeTask = nil
		u.taskMu.Unlock()
		u.setBusy(false, "")
		u.showError(err)
		return
	}

	go func() {
		var runErr error
		switch task.kind {
		case taskDownloadAll:
			runErr = svc.DownloadAll(task.selection)
		case taskDownloadOne:
			runErr = svc.DownloadArticle(task.selection, task.articleIndex)
		default:
			return
		}
		fyne.Do(func() {
			u.refreshLogView()
			if u.finishTask(runErr, true) {
				return
			}
			u.setBusy(false, "")
			if runErr != nil {
				u.showError(runErr)
				return
			}
			if task.kind == taskDownloadAll {
				u.statusLabel.SetText("下载完成")
			} else {
				u.statusLabel.SetText("文章下载完成")
			}
		})
	}()
}

func (u *UI) showError(err error) {
	u.statusLabel.SetText(err.Error())
	u.refreshLogView()
	dialog.ShowError(err, u.window)
}

func (u *UI) refreshLogView() {
	if !u.logTabActive {
		return
	}
	logText := readLogFile()
	if logText == u.lastLogText {
		return
	}
	u.lastLogText = logText
	u.logLines = strings.Split(logText, "\n")
	u.logList.Refresh()
	if len(u.logLines) > 0 {
		u.logList.ScrollToBottom()
	}
}

func (u *UI) confirmDeleteLog() {
	dialog.ShowConfirm("删除日志", "确认删除当前日志文件？", func(confirm bool) {
		if !confirm {
			return
		}
		if err := deleteLogFile(); err != nil {
			u.showError(err)
			return
		}
		u.lastLogText = ""
		u.logLines = []string{"当前还没有日志文件"}
		u.logList.Refresh()
		u.logList.ScrollToTop()
		u.statusLabel.SetText("日志已删除")
	}, u.window)
}

func (u *UI) startLogAutoRefresh(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if !u.logTabActive {
					continue
				}
				fyne.Do(func() {
					u.refreshLogView()
				})
			}
		}
	}()
}

func (u *UI) persistSettings() error {
	outputPDF := u.outputPDFCheck.Checked
	outputMarkdown := u.outputMarkdownCheck.Checked
	outputAudio := u.outputAudioCheck.Checked
	settings := persistedSettings{
		Gcid:                   strings.TrimSpace(u.gcidEntry.Text),
		Gcess:                  strings.TrimSpace(u.gcessEntry.Text),
		DownloadFolder:         strings.TrimSpace(u.downloadFolderEntry.Text),
		Quality:                u.qualitySelect.Selected,
		DownloadComments:       u.commentsSelect.Selected,
		PrintPDFWaitSeconds:    strings.TrimSpace(u.printPDFWaitEntry.Text),
		PrintPDFTimeoutSeconds: strings.TrimSpace(u.printPDFTimeoutEntry.Text),
		Interval:               strings.TrimSpace(u.intervalEntry.Text),
		IsEnterprise:           u.enterpriseCheck.Checked,
		LogLevel:               u.logLevelSelect.Selected,
		OutputPDF:              &outputPDF,
		OutputMarkdown:         &outputMarkdown,
		OutputAudio:            &outputAudio,
		LastProductID:          strings.TrimSpace(u.productIDEntry.Text),
		LastProductType:        u.productTypeSelect.Selected,
	}
	return saveSettings(settings)
}

func (u *UI) applySettings(settings *persistedSettings) {
	if settings == nil {
		return
	}
	if settings.DownloadFolder != "" {
		u.cfg.DownloadFolder = settings.DownloadFolder
	}
	if settings.Quality != "" {
		u.cfg.Quality = settings.Quality
	}
	if settings.DownloadComments != "" {
		u.cfg.DownloadComments = settings.DownloadComments
	}
	if settings.PrintPDFWaitSeconds != "" {
		u.cfg.PrintPDFWaitSeconds = settings.PrintPDFWaitSeconds
	}
	if settings.PrintPDFTimeoutSeconds != "" {
		u.cfg.PrintPDFTimeoutSeconds = settings.PrintPDFTimeoutSeconds
	}
	if settings.Interval != "" {
		u.cfg.Interval = settings.Interval
	}
	if settings.LogLevel != "" {
		u.cfg.LogLevel = settings.LogLevel
	}
	u.cfg.Gcid = settings.Gcid
	u.cfg.Gcess = settings.Gcess
	u.cfg.IsEnterprise = settings.IsEnterprise
	if settings.OutputPDF != nil {
		u.cfg.OutputPDF = *settings.OutputPDF
	}
	if settings.OutputMarkdown != nil {
		u.cfg.OutputMarkdown = *settings.OutputMarkdown
	}
	if settings.OutputAudio != nil {
		u.cfg.OutputAudio = *settings.OutputAudio
	}
	u.cfg.ProductID = settings.LastProductID
	u.cfg.ProductType = settings.LastProductType
}
