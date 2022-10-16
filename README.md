# PlantsVsZombies

植物大战僵尸外挂辅助，纯go语言实现，界面使用[fyne](https://github.com/fyne-io/fyne)

### 主要功能

- 修改阳光值、金币值
- 植物卡槽自动刷新
- 无限阳光
- 无敌锁血
- 磁力菇无CD
- 玉米加农炮无CD
- 自动修复末日菇弹坑
- 一击秒杀僵尸！！！

### 编译

1. 安装go语言环境、fyne工具
2. 克隆项目
3. 编译打包

```shell
go mod tidy
go mod download
fyne package -os windows -icon logo.jpg --name app.exe
```

### 效果图

![image](https://raw.githubusercontent.com/ttmars/image/master/github/PVZ.jpg)

