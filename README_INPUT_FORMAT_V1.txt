本覆盖包包含以下改造：

1. 新增 core/subscription 统一订阅解析层：
   - model.go
   - errors.go
   - detect.go
   - fetch.go
   - parser.go
   - convert.go
   - parse_uri.go
   - parse_base64.go
   - parse_clash.go
   - parse_surge.go
   - parse_loon.go
   - parse_qx.go
   - parse_singbox.go
   - normalize.go

2. 修改 web/profile.go：
   - ParseLinksWithOption 改为优先走统一解析入口
   - getSubscriptionLinks / parseProfiles / parseBase64 / parseClash / parseFile 改为新解析层包装

说明：
- 当前 v1 以“提取可测速节点”为目标，不做完整客户端配置解释。
- 可直接输出为 link 的协议优先为：ss / ssr / vmess / vless / trojan / http。
- socks5 当前先识别和提取，但未承诺进入现有测速链路闭环。
- 由于当前容器无法联网下载/校验 Go 依赖，未能完成完整 go test，只完成了 gofmt 与静态代码整理。
