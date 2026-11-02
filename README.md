# 星树音乐

macOS 上的第三方网易云音乐客户端，用 Go 和 [MyGo](https://github.com/egoist/mygo) 写成，原生界面，无 WebView。

星树音乐源自 [egoist/mygo-cloudmusic](https://github.com/egoist/mygo-cloudmusic)（MyGo 云音乐），在它的基础上换了名字和图标，用自己的 Developer ID 签名并公证，去掉了自动更新，并会持续合并上游的改动。原项目的版权归原作者所有，本项目沿用同样的 [AGPL-3.0](LICENSE) 许可证。

## 功能

- 扫码登录，登录态保存在系统凭据存储中
- 每日推荐、推荐歌单、排行榜、我的歌单，收藏别人的歌单
- 搜索单曲、歌手、专辑、歌单、用户与播客，输入时显示搜索建议，保留搜索历史
- 歌手页的热门与最新歌曲、专辑与动态，关注歌手；专辑页与用户页
- 播客：浏览与收听节目，订阅播客；在本机把节目转写成字幕（Qwen3-ASR），像歌词一样随播放滚动
- 私人 FM（推荐 / 熟悉 / 探索）与心动模式
- 边下边播，支持标准 / 较高 / 极高 / 无损（FLAC）/ Hi-Res 音质
- 听过的歌曲缓存在本机，再次播放无需等待
- 音量均衡，暂停与继续时淡出淡入，拖动进度无杂音
- 退出后再打开，从上次的播放队列与进度继续
- 逐字歌词与翻译，正在播放页的「律动」视图、「选歌」封面墙与歌曲评论
- 刘海播放器：在屏幕顶部刘海处显示迷你播放器，悬停展开（在设置中开启）
- 菜单栏图标，可以打开窗口、控制播放
- 喜欢歌曲、播放队列、随机与循环播放
- 浅色 / 深色 / 跟随系统

## 系统要求

macOS 13 或更高版本，支持 Apple 芯片和 Intel 芯片的 Mac。

## 下载安装

1. 打开 [Releases](https://github.com/chemark/xingshu-music/releases/latest)，下载对应芯片的安装包：Apple 芯片（M 系列）选 `xingshu-music-<版本>-arm64.dmg`，Intel 芯片选 `xingshu-music-<版本>-amd64.dmg`。不确定时打开左上角苹果菜单 →「关于本机」，看「芯片」或「处理器」一项。
2. 打开 dmg，把「星树音乐」拖进「应用程序」文件夹。

安装包已经过 Apple 公证，可以直接打开。星树音乐不会自动更新，有新版本时到 Releases 下载，按同样的方式覆盖安装。

## 构建

需要 Go 1.27+。

```bash
go tool mygo dev     # 开发模式运行
go tool mygo build -sign - -skip-notarize   # 打包到 build/，用临时签名、不公证
go test ./...        # 运行测试
```

## 致谢

- [MyGo 云音乐](https://github.com/egoist/mygo-cloudmusic)：星树音乐的原项目，作者 [EGOIST](https://github.com/egoist)
- [MyGo](https://github.com/egoist/mygo)：原生 UI 与应用框架
- [NeteaseCloudMusicApi](https://github.com/Binaryify/NeteaseCloudMusicApi)：eapi 加密与登录流程的参考
- [folia-major](https://github.com/chthollyphile/folia-major)：歌词动画效果的参考
- [sherpa-onnx](https://github.com/k2-fsa/sherpa-onnx)、[Qwen3-ASR](https://github.com/QwenLM/Qwen3-ASR) 与 [Silero VAD](https://github.com/snakers4/silero-vad)：本机语音识别

本项目与网易云音乐官方无关，音乐版权归各权利人所有。

## 许可证

[AGPL-3.0](LICENSE)。原项目 © EGOIST 及贡献者；星树音乐的修改 © 星树浩。
