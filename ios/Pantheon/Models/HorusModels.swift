import Foundation

// MARK: - Horus (Code Graph)

struct HorusSymbol: Codable, Identifiable {
    var id: String { "\(file):\(name):\(line)" }

    let name: String
    let kind: String
    let file: String
    let line: Int
    let endLine: Int
    let signature: String
    let doc: String?
    let exported: Bool
    let parent: String?

    var kindIcon: String {
        switch kind {
        case "func":      return "f()"
        case "method":    return "m()"
        case "type":      return "T"
        case "struct":    return "S"
        case "interface": return "I"
        case "const":     return "C"
        case "var":       return "V"
        case "package":   return "P"
        default:          return "?"
        }
    }
}

struct HorusSymbolGraph: Codable {
    let root: String
    let packages: [String]
    let symbols: [HorusSymbol]
    let stats: HorusGraphStats
    let builtAt: String
}

struct HorusGraphStats: Codable {
    let files: Int
    let packages: Int
    let types: Int
    let functions: Int
    let methods: Int
    let interfaces: Int
    let totalLines: Int
}

struct HorusOutlineResult: Codable {
    let outline: String
}

struct HorusContextResult: Codable {
    let context: String
}
