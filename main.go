package main

import (
	"PlantsVsZombies/myTheme"
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
	"unsafe"
)

var (
	pid int
	handle windows.Handle
	baseAddr uintptr
	err error
)

func init()  {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
}

func main(){
	modelName := "PlantsVsZombies.exe"
	//modelName = "Tutorial-x86_64.exe"
	pid = getProcPid(modelName)
	handle,err = windows.OpenProcess(0x1F0FFF, false, uint32(pid))
	if err != nil {
		log.Fatal(err)
	}
	baseAddr = getBaseAddr(handle, modelName)
	fmt.Printf("程序名称:%s\n", modelName)
	fmt.Println("pid:", pid)
	fmt.Println("句柄:",handle)
	fmt.Printf("基址:0x%X\n", baseAddr)

	createApp()
}

// 锁定末日蘑CD，默认值3000,30秒
func lockDoomCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x90, 0x90, 0x90, 0x90, 0x90, 0x90}
	writeNBytes(hd, baseAddr+uintptr(0x293C9), data)
}

// 解除锁定末日蘑CD
func unlockDoomCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x83, 0x7E, 0x18, 0x00, 0x75, 0x05}
	writeNBytes(hd, baseAddr+uintptr(0x293C9), data)
}

// 锁定加农炮CD，默认值3000,30秒
func lockCannonCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0xC7, 0x47, 0x54, 0x00, 0x00, 0x00, 0x00}
	writeNBytes(hd, baseAddr+uintptr(0x73196), data)
}

// 解除锁定加农炮CD
func unlockCannonCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0xC7, 0x47, 0x54, 0xB8, 0x0B, 0x00, 0x00}
	writeNBytes(hd, baseAddr+uintptr(0x73196), data)
}

// 锁定磁力菇CD，默认值1500,15秒
func lockMagneticCD(hd windows.Handle, baseAddr uintptr)  {
	//data := []byte{0x90, 0x90, 0x90, 0x90, 0x90, 0x90, 0x90}
	data := []byte{0xC7, 0x46, 0x54, 0x00, 0x00, 0x00, 0x00}
	writeNBytes(hd, baseAddr+uintptr(0x6F9EA), data)
}

// 解除锁定磁力菇CD
func unlockMagneticCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0xC7, 0x46, 0x54, 0xDC, 0x05, 0x00, 0x00}
	writeNBytes(hd, baseAddr+uintptr(0x6F9EA), data)
}

// 锁定所有植物血量
func lockPlantsBlood(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x90, 0x90, 0x90, 0x90}
	writeNBytes(hd, baseAddr+uintptr(0x14BA6A), data)

	data1 := []byte{0x90, 0x90, 0x90, 0x90}				// 巨人僵尸对地刺王造成的伤害逻辑
	writeNBytes(hd, baseAddr+uintptr(0x6CF93), data1)
}

// 解除锁定所有植物血量
func unlockPlantsBlood(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x83, 0x46, 0x40, 0xFC}
	writeNBytes(hd, baseAddr+uintptr(0x14BA6A), data)

	data1 := []byte{0x83, 0x46, 0x40, 0xCE}				// 巨人僵尸对地刺王造成的伤害逻辑
	writeNBytes(hd, baseAddr+uintptr(0x6CF93), data1)
}

// 锁定阳光值
func lockSunshine(hd windows.Handle, baseAddr uintptr)  {
	modifySunshine(hd, baseAddr, uint32(9990))
	//data := []byte{0x01, 0xDE}
	data := []byte{0x90, 0x90}
	writeNBytes(hd, baseAddr+uintptr(0x27694), data)
}

// 解除锁定阳光值
func unlockSunshine(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x2B, 0xF3}
	writeNBytes(hd, baseAddr+uintptr(0x27694), data)
}

// 开启无CD
func openZeroCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0x89, 0x76, 0x24}
	writeNBytes(hd, baseAddr+uintptr(0x9CDF9), data)
}

// 关闭无CD
func closeZeroCD(hd windows.Handle, baseAddr uintptr)  {
	data := []byte{0xFF, 0x46, 0x24}
	writeNBytes(hd, baseAddr+uintptr(0x9CDF9), data)
}

// 获取THREADSTACK0基址
//func getTHREADSTACK0Addr(pid string) uintptr{
//	task := exec.Command("threadstack.exe", pid)
//	output, err := task.CombinedOutput()
//	if err != nil {
//		log.Fatal(err)
//	}
//	addr := strings.Split(strings.Split(string(output), "ADDRESS")[1], "\n")[0][2:]
//	addr = strings.ReplaceAll(addr, "\r", "")
//	fmt.Printf("THREADSTACK0基址:%v\n", addr)
//
//	r,err := strconv.ParseUint(addr, 0, 0)
//	if err != nil {
//		log.Fatal(err)
//	}
//	return uintptr(r)
//}

// 无冷却监测
//func modifyCD(hd windows.Handle, baseAddr uintptr){
//	// 通过获取THREADSTACK0，重新计算基址
//	THREADSTACK0 := getTHREADSTACK0Addr(strconv.Itoa(pid))
//	value := readUint32(hd, THREADSTACK0-uintptr(0x00000204))
//	baseAddr = uintptr(value)
//	cdOffset := []int64{0x0,0x8,0x15C,0x4C}
//	for{
//		if openNOCD {
//			addr := readDynamicAddr(hd, baseAddr, cdOffset)
//			for i:=0; i<10; i++ {											// 10个植物卡片
//				writeUint32(hd, addr+uintptr(80*i), 10000)			// 向日葵冷却周期为0~750，0代表冷却完成；取10000代表冷却上限
//			}
//		}
//		time.Sleep(time.Millisecond * 500)
//	}
//}

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

