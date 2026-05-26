import SwiftUI
import UIKit

struct HistoryView: View {
    @EnvironmentObject private var appState: AppState

    var body: some View {
        NavigationStack {
            List {
                if appState.historyEntries.isEmpty {
                    ContentUnavailableView("暂无历史", systemImage: "clock", description: Text("生成成功后会在这里保留最近记录。"))
                } else {
                    ForEach(appState.historyEntries) { entry in
                        HistoryEntryRow(entry: entry)
                    }
                }
            }
            .navigationTitle("历史记录")
            .toolbar {
                ToolbarItem(placement: .topBarTrailing) {
                    Button("清理") {
                        appState.clearHistory()
                    }
                    .disabled(appState.historyEntries.isEmpty)
                }
            }
        }
    }
}

private struct HistoryEntryRow: View {
    let entry: GenerationHistoryEntry
    @State private var copiedCandidateID: String?

    var body: some View {
        VStack(alignment: .leading, spacing: 10) {
            HStack {
                Text(entry.mode.title)
                    .font(.headline)
                Text(entry.language.title)
                    .font(.caption)
                    .foregroundStyle(.secondary)
                Spacer()
                Text(entry.createdAt, style: .date)
                    .font(.caption)
                    .foregroundStyle(.secondary)
            }

            if !entry.peerMessage.isEmpty {
                Text(entry.peerMessage)
                    .font(.subheadline)
                    .foregroundStyle(.secondary)
                    .lineLimit(2)
            }

            ForEach(entry.candidates) { candidate in
                VStack(alignment: .leading, spacing: 6) {
                    Text(candidate.text)
                        .lineLimit(3)
                    HStack {
                        Text(candidate.toneLabel)
                        Text(candidate.riskLevel)
                        Spacer()
                        Button(copiedCandidateID == candidate.id ? "已复制" : "复制") {
                            UIPasteboard.general.string = candidate.text
                            copiedCandidateID = candidate.id
                        }
                        .buttonStyle(.bordered)
                    }
                    .font(.caption)
                    .foregroundStyle(.secondary)
                }
                .padding(.vertical, 4)
            }

            Text("request_id: \(entry.requestID)")
                .font(.caption2)
                .foregroundStyle(.secondary)
        }
        .padding(.vertical, 6)
    }
}
