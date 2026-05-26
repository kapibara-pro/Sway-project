import Foundation

@MainActor
final class ProviderSettingsViewModel: ObservableObject {
    @Published var provider = "mock"
    @Published var baseURL = ""
    @Published var model = "mock-chat"
    @Published var apiKey = ""
    @Published var status: ProviderStatus?
    @Published var message: String?
    @Published var isLoading = false

    func loadStatus(appState: AppState) async {
        guard let backendURL = appState.backendBaseURL else {
            message = "后端地址格式不正确"
            return
        }

        await run {
            status = try await appState.apiClient.providerStatus(baseURL: backendURL, deviceID: appState.deviceID)
        }
    }

    func useMockProvider() {
        provider = "mock"
        baseURL = ""
        model = "mock-chat"
        apiKey = ""
        message = "已填入 mock provider，可直接保存并测试"
    }

    func save(appState: AppState) async {
        guard let backendURL = appState.backendBaseURL else {
            message = "后端地址格式不正确"
            return
        }

        let config = ProviderConfigRequest(provider: provider, baseURL: baseURL, model: model, apiKey: apiKey)
        await run {
            try await appState.apiClient.saveProvider(baseURL: backendURL, deviceID: appState.deviceID, config: config)
            status = try await appState.apiClient.providerStatus(baseURL: backendURL, deviceID: appState.deviceID)
            message = "模型配置已保存"
        }
    }

    func test(appState: AppState) async {
        guard let backendURL = appState.backendBaseURL else {
            message = "后端地址格式不正确"
            return
        }

        await run {
            let response = try await appState.apiClient.testProvider(baseURL: backendURL, deviceID: appState.deviceID)
            status = response.status
            message = response.ok ? "连接测试通过" : "连接测试失败"
        }
    }

    func clear(appState: AppState) async {
        guard let backendURL = appState.backendBaseURL else {
            message = "后端地址格式不正确"
            return
        }

        await run {
            try await appState.apiClient.deleteProvider(baseURL: backendURL, deviceID: appState.deviceID)
            status = try await appState.apiClient.providerStatus(baseURL: backendURL, deviceID: appState.deviceID)
            message = "模型配置已清除"
        }
    }

    private func run(_ operation: () async throws -> Void) async {
        isLoading = true
        defer { isLoading = false }

        do {
            try await operation()
        } catch let apiError as APIErrorResponse {
            message = apiError.fallbackSuggestion ?? apiError.message
        } catch {
            message = error.localizedDescription
        }
    }
}
