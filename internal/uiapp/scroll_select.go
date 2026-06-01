package uiapp

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const defaultScrollSelectPlaceholder = "(Select one)"

type ScrollableSelect struct {
	widget.DisableableWidget

	Alignment   fyne.TextAlign
	Selected    string
	Options     []string
	PlaceHolder string
	OnChanged   func(string)

	focused bool
	hovered bool
	popUp   *widget.PopUp
}

func NewScrollableSelect(options []string, changed func(string)) *ScrollableSelect {
	s := &ScrollableSelect{
		Options:     options,
		OnChanged:   changed,
		PlaceHolder: defaultScrollSelectPlaceholder,
	}
	s.ExtendBaseWidget(s)
	return s
}

func (s *ScrollableSelect) ClearSelected() {
	s.updateSelected("")
}

func (s *ScrollableSelect) CreateRenderer() fyne.WidgetRenderer {
	s.ExtendBaseWidget(s)

	bg := canvas.NewRectangle(color.Transparent)
	label := widget.NewLabel("")
	label.Truncation = fyne.TextTruncateEllipsis
	icon := widget.NewIcon(theme.Icon(theme.IconNameArrowDropDown))

	r := &scrollableSelectRenderer{
		selectWidget: s,
		background:   bg,
		label:        label,
		icon:         icon,
		objects:      []fyne.CanvasObject{bg, label, icon},
	}
	r.Refresh()
	return r
}

func (s *ScrollableSelect) FocusGained() {
	s.focused = true
	s.Refresh()
}

func (s *ScrollableSelect) FocusLost() {
	s.focused = false
	s.Refresh()
}

func (s *ScrollableSelect) Hide() {
	if s.popUp != nil {
		s.popUp.Hide()
		s.popUp = nil
	}
	s.BaseWidget.Hide()
}

func (s *ScrollableSelect) MouseIn(*desktop.MouseEvent) {
	s.hovered = true
	s.Refresh()
}

func (s *ScrollableSelect) MouseMoved(*desktop.MouseEvent) {}

func (s *ScrollableSelect) MouseOut() {
	s.hovered = false
	s.Refresh()
}

func (s *ScrollableSelect) Move(pos fyne.Position) {
	s.BaseWidget.Move(pos)
	if s.popUp != nil {
		s.popUp.Move(s.popUpPos())
	}
}

func (s *ScrollableSelect) Resize(size fyne.Size) {
	s.BaseWidget.Resize(size)
	if s.popUp != nil {
		s.popUp.Resize(s.popUpSize())
	}
}

func (s *ScrollableSelect) SetOptions(options []string) {
	s.Options = options
	keepSelected := false
	for _, option := range options {
		if option == s.Selected {
			keepSelected = true
			break
		}
	}
	if !keepSelected {
		s.Selected = ""
	}
	s.Refresh()
}

func (s *ScrollableSelect) SetSelected(text string) {
	for _, option := range s.Options {
		if text == option {
			s.updateSelected(text)
			return
		}
	}
}

func (s *ScrollableSelect) Tapped(*fyne.PointEvent) {
	if s.Disabled() {
		return
	}
	s.showPopUp()
}

func (s *ScrollableSelect) TypedKey(event *fyne.KeyEvent) {
	switch event.Name {
	case fyne.KeySpace, fyne.KeyUp, fyne.KeyDown:
		s.showPopUp()
	case fyne.KeyRight:
		i := s.SelectedIndex() + 1
		if i >= len(s.Options) {
			i = 0
		}
		s.SetSelectedIndex(i)
	case fyne.KeyLeft:
		i := s.SelectedIndex() - 1
		if i < 0 {
			i = len(s.Options) - 1
		}
		s.SetSelectedIndex(i)
	}
}

func (s *ScrollableSelect) TypedRune(rune) {}

func (s *ScrollableSelect) SelectedIndex() int {
	for i, option := range s.Options {
		if option == s.Selected {
			return i
		}
	}
	return -1
}

func (s *ScrollableSelect) SetSelectedIndex(index int) {
	if index < 0 || index >= len(s.Options) {
		return
	}
	s.updateSelected(s.Options[index])
}

func (s *ScrollableSelect) popUpPos() fyne.Position {
	return fyne.CurrentApp().Driver().AbsolutePositionForObject(s).Add(
		fyne.NewPos(0, s.Size().Height-s.Theme().Size(theme.SizeNameInputBorder)),
	)
}

func (s *ScrollableSelect) popUpSize() fyne.Size {
	c := fyne.CurrentApp().Driver().CanvasForObject(s)
	if c == nil {
		return fyne.NewSize(s.Size().Width, s.rowHeight()*4)
	}

	rowCount := len(s.Options)
	if rowCount < 1 {
		rowCount = 1
	}
	desiredRows := rowCount
	if desiredRows > 8 {
		desiredRows = 8
	}

	padding := s.Theme().Size(theme.SizeNamePadding) * 2
	desiredHeight := s.rowHeight()*float32(desiredRows) + padding
	maxHeight := c.Size().Height - s.Theme().Size(theme.SizeNamePadding)*4
	if maxHeight < s.rowHeight()*2 {
		maxHeight = s.rowHeight() * 2
	}
	if desiredHeight > maxHeight {
		desiredHeight = maxHeight
	}
	return fyne.NewSize(s.Size().Width, desiredHeight)
}

