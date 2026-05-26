import Foundation

final class SwayAPIClient {
    private let session: URLSession
    private let jsonEncoder: JSONEncoder
    private let jsonDecoder: JSONDecoder

    init(session: URLSession = .shared) {
        self.session = session
        self.jsonEncoder = JSONEncoder()
        self.jsonDecoder = JSONDecoder()
    }

    func providerStatus(baseURL: URL, deviceID: String) async throws -> ProviderStatus {
        let request = makeRequest(baseURL: baseURL, path: "/api/v1/providers/status", method: "GET", deviceID: deviceID)
        return try await send(request)
    }

    func saveProvider(baseURL: URL, deviceID: String, config: ProviderConfigRequest) async throws {
        var request = makeRequest(baseURL: baseURL, path: "/api/v1/providers/config", method: "PUT", deviceID: deviceID)
        request.httpBody = try jsonEncoder.encode(config)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        let _: EmptyResponse = try await send(request)
    }

    func testProvider(baseURL: URL, deviceID: String) async throws -> ProviderTestResponse {
        let request = makeRequest(baseURL: baseURL, path: "/api/v1/providers/test", method: "POST", deviceID: deviceID)
        return try await send(request)
    }

    func deleteProvider(baseURL: URL, deviceID: String) async throws {
        let request = makeRequest(baseURL: baseURL, path: "/api/v1/providers/config", method: "DELETE", deviceID: deviceID)
        let _: EmptyResponse = try await send(request)
    }

    func generate(baseURL: URL, deviceID: String, payload: GenerateRequest) async throws -> GenerateResponse {
        var request = makeRequest(baseURL: baseURL, path: "/api/v1/chat-assist/generate", method: "POST", deviceID: deviceID)
        request.httpBody = try jsonEncoder.encode(payload)
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        return try await send(request)
    }

    private func makeRequest(baseURL: URL, path: String, method: String, deviceID: String) -> URLRequest {
        let cleanedPath = path.hasPrefix("/") ? String(path.dropFirst()) : path
        let url = makeURL(baseURL: baseURL, cleanedPath: cleanedPath)
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.timeoutInterval = 8
        request.setValue(deviceID, forHTTPHeaderField: "X-Sway-Device-ID")
        return request
    }

    private func makeURL(baseURL: URL, cleanedPath: String) -> URL {
        guard var components = URLComponents(url: baseURL, resolvingAgainstBaseURL: false) else {
            return baseURL.appendingPathComponent(cleanedPath)
        }

        let basePath = components.path.trimmingCharacters(in: CharacterSet(charactersIn: "/"))
        components.path = "/" + [basePath, cleanedPath].filter { !$0.isEmpty }.joined(separator: "/")
        return components.url ?? baseURL.appendingPathComponent(cleanedPath)
    }

    private func send<T: Decodable>(_ request: URLRequest) async throws -> T {
        let (data, response) = try await session.data(for: request)

        guard let httpResponse = response as? HTTPURLResponse else {
            throw APIErrorResponse(code: "SERVICE_UNAVAILABLE", message: "Invalid response", fallbackSuggestion: "服务暂时不可用，请稍后重试", requestID: nil, details: nil)
        }

        if (200..<300).contains(httpResponse.statusCode) {
            if T.self == EmptyResponse.self || data.isEmpty {
                return EmptyResponse() as! T
            }
            return try jsonDecoder.decode(T.self, from: data)
        }

        if let apiError = try? jsonDecoder.decode(APIErrorResponse.self, from: data) {
            throw apiError
        }

        throw APIErrorResponse(code: "SERVICE_UNAVAILABLE", message: "HTTP \(httpResponse.statusCode)", fallbackSuggestion: "服务暂时不可用，请稍后重试", requestID: nil, details: nil)
    }
}

private struct EmptyResponse: Decodable {}
