import Foundation

// MARK: - Seba (Hardware Detection)

struct HardwareProfile: Codable {
    let cpuCores: Int
    let cpuModel: String
    let cpuArch: String
    let totalRam: Int64
    let gpu: GPUInfo?
    let neuralEngine: Bool?
    let os: String?
    let kernel: String?

    var formattedRAM: String {
        ByteCountFormatter.string(fromByteCount: totalRam, countStyle: .memory)
    }
}

struct GPUInfo: Codable {
    let type: String
    let name: String
    let vram: Int64?
    let metalFamily: String?
    let cudaVersion: String?
    let driverVer: String?
    let compute: String?
}

struct AcceleratorProfile: Codable {
    let hasGpu: Bool
    let gpuCores: Int?
    let gpuVendor: String?
    let hasAne: Bool
    let aneCores: Int?
    let hasMetal: Bool
    let hasCuda: Bool
    let hasRocm: Bool
    let hasOneapi: Bool
    let cpuCores: Int
    let memBandwidth: String?
    let unifiedMemory: Bool?
    let routing: String?
}
