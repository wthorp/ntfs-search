package main

import (
	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var tv *walk.TableView
	var q *walk.LineEdit

	fileModel := NewFileInfoModel()

	d.MainWindow{
		AssignTo: &mw,
		Title:    "NTFS Search",
		Size:     d.Size{Width: 800, Height: 600},
		Layout:   d.VBox{Margins: d.Margins{Left: 2, Top: 2, Right: 2, Bottom: 2}},
		Children: []d.Widget{
			d.Composite{
				Layout: d.HBox{MarginsZero: true},
				Children: []d.Widget{
					d.LineEdit{
						AssignTo: &q,
						OnTextChanged: func() {
							fileModel.SetQuery(q.Text())
						},
					},
					d.PushButton{
						Text: "Settings",
					},
				},
			},
			d.TableView{
				AssignTo:         &tv,
				ColumnsOrderable: true,
				MultiSelection:   true,
				Columns: []d.TableViewColumn{
					{DataMember: "Name", Width: 192},
					{DataMember: "Path", Width: 192},
					{DataMember: "Size", Format: "%d", Alignment: d.AlignFar, Width: 64},
					{DataMember: "Modified", Format: "2006-01-02 15:04:05", Width: 120},
				},
				Model: fileModel,
			},
		},
	}.Run()
}
