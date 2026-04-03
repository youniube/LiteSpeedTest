应用方式：
1. 直接覆盖仓库中的 .github/workflows/test.yaml、release.yaml、cron.yaml
2. 本地在仓库根目录执行：
   go mod download
   go mod tidy
   go fmt ./...
   go build ./...
3. 把更新后的 go.mod 和 go.sum 一起提交

这三个 workflow 修复了两类问题：
- 去掉旧 workflow 里的 go get -u ./...，避免把依赖升级到与 go 1.19/1.20 不兼容的组合
- 在构建前增加 go mod download / go mod tidy，补齐 go.sum

说明：
- test.yaml 保留 Linux amd64 单平台，先保证 CI 跑通
- release.yaml 仅在打 tag 或手动触发时运行
- cron.yaml 改为单 Ubuntu 定时校验，避免旧版 develop 分支和 macOS matrix 带来的额外干扰
