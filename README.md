替换文件：
- .github/workflows/test.yaml

作用：
- GitHub Actions 改为构建 Windows 64 位版本
- 产物为 lite-windows-amd64.zip
- 保留前端构建、wasm 构建、go mod download / tidy

使用方法：
1. 解压
2. 覆盖仓库中的 .github/workflows/test.yaml
3. 提交并推送
4. 在 Actions 里重新运行 Build Windows

说明：
- 这是单平台 Windows 构建版
- 如果后面你要 Linux + Windows 双平台，我再直接给你新的补丁包
