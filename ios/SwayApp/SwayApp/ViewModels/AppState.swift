import Foundation

@MainActor
final class AppState: ObservableObject {
    @Published var backendBaseURLText: String {
        didSet { UserDefaults.standard.set(backendBaseURLText, forKey: Self.backendURLKey) }
    }

    let deviceID: String
    let apiClient: SwayAPIClient

    private static let backendURLKey = "sway.backendBaseURL"

    init(apiClient: SwayAPIClient = SwayAPIClient()) {
        self.apiClient = apiClient
        self.deviceID = DeviceIDStore.current()
        self.backendBaseURLText = UserDefaults.standard.string(forKey: Self.backendURLKey) ?? "http://127.0.0.1:8080"
    }

    var backendBaseURL: URL? {
        URL(string: backendBaseURLText.trimmingCharacters(in: .whitespacesAndNewlines))
    }
}
