# Releasing

以发布 1.1.0 为例，上一版是 1.0.0。

1. 在 `main` 上提交完要发布的改动，同步 `origin`，跑测试：

   ```bash
   git pull --ff-only origin main
   go test ./internal/...
   ```

2. 把 `mygo.json` 的 `version` 改成 `1.1.0`，在 `CHANGELOG.md` 顶部加一节 `## 1.1.0`，用中文写用户能感知到的变化（参考 `git log v1.0.0..HEAD`），这期间合并进来的上游改动也写进去。两处改动放在一个提交里，标题只写版本号 `1.1.0`。发布页的说明取自这一节，没有这一节工作流会失败。

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

4. `.github/workflows/release.yml` 构建 Apple 芯片和 Intel 两个 dmg，签名、公证后上传到 tag 的草稿 release。等它跑完：

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

6. 发布草稿，标为最新：

   ```bash
   gh release edit v1.1.0 --repo chemark/xingshu-music --draft=false --latest
   ```

## Secrets

工作流用到仓库 Settings → Secrets and variables → Actions 里的 5 个 Secrets：

| 名称 | 内容 | 获取方式 |
| --- | --- | --- |
| `MACOS_CERTIFICATE` | Developer ID Application 证书（含私钥）的 p12，base64 编码 | 钥匙串访问 →「登录」→「我的证书」，选中 `Developer ID Application: hao hoshiki (6SKPUQN55Z)`（展开能看到私钥），右键导出为 `.p12` 并设一个密码；再用 `base64 -i 证书.p12 \| gh secret set MACOS_CERTIFICATE --repo chemark/xingshu-music` 写入，写完删掉 `.p12` 文件 |
| `MACOS_CERTIFICATE_PASSWORD` | 导出 p12 时设的密码 | 同上 |
| `APPLE_API_KEY` | CI 专用的 App Store Connect API 密钥 `.p8` 的内容，供 `notarytool` 公证 | App Store Connect →「用户和访问」→「集成」→「App Store Connect API」→ 团队密钥 → 生成，名称 `xingshu-ci`，访问权限「开发者」，下载 `AuthKey_<密钥 ID>.p8`（只能下载一次）；用 `gh secret set APPLE_API_KEY --repo chemark/xingshu-music < AuthKey_<密钥 ID>.p8` 写入，写完删掉 `.p8` 文件 |
| `APPLE_API_KEY_ID` | 这把密钥的 ID | 密钥列表里的「密钥 ID」，也是 `.p8` 文件名里的那一段 |
| `APPLE_API_ISSUER_ID` | 团队的 Issuer ID | 同一页密钥列表上方的「Issuer ID」 |

`MACOS_CERTIFICATE_PASSWORD`、`APPLE_API_KEY_ID`、`APPLE_API_ISSUER_ID` 可以用 `gh secret set <名称> --repo chemark/xingshu-music` 在提示处粘贴写入。密钥泄露时在同一页吊销 `xingshu-ci`，再生成一把替换 `APPLE_API_KEY` 和 `APPLE_API_KEY_ID`。

## The speech recognition libraries

播客字幕用的 sherpa-onnx 库由 app 从上游的 `sherpa-onnx-2.0.1` release 下载，由上游的 Developer ID 签名，靠 `entitlements.plist` 的 `disable-library-validation` 加载。本仓库不重新打包、不发布这些库。
