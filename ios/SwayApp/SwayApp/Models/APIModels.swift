import Foundation

enum AssistMode: String, CaseIterable, Identifiable, Codable {
    case rewrite
    case reply
    case opener
    case comfort
    case apologize
    case reject

    var id: String { rawValue }

    var title: String {
        switch self {
        case .rewrite: return "改写草稿"
        case .reply: return "回复"
        case .opener: return "破冰"
        case .comfort: return "安慰"
        case .apologize: return "道歉"
        case .reject: return "拒绝"
        }
    }
}

enum AssistLanguage: String, CaseIterable, Identifiable, Codable {
    case zhCN = "zh-CN"
    case enUS = "en-US"

    var id: String { rawValue }
    var title: String { self == .zhCN ? "中文" : "English" }
}

enum AssistTone: String, CaseIterable, Identifiable, Codable {
    case gentle
    case humorous
    case flirty
    case sincere
    case concise
    case proactive
    case restrained

    var id: String { rawValue }

    var title: String {
        switch self {
        case .gentle: return "温柔"
        case .humorous: return "幽默"
        case .flirty: return "暧昧"
        case .sincere: return "真诚"
        case .concise: return "短一点"
        case .proactive: return "主动"
        case .restrained: return "克制"
        }
    }
}

enum RelationshipStage: String, CaseIterable, Identifiable, Codable {
    case stranger
    case earlyChat = "early_chat"
    case dating
    case couple
    case conflict
    case unknown

    var id: String { rawValue }
}

enum InputPolicy: String, CaseIterable, Identifiable, Codable {
    case ephemeral
    case storeAllowed = "store_allowed"

    var id: String { rawValue }
}

struct ProviderStatus: Codable, Equatable {
    let configured: Bool
    let provider: String?
    let model: String?
    let keyMask: String?
    let updatedAt: String?
    let lastTestOK: Bool

    enum CodingKeys: String, CodingKey {
        case configured
        case provider
        case model
        case keyMask = "key_mask"
        case updatedAt = "updated_at"
        case lastTestOK = "last_test_ok"
    }
}

struct ProviderConfigRequest: Codable {
    var provider: String
    var baseURL: String
    var model: String
    var apiKey: String

    enum CodingKeys: String, CodingKey {
        case provider
        case baseURL = "base_url"
        case model
        case apiKey = "api_key"
    }
}

struct ProviderTestResponse: Codable {
    let ok: Bool
    let status: ProviderStatus
}

struct GenerateRequest: Codable {
    struct ClientContext: Codable {
        let platform: String
        let appVersion: String
        let locale: String

        enum CodingKeys: String, CodingKey {
            case platform
            case appVersion = "app_version"
            case locale
        }
    }

    let mode: AssistMode
    let source: String
    let inputPolicy: InputPolicy
    let peerMessage: String
    let draft: String
    let tone: AssistTone
    let relationshipStage: RelationshipStage
    let language: AssistLanguage
    let length: String
    let count: Int
    let clientContext: ClientContext

    enum CodingKeys: String, CodingKey {
        case mode
        case source
        case inputPolicy = "input_policy"
        case peerMessage = "peer_message"
        case draft
        case tone
        case relationshipStage = "relationship_stage"
        case language
        case length
        case count
        case clientContext = "client_context"
    }
}

struct GenerateResponse: Codable {
    let requestID: String
    let candidates: [Candidate]
    let safety: Safety
    let usage: Usage

    enum CodingKeys: String, CodingKey {
        case requestID = "request_id"
        case candidates
        case safety
        case usage
    }
}

struct Candidate: Identifiable, Codable, Equatable {
    let id: String
    let text: String
    let toneLabel: String
    let scenarioLabel: String
    let riskLevel: String
    let whyThisWorks: String?

    enum CodingKeys: String, CodingKey {
        case id
        case text
        case toneLabel = "tone_label"
        case scenarioLabel = "scenario_label"
        case riskLevel = "risk_level"
        case whyThisWorks = "why_this_works"
    }
}

struct Safety: Codable {
    let blocked: Bool
    let reason: String?
}

struct Usage: Codable {
    let provider: String
    let model: String
    let promptTokens: Int
    let completionTokens: Int
    let totalTokens: Int
    let latencyMS: Int

    enum CodingKeys: String, CodingKey {
        case provider
        case model
        case promptTokens = "prompt_tokens"
        case completionTokens = "completion_tokens"
        case totalTokens = "total_tokens"
        case latencyMS = "latency_ms"
    }
}

struct APIErrorResponse: Codable, Error {
    let code: String
    let message: String
    let fallbackSuggestion: String?
    let requestID: String?
    let details: [String: JSONValue]?

    enum CodingKeys: String, CodingKey {
        case code
        case message
        case fallbackSuggestion = "fallback_suggestion"
        case requestID = "request_id"
        case details
    }
}

enum JSONValue: Codable, Equatable {
    case string(String)
    case number(Double)
    case bool(Bool)
    case null

    init(from decoder: Decoder) throws {
        let container = try decoder.singleValueContainer()
        if container.decodeNil() {
            self = .null
        } else if let value = try? container.decode(Bool.self) {
            self = .bool(value)
        } else if let value = try? container.decode(Double.self) {
            self = .number(value)
        } else {
            self = .string(try container.decode(String.self))
        }
    }

    func encode(to encoder: Encoder) throws {
        var container = encoder.singleValueContainer()
        switch self {
        case .string(let value):
            try container.encode(value)
        case .number(let value):
            try container.encode(value)
        case .bool(let value):
            try container.encode(value)
        case .null:
            try container.encodeNil()
        }
    }
}
