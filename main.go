package main

import (
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/ld-2022/jsonx"
	"github.com/shirou/gopsutil/process"
	"path/filepath"
	"strings"
)

func main() {
	a := app.New()
	w := a.NewWindow("process list")
	dataList := jsonx.NewJSONArray()
	selectItem := jsonx.NewJSONArray()
	killBtn := widget.NewButton("Kill", func() {
		if selectItem.Size() > 0 {
			dialog.ShowConfirm("Kill", "Are you sure to kill the process?", func(b bool) {
				if b {
					mm := dataList.Get(0).(*jsonx.JSONObject)
					pidList := mm.GetJSONArray("pid").ToArray()
					killFlag := true
					for _, pid := range pidList {
						p, _ := process.NewProcess(pid.(int32))
						killErr := p.Kill()
						if killErr != nil {
							killFlag = false
							return
						}
					}
					if killFlag {
						fmt.Println("kill success")
						dataList.Remove(mm)
					} else {
						fmt.Println("kill failed")
						dialog.ShowError(fmt.Errorf("kill failed:"+mm.GetString("name")), w)
					}
				}
			}, w)
		} else {
			fmt.Println("selectID is -1")
			dialog.ShowError(fmt.Errorf("no select"), w)
		}
	})
	appList := AppList(dataList, selectItem)
	search := widget.NewEntry()
	search.OnChanged = func(s string) {
		GetAppList(dataList, search, appList)
	}
	go GetAppList(dataList, search, appList)
	content := container.NewBorder(search, killBtn, nil, nil, appList)
	w.SetContent(content)
	w.Resize(fyne.NewSize(350, 600))
	w.ShowAndRun()

}

func GetAppList(dataList *jsonx.JSONArray, search *widget.Entry, appList fyne.CanvasObject) {
	dataList.Clear()
	appList.Refresh()
	searchList, _ := GetProcessList(search.Text)
	dataList.AddAll(searchList.ToArray())
	appList.Refresh()
}

func AppList(data *jsonx.JSONArray, selectID *jsonx.JSONArray) fyne.CanvasObject {
	icon := widget.NewIcon(nil)
	label := widget.NewLabel("Select An Item From The List")

	list := widget.NewList(
		func() int {
			return data.Size()
		},
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.DocumentIcon()), widget.NewLabel("Template Object"))
		},
		func(id widget.ListItemID, item fyne.CanvasObject) {
			d := data.Get(id).(*jsonx.JSONObject)
			item.(*fyne.Container).Objects[1].(*widget.Label).SetText(d.GetString("name"))
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		d := data.Get(id).(*jsonx.JSONObject)
		selectID.Add(d)
		label.SetText(d.GetString("name"))
		icon.SetResource(theme.DocumentIcon())
	}
	list.OnUnselected = func(id widget.ListItemID) {
		selectID.Clear()
		label.SetText("Select An Item From The List")
		icon.SetResource(nil)
	}

	return list
}

func GetProcessList(searchName string) (array *jsonx.JSONArray, err error) {
	processes, err := process.Processes()
	if err != nil {
		return
	}
	var name string
	j := jsonx.NewJSONObject()
	array = jsonx.NewJSONArray()
	for _, p := range processes {
		name, err = p.Name()
		if err != nil {
			return
		}
		p = GetParent(p, name)
		name, err = p.Name()
		exe, exeErr := p.Exe()
		if exeErr != nil {
			continue
		}
		if strings.Contains(name, "/") {
			name = filepath.Base(exe)
		}
		if searchName != "" && !strings.Contains(name, searchName) {
			continue
		}
		if j.ContainsKey(name) {
			pidList := j.GetJSONArray(name)
			if pidList.Contains(p.Pid) == false {
				pidList.Add(p.Pid)
			}
		} else {
			pidList := jsonx.NewJSONArray()
			pidList.Add(p.Pid)
			j.Put(name, pidList)
		}
	}
	j.ForEach(func(key string, value interface{}) bool {
		array.Add(jsonx.NewJSONObject().FluentPut("name", key).FluentPut("pid", value))
		return true
	})
	return
}

func GetParent(p *process.Process, name string) *process.Process {
	parent, err := p.Parent()
	if err != nil {
		return p
	}
	parentName, err := parent.Name()
	if err != nil {
		return p
	}
	if parentName != name {
		return p
	}
	return GetParent(parent, name)
}
