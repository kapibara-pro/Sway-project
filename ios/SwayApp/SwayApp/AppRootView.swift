import SwiftUI

struct AppRootView: View {
    @StateObject private var appState = AppState()

    var body: some View {
        TabView {
            GenerateView()
                .tabItem {
                    Label("生成", systemImage: "sparkles")
                }

            ProviderSettingsView()
                .tabItem {
                    Label("模型", systemImage: "server.rack")
                }

            HistoryView()
                .tabItem {
                    Label("历史", systemImage: "clock.arrow.circlepath")
                }

            SettingsView()
                .tabItem {
                    Label("设置", systemImage: "gearshape")
                }
        }
        .environmentObject(appState)
    }
}
