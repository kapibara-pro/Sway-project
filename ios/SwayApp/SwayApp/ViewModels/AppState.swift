import Foundation

@MainActor
final class AppState: ObservableObject {
    @Published var backendBaseURLText: String {
        didSet { UserDefaults.standard.set(backendBaseURLText, forKey: Self.backendURLKey) }
    }
    @Published private(set) var historyEntries: [GenerationHistoryEntry]

    let deviceID: String
    let apiClient: SwayAPIClient

    private static let backendURLKey = "sway.backendBaseURL"
    private static let historyKey = "sway.generationHistory"

    init(apiClient: SwayAPIClient = SwayAPIClient()) {
        self.apiClient = apiClient
        self.deviceID = DeviceIDStore.current()
        self.backendBaseURLText = UserDefaults.standard.string(forKey: Self.backendURLKey) ?? "http://127.0.0.1:8080"
        self.historyEntries = Self.loadHistory()
    }

    var backendBaseURL: URL? {
        URL(string: backendBaseURLText.trimmingCharacters(in: .whitespacesAndNewlines))
    }

    func addHistoryEntry(_ entry: GenerationHistoryEntry) {
        historyEntries.insert(entry, at: 0)
        historyEntries = Array(historyEntries.prefix(100))
        saveHistory()
    }

    func clearHistory() {
        historyEntries = []
        UserDefaults.standard.removeObject(forKey: Self.historyKey)
    }

    private func saveHistory() {
        guard let data = try? JSONEncoder().encode(historyEntries) else { return }
        UserDefaults.standard.set(data, forKey: Self.historyKey)
    }

    private static func loadHistory() -> [GenerationHistoryEntry] {
        guard let data = UserDefaults.standard.data(forKey: historyKey),
              let entries = try? JSONDecoder().decode([GenerationHistoryEntry].self, from: data) else {
            return []
        }
        return entries
    }
}
