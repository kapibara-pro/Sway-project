import Foundation

struct DeviceIDStore {
    private static let key = "sway.deviceID"

    static func current() -> String {
        let defaults = UserDefaults.standard
        if let existing = defaults.string(forKey: key), !existing.isEmpty {
            return existing
        }

        let newID = "ios-" + UUID().uuidString.lowercased()
        defaults.set(newID, forKey: key)
        return newID
    }
}
