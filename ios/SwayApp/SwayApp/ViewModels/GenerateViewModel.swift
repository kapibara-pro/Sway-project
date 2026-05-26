import Foundation

@MainActor
final class GenerateViewModel: ObservableObject {
    @Published var mode: AssistMode = .reply
    @Published var language: AssistLanguage = .zhCN
    @Published var tone: AssistTone = .gentle
    @Published var inputPolicy: InputPolicy = .storeAllowed
    @Published var peerMessage = ""
    @Published var draft = ""
    @Published var candidates: [Candidate] = []
    @Published var requestID: String?
    @Published var statusMessage: String?
    @Published var isLoading = false

    func generate(appState: AppState) async {
        guard let backendURL = appState.backendBaseURL else {
            statusMessage = "后端地址格式不正确"
            return
        }

        let request = GenerateRequest(
            mode: mode,
            source: "app",
            inputPolicy: inputPolicy,
            peerMessage: peerMessage,
            draft: draft,
            tone: tone,
            relationshipStage: .unknown,
            language: language,
            length: "short",
            count: 3,
            clientContext: .init(platform: "ios", appVersion: "0.1.0", locale: Locale.current.identifier)
        )

        isLoading = true
        defer { isLoading = false }

        do {
            let response = try await appState.apiClient.generate(baseURL: backendURL, deviceID: appState.deviceID, payload: request)
            candidates = response.candidates
            requestID = response.requestID
            statusMessage = "已生成 3 条候选"
        } catch let apiError as APIErrorResponse {
            candidates = []
            requestID = apiError.requestID
            statusMessage = apiError.fallbackSuggestion ?? apiError.message
        } catch {
            candidates = []
            statusMessage = error.localizedDescription
        }
    }
}