// 读取n个字节
func readNBytes(hd windows.Handle, addr uintptr, n int) []byte {
	data := make([]byte, n)
	err = windows.ReadProcessMemory(hd, addr, &data[0], uintptr(n), nil)
	if err != nil {
		log.Println("读取内存失败！")
	}

	fmt.Printf("读取%d个字节:", n)
	for i:=0;i<len(data);i++{
		fmt.Printf("0x%X, ", data[i])
	}
	return data
}

// 写入字节切片
func writeNBytes(hd windows.Handle, addr uintptr, data []byte) {
	var writeSize uintptr
	err = windows.WriteProcessMemory(hd, addr, &data[0], uintptr(len(data)), &writeSize)
	if err != nil {
		log.Println("写入内存失败！")
	}
	fmt.Printf("成功写入%d个字节\n", writeSize)
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

	myApp.Settings().SetTheme(&myTheme.MyTheme{})			// 设置APP主题，嵌入字体，解决乱码
	myWindow.Resize(fyne.NewSize(250,150))			// 设置窗口大小
	myWindow.CenterOnScreen()								// 窗口居中显示
	myWindow.SetMaster()									// 设置为主窗口

	// 阳光
	sunshineLabel := widget.NewLabel("阳光值")
	sunshineEntry := widget.NewEntry()
	sunshineButton := widget.NewButton("修改", func() {
		data, err := strconv.Atoi(sunshineEntry.Text)
		if err != nil {
			return
		}
		modifySunshine(handle, baseAddr, uint32(data))
	})
	c1 := container.NewGridWithColumns(3, sunshineLabel, sunshineEntry, sunshineButton)

	// 金币
	moneyLabel := widget.NewLabel("金币值")
	moneyEntry := widget.NewEntry()
	moneyButton := widget.NewButton("修改", func() {
		data, err := strconv.Atoi(moneyEntry.Text)
		if err != nil {
			return
		}
		modifyMoney(handle, baseAddr, uint32(data))
	})
	c2 := container.NewGridWithColumns(3, moneyLabel, moneyEntry, moneyButton)

	// 冷却
	CDLabel := widget.NewLabel("卡槽自动刷新")
	CDCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			openZeroCD(handle, baseAddr)
		}else {
			closeZeroCD(handle, baseAddr)
		}
	})
	c3 := container.NewGridWithColumns(3, CDLabel, layout.NewSpacer(), CDCheck)

	lockSunshineLabel := widget.NewLabel("无限阳光")
	lockSunshineCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			lockSunshine(handle, baseAddr)
		}else {
			unlockSunshine(handle, baseAddr)
		}
	})
	c4 := container.NewGridWithColumns(3, lockSunshineLabel, layout.NewSpacer(), lockSunshineCheck)

	lockBloodLabel := widget.NewLabel("植物锁血")
	lockBloodCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			lockPlantsBlood(handle, baseAddr)
		}else {
			unlockPlantsBlood(handle, baseAddr)
		}
	})
	c5 := container.NewGridWithColumns(3, lockBloodLabel, layout.NewSpacer(), lockBloodCheck)

	lockMagneticLabel := widget.NewLabel("磁力菇无CD")
	lockMagneticCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			lockMagneticCD(handle, baseAddr)
		}else {
			unlockMagneticCD(handle, baseAddr)
		}
	})
	c6 := container.NewGridWithColumns(3, lockMagneticLabel, layout.NewSpacer(), lockMagneticCheck)

	lockCannonLabel := widget.NewLabel("加农炮零冷却")
	lockCannonCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			lockCannonCD(handle, baseAddr)
		}else {
			unlockCannonCD(handle, baseAddr)
		}
	})
	c7 := container.NewGridWithColumns(3, lockCannonLabel, layout.NewSpacer(), lockCannonCheck)

	// 末日菇
	lockDoomLabel := widget.NewLabel("修复末日菇弹坑")
	lockDoomCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			lockDoomCD(handle, baseAddr)
		}else {
			unlockDoomCD(handle, baseAddr)
		}
	})
	c8 := container.NewGridWithColumns(3, lockDoomLabel, layout.NewSpacer(), lockDoomCheck)

	// 一键开启/关闭
	openAllLabel := widget.NewLabel("一键开启")
	openAllCheck := widget.NewCheck("开启", func(b bool) {
		if b {
			CDCheck.SetChecked(true)
			lockSunshineCheck.SetChecked(true)
			lockBloodCheck.SetChecked(true)
			lockMagneticCheck.SetChecked(true)
			lockCannonCheck.SetChecked(true)
			lockDoomCheck.SetChecked(true)
		}else {
			CDCheck.SetChecked(false)
			lockSunshineCheck.SetChecked(false)
			lockBloodCheck.SetChecked(false)
			lockMagneticCheck.SetChecked(false)
			lockCannonCheck.SetChecked(false)
			lockDoomCheck.SetChecked(false)
		}
	})
	c0 := container.NewGridWithColumns(3, openAllLabel, layout.NewSpacer(), openAllCheck)

	c := container.NewVBox(c1, c2, widget.NewSeparator(), c0, c3, c4, c5, c6, c7, c8)
	myWindow.SetContent(c)			// 创建导航
	myWindow.ShowAndRun()			// 事件循环
}