package main

import (
	"encoding/binary"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sys/windows"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"unsafe"
)

var (
	modelName = "PlantsVsZombies.exe"
	pid int
	handle windows.Handle
	baseAddr uintptr
	openNOCD bool = false			// 无冷却开关
	err error
)

func init()  {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	pid = getProcPid(modelName)
	handle,err = windows.OpenProcess(0x1F0FFF, false, uint32(pid))
	if err != nil {
		log.Fatal(err)
	}
	baseAddr = getBaseAddr(handle, modelName)
}

func main(){
	fmt.Printf("程序名称:%s\n", modelName)
	fmt.Println("pid:", pid)
	fmt.Println("句柄:",handle)
	fmt.Printf("基址:0x%X\n", baseAddr)

	go modifyCD(handle, baseAddr)
	createApp()
}

// 获取THREADSTACK0基址
func getTHREADSTACK0Addr(pid string) uintptr{
	task := exec.Command("threadstack.exe", pid)
	output, err := task.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	addr := strings.Split(strings.Split(string(output), "ADDRESS")[1], "\n")[0][2:]
	addr = strings.ReplaceAll(addr, "\r", "")
	fmt.Printf("THREADSTACK0基址:%v\n", addr)

	r,err := strconv.ParseUint(addr, 0, 0)
	if err != nil {
		log.Fatal(err)
	}
	return uintptr(r)
}

// 无冷却监测
func modifyCD(hd windows.Handle, baseAddr uintptr){
	// 通过获取THREADSTACK0，重新计算基址
	THREADSTACK0 := getTHREADSTACK0Addr(strconv.Itoa(pid))
	value := readUint32(hd, THREADSTACK0-uintptr(0x00000204))
	baseAddr = uintptr(value)
	cdOffset := []int64{0x0,0x8,0x15C,0x4C}
	for{
		if openNOCD {
			addr := readDynamicAddr(hd, baseAddr, cdOffset)
			for i:=0; i<10; i++ {											// 10个植物卡片
				writeUint32(hd, addr+uintptr(80*i), 10000)			// 向日葵冷却周期为0~750，0代表冷却完成；取10000代表冷却上限
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// 修改阳光
func modifySunshine(hd windows.Handle, baseAddr uintptr, newValue uint32){
	sunshineOffset := []int64{0x355E0C,0x320, 0x18, 0x0, 0x8, 0x5578}
	addr := readDynamicAddr(hd, baseAddr, sunshineOffset)
	writeUint32(hd, addr, newValue)
}

// 修改金币
func modifyMoney(hd windows.Handle, baseAddr uintptr, newValue uint32){
	moneyOffset := []int64{0x355E0C,0x950,0x50}
	addr := readDynamicAddr(hd, baseAddr, moneyOffset)
	writeUint32(hd, addr, newValue)
}

// 根据进程句柄和地址，写入一个4字节整形
func writeUint32(hd windows.Handle, addr uintptr, value uint32) {
	sli := make([]byte, 4)
	binary.LittleEndian.PutUint32(sli, value)
	err = windows.WriteProcessMemory(hd, addr, &sli[0],4, nil)
	if err != nil {
		log.Println("写入内存失败！")
	}
}

// 根据进程句柄和地址，读取4字节整形，可解释为整数值或指针值
func readUint32(hd windows.Handle, addr uintptr) uint32 {
	data := make([]byte, 4)
	err = windows.ReadProcessMemory(hd, addr, &data[0], 4, nil)
	if err != nil {
		log.Println("读取内存失败！")
	}
	return binary.LittleEndian.Uint32(data)
}

// 根据基址和偏移读取实际的动态地址
func readDynamicAddr(hd windows.Handle, baseAddr uintptr, offset[]int64) uintptr {
	var addr uintptr
	var value uint32
	tmpAddr := baseAddr
	for _,v := range offset {
		addr = tmpAddr + uintptr(v)
		value = readUint32(hd, addr)
		tmpAddr = uintptr(value)
	}
	return addr
}

// 根据程序名获取进程ID
func getProcPid(PROCESS string) int {
	task := exec.Command("cmd","/c","wmic", "process", "get", "name,","ProcessId","|","findstr",PROCESS)
	output, err := task.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	sli := strings.Split(strings.TrimSpace(string(output)), " ")
	pid,_ := strconv.Atoi(sli[len(sli) - 1])
	return pid
}

// 根据进程名和进程ID获取基址
func getBaseAddr(hd windows.Handle, modelName string) uintptr {
	var base windows.Handle
	var module [100]windows.Handle
	var cbNeeded uint32

	err = windows.EnumProcessModules(hd, &module[0], uint32(unsafe.Sizeof(module)), &cbNeeded)
	if err != nil {
		log.Fatal(err)
	}
	moduleNum := cbNeeded/uint32(unsafe.Sizeof(module[0]))
	//fmt.Println(module, cbNeeded, err, moduleNum)
	for i:=0; i<int(moduleNum); i++ {
		baseName := make([]uint16, 50)
		err = windows.GetModuleBaseName(hd, module[i], &baseName[0], 50)
		s := windows.UTF16ToString(baseName)
		//fmt.Printf("0x%X %s\n", module[i], s)
		if s == modelName {
			base =  module[i]
		}
	}
	return uintptr(base)
}

// 图形界面
func createApp()  {
	myApp := app.NewWithID("hello,world!")				// 创建APP
	myWindow := myApp.NewWindow("植物大战僵尸辅助")			// 创建窗口

	//myApp.SetIcon(theme.FyneLogo())
	myWindow.Resize(fyne.NewSize(250,150))			// 设置窗口大小
	myWindow.CenterOnScreen()								// 窗口居中显示
	myWindow.SetMaster()									// 设置为主窗口

	// 阳光
	sunshineLabel := widget.NewLabel("Sunshine")
	sunshineEntry := widget.NewEntry()
	sunshineButton := widget.NewButton("modify", func() {
		data, err := strconv.Atoi(sunshineEntry.Text)
		if err != nil {
			return
		}
		modifySunshine(handle, baseAddr, uint32(data))
	})
	c1 := container.NewGridWithColumns(3, sunshineLabel, sunshineEntry, sunshineButton)

	// 金币
	moneyLabel := widget.NewLabel("Money")
	moneyEntry := widget.NewEntry()
	moneyButton := widget.NewButton("modify", func() {
		data, err := strconv.Atoi(moneyEntry.Text)
		if err != nil {
			return
		}
		modifyMoney(handle, baseAddr, uint32(data))
	})
	c2 := container.NewGridWithColumns(3, moneyLabel, moneyEntry, moneyButton)

	// 冷却
	CDLabel := widget.NewLabel("ZeroCD")
	CDCheck := widget.NewCheck("open", func(b bool) {
		if b {
			openNOCD = true
		}else {
			openNOCD = false
		}
	})
	c3 := container.NewGridWithColumns(3, CDLabel, layout.NewSpacer(), CDCheck)

	c := container.NewVBox(c1, c2, c3)
	myWindow.SetContent(c)			// 创建导航
	myWindow.ShowAndRun()			// 事件循环
}