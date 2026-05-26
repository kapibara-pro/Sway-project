# Sway iOS

M1 iOS 主 App 最小闭环：

- 游客设备 ID：通过 `X-Sway-Device-ID` 请求头传给后端。
- Provider 配置：状态、保存、测试、清除。
- 生成页：六模式参数骨架、中文/英文、`store_allowed` / `ephemeral`、调用生成 API、展示 3 条候选、复制候选。

## 本地联调

1. 启动后端：

```bash
cd backend
go run ./cmd/server
```

2. 用 Xcode 打开 `ios/Sway.xcodeproj`。
3. 运行 `SwayApp` target。
4. 在“模型”页确认后端地址：

```text
http://127.0.0.1:8080
```

真机访问 Mac 本机服务时，需要替换成 Mac 的局域网 IP。

5. 点击“填入 Mock Provider” -> “保存配置” -> “测试连接”。
6. 到“生成”页输入上下文，点击“生成 3 条候选”。

## 说明

当前环境没有 Xcode，工程文件和 Swift 源码已按 iOS 17 SwiftUI App 编写；构建验证需要在 macOS/Xcode 上执行。