func (s *ScrollableSelect) rowHeight() float32 {
	return widget.NewLabel("Ag").MinSize().Height + s.Theme().Size(theme.SizeNameInnerPadding)
}

func (s *ScrollableSelect) showPopUp() {
	if s.popUp != nil {
		s.popUp.Hide()
		s.popUp = nil
		return
	}

	list := widget.NewList(
		func() int {
			return len(s.Options)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			obj.(*widget.Label).SetText(s.Options[id])
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || id >= len(s.Options) {
			return
		}
		s.updateSelected(s.Options[id])
		if s.popUp != nil {
			s.popUp.Hide()
			s.popUp = nil
		}
	}

	if selectedIndex := s.SelectedIndex(); selectedIndex >= 0 {
		list.Select(selectedIndex)
		list.ScrollTo(selectedIndex)
	}

	content := container.NewStack(list)
	pop := widget.NewPopUp(content, fyne.CurrentApp().Driver().CanvasForObject(s))
	pop.ShowAtPosition(s.popUpPos())
	pop.Resize(s.popUpSize())
	s.popUp = pop
}

func (s *ScrollableSelect) updateSelected(text string) {
	s.Selected = text
	if s.OnChanged != nil {
		s.OnChanged(text)
	}
	s.Refresh()
}

type scrollableSelectRenderer struct {
	selectWidget *ScrollableSelect
	background   *canvas.Rectangle
	label        *widget.Label
	icon         *widget.Icon
	objects      []fyne.CanvasObject
}

func (r *scrollableSelectRenderer) Destroy() {}

func (r *scrollableSelectRenderer) Layout(size fyne.Size) {
	th := r.selectWidget.Theme()
	pad := th.Size(theme.SizeNamePadding)
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)

	r.background.Resize(size)

	iconX := size.Width - iconSize - innerPad
	if iconX < pad {
		iconX = pad
	}
	r.icon.Resize(fyne.NewSquareSize(iconSize))
	r.icon.Move(fyne.NewPos(iconX, (size.Height-iconSize)/2))

	labelWidth := iconX - pad
	if labelWidth < 0 {
		labelWidth = 0
	}
	r.label.Resize(fyne.NewSize(labelWidth, r.label.MinSize().Height))
	r.label.Move(fyne.NewPos(pad, (size.Height-r.label.MinSize().Height)/2))
}

func (r *scrollableSelectRenderer) MinSize() fyne.Size {
	th := r.selectWidget.Theme()
	innerPad := th.Size(theme.SizeNameInnerPadding)
	iconSize := th.Size(theme.SizeNameInlineIcon)
	labelSize := widget.NewLabel(r.currentText()).MinSize()
	return fyne.NewSize(labelSize.Width+iconSize+innerPad*4, labelSize.Height+innerPad*2)
}

func (r *scrollableSelectRenderer) Objects() []fyne.CanvasObject {
	return r.objects
}

func (r *scrollableSelectRenderer) Refresh() {
	th := r.selectWidget.Theme()
	v := fyne.CurrentApp().Settings().ThemeVariant()

	r.label.Alignment = r.selectWidget.Alignment
	r.label.SetText(r.currentText())

	if r.selectWidget.Disabled() {
		r.background.FillColor = th.Color(theme.ColorNameDisabledButton, v)
		r.label.TextStyle = fyne.TextStyle{}
		r.label.Importance = widget.LowImportance
		r.icon.SetResource(theme.NewDisabledResource(theme.Icon(theme.IconNameArrowDropDown)))
	} else {
		switch {
		case r.selectWidget.focused:
			r.background.FillColor = th.Color(theme.ColorNameFocus, v)
		case r.selectWidget.hovered:
			r.background.FillColor = th.Color(theme.ColorNameHover, v)
		default:
			r.background.FillColor = th.Color(theme.ColorNameInputBackground, v)
		}
		r.label.Importance = widget.MediumImportance
		r.icon.SetResource(theme.Icon(theme.IconNameArrowDropDown))
	}
	r.background.CornerRadius = th.Size(theme.SizeNameInputRadius)
	r.Layout(r.selectWidget.Size())
	r.background.Refresh()
	r.label.Refresh()
	r.icon.Refresh()
}

func (r *scrollableSelectRenderer) currentText() string {
	if r.selectWidget.Selected != "" {
		return r.selectWidget.Selected
	}
	if r.selectWidget.PlaceHolder != "" {
		return r.selectWidget.PlaceHolder
	}
	return defaultScrollSelectPlaceholder
}

