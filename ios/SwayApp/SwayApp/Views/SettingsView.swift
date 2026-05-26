import SwiftUI

struct SettingsView: View {
    @EnvironmentObject private var appState: AppState

    var body: some View {
        NavigationStack {
            Form {
                Section("设备") {
                    Text(appState.deviceID)
                        .font(.footnote)
                        .textSelection(.enabled)
                    Text("M1 使用游客设备身份，后续可迁移到登录账号。")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                Section("隐私") {
                    Text("默认可按 store_allowed 保存历史和样本 1 个月；也可以在生成页选择 ephemeral，不保存原文。")
                        .foregroundStyle(.secondary)
                    Text("言和不会自动读取微信、QQ、Telegram 等第三方 App 的完整聊天内容，只处理你主动输入、粘贴或导入的上下文。")
                        .foregroundStyle(.secondary)
                }
            }
            .navigationTitle("设置")
        }
    }
}
