# Releasing

以发布 1.1.0 为例，上一版是 1.0.0。

1. 在 `main` 上提交完要发布的改动，同步 `origin`，跑测试：

   ```bash
   git pull --ff-only origin main
   go test ./internal/...
   ```

2. 把 `mygo.json` 的 `version` 改成 `1.1.0`，在 `CHANGELOG.md` 顶部加一节 `## 1.1.0`，用中文写用户能感知到的变化（参考 `git log v1.0.0..HEAD`）。两处改动放在一个提交里，标题只写版本号 `1.1.0`。发布页的说明取自这一节，没有这一节工作流会失败。

3. 先推 `main`，推成功后再打 tag、推 tag：

   ```bash
   git push origin main
   git tag v1.1.0
   git push origin v1.1.0
   ```

   tag 打错了提交时，先取消它触发的运行，再删掉 tag 和它建的草稿：

   ```bash
   git push origin :refs/tags/v1.1.0
   gh release delete v1.1.0 --repo chemark/xingshu-music --yes
   ```

4. `.github/workflows/release.yml` 构建 Apple 芯片和 Intel 两个 dmg，签名、公证后以 `xingshu-music-arm64.dmg`、`xingshu-music-amd64.dmg` 上传到 tag 的草稿 release，同时上传两个架构的签名更新包（`.tar.gz`）和更新清单（`update-darwin-*.json`）。等它跑完：

   ```bash
   gh run watch "$(gh run list --repo chemark/xingshu-music --workflow release.yml -L1 --json databaseId -q '.[0].databaseId')" --repo chemark/xingshu-music --exit-status
   ```

   失败了就不要发布，先修好。

5. 下载草稿里的 dmg，确认已公证：

   ```bash
   gh release download v1.1.0 --repo chemark/xingshu-music --pattern '*.dmg' --dir /tmp/xingshu-release
   for f in /tmp/xingshu-release/*.dmg; do
     xcrun stapler validate "$f"
     spctl -a -t open --context context:primary-signature -v "$f"
   done
   ```

   同时确认每个架构的更新清单都包含本次版本号，URL 指向本仓库该 tag 的更新包，且对应文件已上传。更新包和清单必须来自同一次构建，不手动改写签名或替换已有版本的更新包。

6. 发布草稿，标为最新：

   ```bash
   gh release edit v1.1.0 --repo chemark/xingshu-music --draft=false --latest
   ```

7. 把 `site/index.html` 里所有的版本号（首屏按钮和下载区）改成 `v1.1.0`，提交并推 `main`，再在 `site/` 运行 `wrangler deploy`，然后检查 https://music.hoshikihao.com 显示的是新版本。

## Secrets

工作流用到仓库 Settings → Secrets and variables → Actions 里的 6 个 Secrets：

| 名称 | 内容 | 获取方式 |
| --- | --- | --- |
| `MYGO_UPDATER_PRIVATE_KEY` | MyGo 更新包的 Ed25519 签名私钥 | 首次用 `go tool mygo keygen` 生成，将公钥写入 `mygo.json` 的 `updates.publicKey`；私钥用 `gh secret set MYGO_UPDATER_PRIVATE_KEY --repo chemark/xingshu-music < <私钥文件路径>` 上传，保留仓库外的安全备份，不输出或提交私钥 |
| `MACOS_CERTIFICATE` | Developer ID Application 证书（含私钥）的 p12，base64 编码 | 钥匙串访问 →「登录」→「我的证书」，选中 `Developer ID Application: hao hoshiki (6SKPUQN55Z)`（展开能看到私钥），右键导出为 `.p12` 并设一个密码；再用 `base64 -i 证书.p12 \| gh secret set MACOS_CERTIFICATE --repo chemark/xingshu-music` 写入，写完删掉 `.p12` 文件 |
| `MACOS_CERTIFICATE_PASSWORD` | 导出 p12 时设的密码 | 同上 |
| `APPLE_API_KEY` | CI 专用的 App Store Connect API 密钥 `.p8` 的内容，供 `notarytool` 公证 | App Store Connect →「用户和访问」→「集成」→「App Store Connect API」→ 团队密钥 → 生成，名称 `xingshu-ci`，访问权限「开发者」，下载 `AuthKey_<密钥 ID>.p8`（只能下载一次）；用 `gh secret set APPLE_API_KEY --repo chemark/xingshu-music < AuthKey_<密钥 ID>.p8` 写入，写完删掉 `.p8` 文件 |
| `APPLE_API_KEY_ID` | 这把密钥的 ID | 密钥列表里的「密钥 ID」，也是 `.p8` 文件名里的那一段 |
| `APPLE_API_ISSUER_ID` | 团队的 Issuer ID | 同一页密钥列表上方的「Issuer ID」 |

`MACOS_CERTIFICATE_PASSWORD`、`APPLE_API_KEY_ID`、`APPLE_API_ISSUER_ID` 可以用 `gh secret set <名称> --repo chemark/xingshu-music` 在提示处粘贴写入。密钥泄露时在同一页吊销 `xingshu-ci`，再生成一把替换 `APPLE_API_KEY` 和 `APPLE_API_KEY_ID`。

## Automatic updates

更新源固定为 `chemark/xingshu-music`。首次启用默认自动检查和后台下载安装，之后保留用户在设置中的选择；更新在下次启动时生效。旧版未启用更新器，需要手动覆盖安装一次。

本机签名密钥备份位于用户配置目录的 `mygo/update-keys/`，私钥权限为 `600`。发布必须使用同一把私钥；不要重新生成或替换公钥，否则已安装版本无法验证后续更新。工作流通过 `MYGO_UPDATER_PRIVATE_KEY` 签名，缺少 Secret 时会失败。

## The speech recognition libraries

播客字幕用的 sherpa-onnx 库由 app 从本仓库的 `sherpa-onnx-2.0.1` release 下载，由本仓库的 Developer ID 签名。发布新版不用动它。

只有升级 sherpa-onnx 时才重新打包：改 `scripts/sherpa/main.go` 的 `version` 和压缩包的 digest，然后：

```bash
go run ./scripts/sherpa -sign "Developer ID Application: hao hoshiki (6SKPUQN55Z)" -out build/sherpa
gh release create sherpa-onnx-<version> --repo chemark/xingshu-music --prerelease --latest=false --title "sherpa-onnx <version>" --notes-file <说明文件> build/sherpa/*
```

- 打包会重写 `internal/speech/natives.go`；每次打包的 digest 都不同，上传的文件必须和提交的 `natives.go` 来自同一次打包。
- 必须带 `--prerelease --latest=false`，否则会占掉官网下载链接用的 `releases/latest`。
- 已发布版本的清单写死了这些文件的 digest：不替换、不删除已有 tag 的文件，换库一律用新 tag。
- release 说明写明 sherpa-onnx（Apache-2.0）、ONNX Runtime 和 Silero VAD（MIT）的来源和许可证。
