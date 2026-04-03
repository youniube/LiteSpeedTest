# LiteSpeedTest sing-box 外挂内核改造包

这个包按当前 `xxf098/LiteSpeedTest` 的实际模块名 `github.com/xxf098/lite-proxy` 写的。

## 目录说明

- `engine/`：新增的外部核心抽象和 sing-box runner
- `config/vless.go`：新增的 VLESS 解析和 link 生成
- `request/ping_dial.go`：新增的通用拨号 ping
- `download/dial.go`：新增的通用拨号测速入口
- `proxy/socksdial.go`：SOCKS5 CONNECT 拨号
- `proxy/localsocks_client.go`：给 `core` 本地代理模式用的 tunnel client
- `core/config.go`：完整替换
- `core/core.go`：完整替换
- `main.go`：完整替换
- `PATCH_EXISTING_FILES.diff`：给 `config/config.go`、`config/parser.go`、`web/profile.go` 打补丁

## 建议替换顺序

1. 先把本包里的新文件复制到你的仓库对应路径
2. 用本包里的版本覆盖：
   - `core/config.go`
   - `core/core.go`
   - `main.go`
3. 再应用 `PATCH_EXISTING_FILES.diff`
4. 本地安装 sing-box，确认 `sing-box version` 可执行
5. 编译前先 `go fmt ./...`
6. 再 `go build ./...`

## 先验证什么

### 1. 单节点本地代理

```bash
./lite --engine singbox --singbox-bin sing-box 'vless://UUID@host:443?security=tls&type=ws&path=%2Fws&host=example.com&sni=example.com#test'
```

### 2. 命令行测速

在 `config.json` 里额外加：

```json
{
  "engine": "singbox",
  "singboxBin": "sing-box",
  "singboxWorkDir": ".lite-singbox",
  "keepTempFile": false
}
```

再跑：

```bash
./lite --config config.json --test sub.txt
```

## 我认为最可能还要你手动微调的地方

- `web/profile.go` 的 patch：因为这个文件较大，如果你本地仓库和 master 有少量差异，可能要手动贴进去
- `config/parser.go`：如果你本地已经自己改过 clash 解析，也要手动合并 `case "vless"`
- `core` 的本地代理模式：我这里已经给了 `LocalSocksClient`，但如果你本地 fork 过 `tunnel.Address` 的实现，`addr.String()` 那一行可能要改一下

## 说明

这套代码我是在无法本地联网拉源码编译的前提下，按仓库当前公开结构对齐写的，所以我更建议你：

- 先复制这些文件
- 执行 `go build ./...`
- 把第一轮编译报错贴给我

这样我下一轮可以按你的实际报错继续补齐。
