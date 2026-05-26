import Foundation
import UIKit
import Vision

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
    @Published var isImportingImage = false

    var peerMessagePlaceholder: String {
        switch mode {
        case .rewrite:
            return "改写模式可留空，重点填写“我的草稿”"
        case .reply:
            return "粘贴对方消息或聊天上下文"
        case .opener:
            return "粘贴对方资料、动态或你想开启的话题"
        case .comfort:
            return "粘贴对方的负面情绪、抱怨或低落表达"
        case .apologize:
            return "描述发生了什么、对方在意什么"
        case .reject:
            return "粘贴对方请求，或写下需要拒绝的事情"
        }
    }

    var draftPlaceholder: String {
        switch mode {
        case .rewrite:
            return "输入你原本想说的话"
        case .reply:
            return "可选：输入你的原始想法"
        case .opener:
            return "可选：写下你想给人的感觉"
        case .comfort:
            return "可选：写下你想传达的关心"
        case .apologize:
            return "可选：写下你想承认或解释的部分"
        case .reject:
            return "可选：写下拒绝原因或边界"
        }
    }

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
            appState.addHistoryEntry(
                GenerationHistoryEntry(
                    id: UUID(),
                    createdAt: Date(),
                    requestID: response.requestID,
                    mode: mode,
                    language: language,
                    tone: tone,
                    inputPolicy: inputPolicy,
                    peerMessage: peerMessage,
                    draft: draft,
                    candidates: response.candidates
                )
            )
        } catch let apiError as APIErrorResponse {
            candidates = []
            requestID = apiError.requestID
            statusMessage = apiError.fallbackSuggestion ?? apiError.message
        } catch {
            candidates = []
            statusMessage = error.localizedDescription
        }
    }

    func pasteContextFromClipboard() {
        guard let text = UIPasteboard.general.string, !text.trimmingCharacters(in: .whitespacesAndNewlines).isEmpty else {
            statusMessage = "剪贴板里没有可用文本"
            return
        }
        peerMessage = text
        statusMessage = "已粘贴上下文"
    }

    func importTextFromImageData(_ data: Data?) async {
        guard let data, let image = UIImage(data: data), let cgImage = image.cgImage else {
            statusMessage = "无法读取图片"
            return
        }

        isImportingImage = true
        defer { isImportingImage = false }

        do {
            let text = try await recognizeText(in: cgImage)
            if text.isEmpty {
                statusMessage = "没有识别到文字，可以改用粘贴"
            } else {
                peerMessage = text
                statusMessage = "已从图片识别文字，可编辑后再生成"
            }
        } catch {
            statusMessage = "OCR 识别失败，可以改用粘贴"
        }
    }

    private func recognizeText(in cgImage: CGImage) async throws -> String {
        try await withCheckedThrowingContinuation { continuation in
            let request = VNRecognizeTextRequest { request, error in
                if let error {
                    continuation.resume(throwing: error)
                    return
                }

                let observations = request.results as? [VNRecognizedTextObservation] ?? []
                let text = observations
                    .compactMap { $0.topCandidates(1).first?.string }
                    .joined(separator: "\n")
                    .trimmingCharacters(in: .whitespacesAndNewlines)
                continuation.resume(returning: text)
            }
            request.recognitionLevel = .accurate
            request.usesLanguageCorrection = true
            request.recognitionLanguages = ["zh-Hans", "en-US"]

            let handler = VNImageRequestHandler(cgImage: cgImage)
            do {
                try handler.perform([request])
            } catch {
                continuation.resume(throwing: error)
            }
        }
    }
}
