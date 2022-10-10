package main

import (
	"encoding/binary"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"golang.org/x/sys/windows"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"unsafe"
)

var (
	err error
	modelName string
	handle windows.Handle
	baseAddr windows.Handle
)

func init()  {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main(){
	// 获取pid
	modelName = "PlantsVsZombies.exe"
	fmt.Printf("程序名称:%s\n", modelName)
	pid := getProcPid(modelName)
	fmt.Println("pid:", pid)

	// 获取进程句柄
	handle,err = windows.OpenProcess(0x1F0FFF, false, uint32(pid))
	if err != nil {
		log.Fatal(err)
	}
	defer windows.CloseHandle(handle)
	fmt.Println("句柄:",handle)

	// 获取模块基址
	baseAddr = getBaseAddr(handle, modelName)
	fmt.Printf("基址:0x%X\n", baseAddr)

	//// 修改阳光值
	//sunshineOffset := []int64{0x355E0C,0x868,0x5578}
	//writeMemory(handle, baseAddr, sunshineOffset, uint32(9999))
	//
	//// 修改金币值
	//moneyOffset := []int64{0x355E0C,0x950,0x50}
	//writeMemory(handle, baseAddr, moneyOffset, uint32(56780))

	createApp()
}

func createApp()  {
	myApp := app.NewWithID("hello,world!")				// 创建APP
	myWindow := myApp.NewWindow("植物大战僵尸辅助")			// 创建窗口

	//myApp.SetIcon(theme.FyneLogo())
	myWindow.Resize(fyne.NewSize(250,150))			// 设置窗口大小
	myWindow.CenterOnScreen()								// 窗口居中显示
	myWindow.SetMaster()									// 设置为主窗口

	sunshineLabel := widget.NewLabel("sunshine")
	sunshineEntry := widget.NewEntry()
	sunshineButton := widget.NewButton("modify", func() {
		data, err := strconv.Atoi(sunshineEntry.Text)
		if err != nil {
			return
		}
		sunshineOffset := []int64{0x355E0C,0x868,0x5578}
		writeMemory(handle, baseAddr, sunshineOffset, uint32(data))
	})
	c1 := container.NewGridWithColumns(3, sunshineLabel, sunshineEntry, sunshineButton)

	moneyLabel := widget.NewLabel("money")
	moneyEntry := widget.NewEntry()
	moneyButton := widget.NewButton("modify", func() {
		data, err := strconv.Atoi(moneyEntry.Text)
		if err != nil {
			return
		}
		moneyOffset := []int64{0x355E0C,0x950,0x50}
		writeMemory(handle, baseAddr, moneyOffset, uint32(data))
	})
	c2 := container.NewGridWithColumns(3, moneyLabel, moneyEntry, moneyButton)

	c := container.NewVBox(c1, c2)
	myWindow.SetContent(c)			// 创建导航
	myWindow.ShowAndRun()			// 事件循环
}

func writeMemory(hd windows.Handle, baseAddr windows.Handle, offset []int64, newValue uint32) (uintptr, uint32){
	// 读取实际地址和值
	addr,value := readMemory(hd, baseAddr, offset)

	// 写入新值
	sli := make([]byte, 4)
	binary.LittleEndian.PutUint32(sli, newValue)
	err = windows.WriteProcessMemory(hd, addr, &sli[0],4, nil)
	if err != nil {
		log.Println(err)
	}
	return addr,value
}

func readMemory(hd windows.Handle, baseAddr windows.Handle, offset[]int64) (uintptr, uint32) {
	var addr uintptr
	var value uint32
	tmpAddr := uintptr(baseAddr)
	for _,v := range offset {
		//func ReadProcessMemory(process Handle, baseAddress uintptr, buffer *byte, size uintptr, numberOfBytesRead *uintptr) (err error) {
		var readSize uintptr
		data := make([]byte, 4)
		addr = tmpAddr + uintptr(v)
		err = windows.ReadProcessMemory(hd, addr, &data[0], 4, &readSize)
		if err != nil {
			log.Println(err)
		}
		value = binary.LittleEndian.Uint32(data)
		//fmt.Printf("基址:0x%X+偏移:0x%X=地址:0x%X	16进制值:0x%X 10进制值:%d\n", tmpAddr, v, addr, value, value)
		tmpAddr = uintptr(value)
	}
	return addr, value
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
func getBaseAddr(hd windows.Handle, modelName string) windows.Handle {
	var base windows.Handle
	var module [10000]windows.Handle
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
	return base
}