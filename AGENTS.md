# AGENTS.md

## Build

- 用 `go tool mygo build -skip-notarize -skip-dmg` 打包；先退出正在运行的 app，再用 `build/darwin-arm64/星树音乐.app` 替换 `/Applications/星树音乐.app`。
- `mygo.json` 里的 `name`、`identifier`、`signingIdentity`、`entitlements`、`notarize` 保持不变。
- 不要加回更新功能：`mygo.json` 的 `updates`、`useUpdater`、「检查更新」菜单项。
- 在 `main` 上工作，`main` 跟踪 `origin`（`chemark/xingshu-music`）；上游改动通过合并 `upstream/main`（`egoist/mygo-cloudmusic`）获取。
- `website/` 是上游的官网，不改动。

## Release

- 按 `RELEASING.md` 发布：推 `v<version>` tag 触发 `.github/workflows/release.yml`，只构建 macOS 的 `darwin/arm64` 和 `darwin/amd64` 两个 dmg，签名、公证后上传到草稿 release，检查后手动发布。
- 不用 `mygo build -upload`，不生成、不上传更新清单。
- 版本号从 1.0.0 起独立编号，不跟上游；每次发布在 `CHANGELOG.md` 顶部加 `## <version>` 一节。
- 合并上游时，`mygo.json` 的 `version` 和 `CHANGELOG.md` 顶部的冲突保留本仓库的版本，上游的新变化写进下一版的 `## <version>` 一节。

## Design

- 设计稿放 `design/<主题>/`（如 `design/app-icon/`），文件名用小写英文加连字符。
- `design/` 只保留最终采用的设计稿，代码和构建不引用它；采用的方案落到 `icon/`、`resources/` 等正式位置。
- 过程稿、落选方案和临时文件放在临时目录，定稿后删掉，不提交进 `design/`。

## Tests

只为真的可能出错的行为写测试。被测代码改了测试却察觉不到，或者措辞上无害的改动就会让测试失败，这样的测试不要写。

**Write tests for:**

- 输入带来的效果：点击、按键、拖动、手势和滚动，以及它们导致的状态或效果（松开进度条时跳转、打开正在播放后按空格、后退与前进、甩动封面、喜欢一首歌、退出电台）。
- 不依赖视图的逻辑：解析歌词、格式化时间和数量、播放器的队列、跳转、淡入淡出和缓存、解码音频流和图片。
- 刚修好的回归问题。在测试注释里写明当时出了什么错。
- 帧预算，写在 `perf_test.go` 里，确保空闲的视图不会开始请求不需要的帧。

**Don't write tests that:**

- 检查一组标签是否显示在屏幕上。这只是重复视图代码，措辞一变就失败。
- 检查默认设置，或表单里各行的顺序。
- 测试 MyGo 本身，比如 `AnimateTo` 或弹簧动画是否从 0 走到 1。
- 依赖真实时钟。`ui.NewTester` 的时钟从当前时间开始，所以任何相位取自时钟的东西，通过与否取决于测试运行的时刻。另外，`FramePending` 不计入等待 `After` 或 `Painter.After` 的帧。

**How:**

- 用 `ui.NewTester` 构建视图，用 `newTestApp`（在 `internal/app/helpers_test.go`）准备 app。按无障碍标签找元素，然后 `Click`、`RightClick`、`Press`/`Move`/`Release` 或 `Key`，之后检查 app 的状态。
- 通过 `likeSong`、`copyText` 这类包级变量替换网络和剪贴板，并用 `defer` 换回真实实现。
- 调用网易云的在线测试，没设置 `LIVE` 时跳过。
- 视觉效果用眼看，不要写断言。`NM_FRAMES` 测试把 PNG 存进该目录，没设置时跳过。只在检查覆盖范围时比对像素，比如某一层是否仍盖满窗口。
- 每个测试开头用一句注释说明它检查的行为，风格与代码一致。
- 提交前运行 `go test ./internal/...`，新的 UI 测试用 `-count=5` 运行。不稳定的测试要修好或删掉，不要靠重跑。
