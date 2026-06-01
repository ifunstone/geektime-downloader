package uiapp

import (
	"context"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/nicoxiang/geektime-downloader/internal/app"
	"github.com/nicoxiang/geektime-downloader/internal/config"
)

type UI struct {
	fyneApp fyne.App
	window  fyne.Window

	cfg appConfigState

	productOptions []app.ProductTypeOption
	selectedType   *app.ProductTypeOption
	selection      *app.SelectionResult

	productTypeSelect *widget.Select
	articleSelect     *widget.Select
	statusLabel       *widget.Label
	courseLabel       *widget.Label
	progressBar       *widget.ProgressBarInfinite
	logEntry          *widget.Entry
	loadButton        *widget.Button
	downloadAllButton *widget.Button
	downloadOneButton *widget.Button
	refreshLogButton  *widget.Button
	saveConfigButton  *widget.Button

	gcidEntry            *widget.Entry
	gcessEntry           *widget.Entry
	downloadFolderEntry  *widget.Entry
	qualitySelect        *widget.Select
	commentsSelect       *widget.Select
	outputPDFCheck       *widget.Check
	outputMarkdownCheck  *widget.Check
	outputAudioCheck     *widget.Check
	printPDFWaitEntry    *widget.Entry
	printPDFTimeoutEntry *widget.Entry
	intervalEntry        *widget.Entry
	enterpriseCheck      *widget.Check
	logLevelSelect       *widget.Select
	productIDEntry       *widget.Entry
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
	u.progressBar = widget.NewProgressBarInfinite()
	u.progressBar.Hide()
	u.logEntry = widget.NewMultiLineEntry()
	u.logEntry.Disable()
	u.logEntry.Wrapping = fyne.TextWrapWord
	u.logEntry.SetMinRowsVisible(14)

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

	u.qualitySelect = widget.NewSelect([]string{"ld", "sd", "hd"}, func(value string) {
		u.cfg.Quality = value
	})
	u.qualitySelect.SetSelected(u.cfg.Quality)

	u.commentsSelect = widget.NewSelect([]string{"0", "1", "2"}, func(value string) {
		u.cfg.DownloadComments = value
	})
	u.commentsSelect.SetSelected(u.cfg.DownloadComments)

	u.logLevelSelect = widget.NewSelect([]string{"debug", "info", "warn", "error", "none"}, func(value string) {
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

	u.productTypeSelect = widget.NewSelect(nil, func(value string) {
		for i := range u.productOptions {
			if u.productOptions[i].Text == value {
				u.selectedType = &u.productOptions[i]
				break
			}
		}
	})

	u.articleSelect = widget.NewSelect(nil, nil)
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
	u.refreshLogButton = widget.NewButton("刷新日志", func() {
		u.refreshLogView()
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

	u.refreshProductOptions()
	u.restoreSelectedProductType()
	u.refreshLogView()

	downloadPage := container.NewPadded(u.buildDownloadPage())
	settingsPage := container.NewPadded(u.buildSettingsPage())
	logPage := container.NewPadded(u.buildLogPage())

	tabs := container.NewAppTabs(
		container.NewTabItem("下载操作", downloadPage),
		container.NewTabItem("参数配置", settingsPage),
		container.NewTabItem("日志查看", logPage),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	content := container.NewBorder(
		nil,
		container.NewVBox(widget.NewSeparator(), u.statusLabel),
		nil,
		nil,
		tabs,
	)

	u.window.SetContent(content)
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
		u.courseLabel,
		u.progressBar,
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
		container.NewHBox(u.refreshLogButton),
		nil,
		nil,
		u.logEntry,
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
	svc, err := u.newService(ctx)
	if err != nil {
		u.showError(err)
		return
	}
	selection := u.selection
	u.setBusy(true, "正在下载全部内容，执行期间界面不会展示细粒度进度")
	go func() {
		err := svc.DownloadAll(selection)
		fyne.Do(func() {
			u.setBusy(false, "")
			u.refreshLogView()
			if err != nil {
				u.showError(err)
				return
			}
			u.statusLabel.SetText("下载完成")
		})
	}()
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

	svc, err := u.newService(ctx)
	if err != nil {
		u.showError(err)
		return
	}
	selection := u.selection
	u.setBusy(true, "正在下载选中文章...")
	go func() {
		err := svc.DownloadArticle(selection, index)
		fyne.Do(func() {
			u.setBusy(false, "")
			u.refreshLogView()
			if err != nil {
				u.showError(err)
				return
			}
			u.statusLabel.SetText("文章下载完成")
		})
	}()
}

func (u *UI) newService(ctx context.Context) (*app.Service, error) {
	cfg, err := u.buildConfig()
	if err != nil {
		return nil, err
	}
	return app.NewService(ctx, &cfg)
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
		u.loadButton.Disable()
		u.downloadAllButton.Disable()
		u.downloadOneButton.Disable()
		u.saveConfigButton.Disable()
		u.refreshLogButton.Disable()
		return
	}

	u.progressBar.Hide()
	u.loadButton.Enable()
	u.saveConfigButton.Enable()
	u.refreshLogButton.Enable()
	if u.selection != nil {
		u.downloadAllButton.Enable()
		if !u.selection.IsDirectMode && len(u.selection.Course.Articles) > 0 {
			u.downloadOneButton.Enable()
		}
	}
}

func (u *UI) showError(err error) {
	u.statusLabel.SetText(err.Error())
	u.refreshLogView()
	dialog.ShowError(err, u.window)
}

func (u *UI) refreshLogView() {
	u.logEntry.SetText(readLogFile())
	u.logEntry.CursorRow = len(strings.Split(u.logEntry.Text, "\n"))
	u.logEntry.CursorColumn = 0
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
