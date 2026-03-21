# 𓂀 Sirsi Anubis — Launch Copy

## Product Hunt Copy

**Tagline:** Infrastructure hygiene for the AI era — Anubis cleans what AI agents leave behind.

**Description:**

Every cleaning tool treats your machine as a consumer device. But if you're a developer or AI engineer, your machine is a workstation. It accumulates a completely different class of junk.

Sirsi Anubis scans, judges, and purges infrastructure waste — from 200 GB of stale HuggingFace model caches to ghost apps haunting your Spotlight, to zombie Node processes eating 5.7 GB of RAM.

Built as an Egyptian-themed CLI toolkit, Anubis speaks the language of judgment:

- `anubis weigh` — discover waste across 7 domains, 64 scan rules
- `anubis ka` — hunt ghost apps (remnants of uninstalled applications)
- `anubis mcp` — connect to Claude, Cursor, or Windsurf as a context sanitizer
- `anubis install-brain` — download neural models for semantic file classification
- `anubis scales enforce` — enforce hygiene policies across your fleet

Free forever. Open source under MIT. Written in Go. Ships as two binaries under 10 MB total. No telemetry. No tracking. Zero footprint with `--stealth` mode.

*The jackal sees everything. Nothing escapes the Weighing.*

**Maker's Comment:**

I built Anubis after spending 3 hours manually cleaning up Parallels VMs and realizing no tool understood what I, as a developer, needed cleaned. CleanMyMac doesn't know about `__pycache__`. Mole doesn't see ghost `.plist` files. Nobody touches AI model caches.

Anubis is the open-source answer: scan, judge, purge.

The MCP integration is the secret weapon — your AI coding assistant can now understand your machine's hygiene before indexing your workspace.

**Tags:** Developer Tools, Command Line, Open Source, AI, macOS

---

## Hacker News Copy

**Title:** Show HN: Sirsi Anubis – Open-source infrastructure hygiene CLI for developers and AI engineers

**Description:**

I built Anubis because existing Mac cleaners don't understand developer workstations. After spending hours cleaning Parallels remnants, I realized nobody makes a tool that knows about:

- 200 GB of stale HuggingFace/Ollama model caches
- Ghost apps leaving .plist files and Spotlight registrations behind
- Zombie LSP servers eating 17 GB of RAM
- Docker containers from projects you abandoned 6 months ago

Anubis has 64 scan rules across 7 domains, a ghost app hunter (Ka), a neural file classifier (Brain), and an MCP server so your AI coding assistant can use it as a context sanitizer.

Written in Go. Two binaries under 10 MB. No telemetry, no tracking, no analytics. MIT licensed.

The Egyptian theme is intentional — every module is named after mythology: Jackal (scanner), Ka (ghost hunter), Scales (policy engine), Hapi (resource optimizer), Scarab (fleet sweep).

`go install github.com/SirsiMaster/sirsi-anubis/cmd/anubis@latest`

GitHub: https://github.com/SirsiMaster/sirsi-anubis

---

## Twitter/X Launch Thread

**Tweet 1:**
𓂀 Introducing Sirsi Anubis — infrastructure hygiene for the AI era.

Every Mac cleaner treats your machine as a consumer device.
Anubis treats it as a developer workstation.

64 scan rules. Ghost app hunting. Neural classification. MCP server.
Free. Open source. MIT.

github.com/SirsiMaster/sirsi-anubis

**Tweet 2:**
What makes Anubis different?

• Finds AI/ML caches (HuggingFace, Ollama, MLX) — often 50-200 GB
• Hunts ghost apps lingering after uninstall (.plist, caches, Spotlight)
• Kills zombie dev processes eating your RAM
• Classifies files semantically (junk vs project vs config)

**Tweet 3:**
The secret weapon: `anubis mcp`

Connect Anubis to Claude Code, Cursor, or Windsurf as a context sanitizer.

Your AI assistant cleans your workspace before indexing it.

No more "why is my AI indexing node_modules?"

**Tweet 4:**
No telemetry. No tracking. Zero footprint.

`--stealth` mode: Anubis comes, judges, and vanishes.

Written in Go. Under 10 MB. Ships today.

*The jackal sees everything. Nothing escapes the Weighing.* 𓂀
