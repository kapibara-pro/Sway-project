import SwiftUI

struct ProviderSettingsView: View {
    @EnvironmentObject private var appState: AppState
    @StateObject private var viewModel = ProviderSettingsViewModel()

    var body: some View {
        NavigationStack {
            Form {
                Section("后端") {
                    TextField("Base URL", text: $appState.backendBaseURLText)
                        .keyboardType(.URL)
                        .textInputAutocapitalization(.never)
                    Text("本地联调用 `http://127.0.0.1:8080`。真机访问 Mac 本机服务时需要替换为局域网 IP。")
                        .font(.footnote)
                        .foregroundStyle(.secondary)
                }

                Section("当前状态") {
                    if let status = viewModel.status {
                        LabeledContent("已配置", value: status.configured ? "是" : "否")
                        if let provider = status.provider {
                            LabeledContent("Provider", value: provider)
                        }
                        if let model = status.model {
                            LabeledContent("Model", value: model)
                        }
                        LabeledContent("测试", value: status.lastTestOK ? "通过" : "未通过")
                    } else {
                        Text("尚未加载")
                            .foregroundStyle(.secondary)
                    }

                    Button("刷新状态") {
                        Task { await viewModel.loadStatus(appState: appState) }
                    }
                }

                Section("模型配置") {
                    TextField("Provider", text: $viewModel.provider)
                        .textInputAutocapitalization(.never)
                    TextField("Base URL", text: $viewModel.baseURL)
                        .keyboardType(.URL)
                        .textInputAutocapitalization(.never)
                    TextField("Model", text: $viewModel.model)
                        .textInputAutocapitalization(.never)
                    SecureField("API Key", text: $viewModel.apiKey)

                    Button("填入 Mock Provider") {
                        viewModel.useMockProvider()
                    }

                    Button("保存配置") {
                        Task { await viewModel.save(appState: appState) }
                    }
                    .disabled(viewModel.isLoading)

                    Button("测试连接") {
                        Task { await viewModel.test(appState: appState) }
                    }
                    .disabled(viewModel.isLoading)

                    Button("清除配置", role: .destructive) {
                        Task { await viewModel.clear(appState: appState) }
                    }
                    .disabled(viewModel.isLoading)
                }

                if let message = viewModel.message {
                    Section("提示") {
                        Text(message)
                            .foregroundStyle(.secondary)
                    }
                }
            }
            .navigationTitle("模型配置")
            .task {
                await viewModel.loadStatus(appState: appState)
            }
        }
    }
}
