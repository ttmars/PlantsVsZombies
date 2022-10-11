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
	err error
	modelName string
	handle windows.Handle
	baseAddr windows.Handle
	openNOCD bool = false			// 无冷却开关
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

	// 无冷却
	go modifyCD(handle, baseAddr)

	createApp()
}

// 无冷却监测
func modifyCD(hd windows.Handle, baseAddr windows.Handle){
	var newValue uint32 = 10000			// 冷却区间，豌豆射手为0~750，每种植物冷却上线不一致
	for{
		if openNOCD {
			cdOffset := []int64{0x0003DAD8,0x14,0x4,0x15C,0x4C}
			//cdOffset := []int64{0x0003DAD8,0x8,0x4,0x15C,0x4C}

			addr,_ := readMemory(hd, baseAddr, cdOffset)
			sli := make([]byte, 4)
			binary.LittleEndian.PutUint32(sli, newValue)
			n := 10						// n表示植物槽，前n个植物槽都无冷却
			for i:=0;i<n;i++{
				err = windows.WriteProcessMemory(hd, addr+uintptr(80*i), &sli[0],4, nil)
				if err != nil {
					log.Println(err)
				}
			}
		}
		time.Sleep(time.Millisecond * 500)
	}
}

// 修改阳光
func modifySunshine(hd windows.Handle, baseAddr windows.Handle, newValue uint32){
	//sunshineOffset := []int64{0x355E0C,0x868,0x5578}
	sunshineOffset := []int64{0x355E0C,0x320, 0x18, 0x0, 0x8, 0x5578}
	addr,_ := readMemory(hd, baseAddr, sunshineOffset)

	// 写入新值
	sli := make([]byte, 4)
	binary.LittleEndian.PutUint32(sli, newValue)
	err = windows.WriteProcessMemory(hd, addr, &sli[0],4, nil)
	if err != nil {
		log.Println(err)
	}
}

// 修改金币
func modifyMoney(hd windows.Handle, baseAddr windows.Handle, newValue uint32){
	moneyOffset := []int64{0x355E0C,0x950,0x50}
	addr,_ := readMemory(hd, baseAddr, moneyOffset)

	// 写入新值
	sli := make([]byte, 4)
	binary.LittleEndian.PutUint32(sli, newValue)
	err = windows.WriteProcessMemory(hd, addr, &sli[0],4, nil)
	if err != nil {
		log.Println(err)
	}
}

// 根据基址和偏移读取实际地址和内容
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
	return base
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