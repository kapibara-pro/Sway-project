import SwiftUI
import UIKit

struct GenerateView: View {
    @EnvironmentObject private var appState: AppState
    @StateObject private var viewModel = GenerateViewModel()

    var body: some View {
        NavigationStack {
            Form {
                Section("模式") {
                    Picker("模式", selection: $viewModel.mode) {
                        ForEach(AssistMode.allCases) { mode in
                            Text(mode.title).tag(mode)
                        }
                    }

                    Picker("语言", selection: $viewModel.language) {
                        ForEach(AssistLanguage.allCases) { language in
                            Text(language.title).tag(language)
                        }
                    }
                    .pickerStyle(.segmented)
                }

                Section("上下文") {
                    TextEditor(text: $viewModel.peerMessage)
                        .frame(minHeight: 96)
                        .overlay(alignment: .topLeading) {
                            if viewModel.peerMessage.isEmpty {
                                Text("粘贴对方消息或聊天上下文")
                                    .foregroundStyle(.secondary)
                                    .padding(.top, 8)
                                    .padding(.leading, 4)
                            }
                        }

                    TextEditor(text: $viewModel.draft)
                        .frame(minHeight: 80)
                        .overlay(alignment: .topLeading) {
                            if viewModel.draft.isEmpty {
                                Text("可选：输入你原本想说的话")
                                    .foregroundStyle(.secondary)
                                    .padding(.top, 8)
                                    .padding(.leading, 4)
                            }
                        }
                }

                Section("语气") {
                    Picker("语气", selection: $viewModel.tone) {
                        ForEach(AssistTone.allCases) { tone in
                            Text(tone.title).tag(tone)
                        }
                    }

                    Picker("保存策略", selection: $viewModel.inputPolicy) {
                        Text("保存 1 个月").tag(InputPolicy.storeAllowed)
                        Text("不保存原文").tag(InputPolicy.ephemeral)
                    }
                    .pickerStyle(.segmented)
                }

                Section {
                    Button {
                        Task { await viewModel.generate(appState: appState) }
                    } label: {
                        HStack {
                            if viewModel.isLoading {
                                ProgressView()
                            }
                            Text(viewModel.isLoading ? "生成中" : "生成 3 条候选")
                        }
                    }
                    .disabled(viewModel.isLoading)
                }

                if let statusMessage = viewModel.statusMessage {
                    Section("状态") {
                        Text(statusMessage)
                            .foregroundStyle(.secondary)
                        if let requestID = viewModel.requestID {
                            Text("request_id: \(requestID)")
                                .font(.footnote)
                                .foregroundStyle(.secondary)
                        }
                    }
                }

                if !viewModel.candidates.isEmpty {
                    Section("3 条候选") {
                        ForEach(viewModel.candidates) { candidate in
                            CandidateRow(candidate: candidate)
                        }
                    }
                }
            }
            .navigationTitle("生成回复")
        }
    }
}

private struct CandidateRow: View {
    let candidate: Candidate
    @State private var copied = false

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            Text(candidate.text)
                .font(.body)

            HStack {
                Text(candidate.toneLabel)
                Text(candidate.scenarioLabel)
                Text(candidate.riskLevel)
                Spacer()
                Button(copied ? "已复制" : "复制") {
                    UIPasteboard.general.string = candidate.text
                    copied = true
                }
                .buttonStyle(.bordered)
            }
            .font(.footnote)
            .foregroundStyle(.secondary)
        }
        .padding(.vertical, 4)
    }
}
