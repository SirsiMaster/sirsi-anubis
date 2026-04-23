package jackal

// Advisory intelligence for scan findings.

// EnrichAdvisory populates Advisory, Remediation, CanFix, and Breaking
// on each finding based on the rule that generated it. This is the
// intelligence layer — Sirsi tells the user what it can do and what to expect.
func EnrichAdvisory(result *ScanResult) {
	for i := range result.Findings {
		f := &result.Findings[i]
		enrichFinding(f)
	}
}

func enrichFinding(f *Finding) {
	switch f.RuleName {

	// ── Caches (always safe, always rebuildable) ─────────────────
	case "system_caches":
		f.Advisory = "Application cache. Rebuilds automatically on next use."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "browser_caches":
		f.Advisory = "Browser cache. Rebuilds as you browse. No data loss."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "font_caches":
		f.Advisory = "Font rendering cache. Rebuilds automatically."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "mail_attachments_cache":
		f.Advisory = "Mail attachment cache. Re-downloads from server on access."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "time_machine_local":
		f.Advisory = "Local Time Machine snapshots. macOS recreates as needed."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "vscode_caches":
		f.Advisory = "VS Code extension and workspace caches. Rebuilds on launch."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "jetbrains_caches":
		f.Advisory = "JetBrains IDE caches. Rebuilds on project open."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "cursor_cache":
		f.Advisory = "Cursor IDE cache. Rebuilds on launch."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "windsurf_cache":
		f.Advisory = "Windsurf IDE cache. Rebuilds on launch."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "zed_cache":
		f.Advisory = "Zed editor cache. Rebuilds on launch."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "android_studio":
		f.Advisory = "Android Studio cache and build data. Rebuilds on sync."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "xcode_derived_data":
		f.Advisory = "Xcode build artifacts. Rebuilds on next build (Cmd+B)."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "neovim_cache":
		f.Advisory = "Neovim plugin and swap file cache. Plugins re-download."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "claude_code_cache":
		f.Advisory = "Claude Code cache. Rebuilds on next session."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "gemini_cli_cache":
		f.Advisory = "Gemini CLI cache. Rebuilds on next use."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "codex_cli_cache":
		f.Advisory = "Codex CLI cache. Rebuilds on next use."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "mlx_cache":
		f.Advisory = "MLX model cache. Re-downloads from hub on next use."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "metal_shader_cache":
		f.Advisory = "Metal shader compilation cache. GPU recompiles as needed."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "homebrew_cache":
		f.Advisory = "Homebrew downloaded packages. Re-downloads on next install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "cocoapods_cache":
		f.Advisory = "CocoaPods spec cache. Re-downloads on pod install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "spm_cache":
		f.Advisory = "Swift Package Manager cache. Re-resolves on next build."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "npm_global_cache":
		f.Advisory = "npm/yarn/pnpm global cache. Packages re-download on install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "gradle_cache":
		f.Advisory = "Gradle dependency cache. Re-downloads on next build."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "maven_cache":
		f.Advisory = "Maven local repository. Artifacts re-download on build."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "composer_cache":
		f.Advisory = "PHP Composer cache. Re-downloads on composer install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "rubygems_cache":
		f.Advisory = "RubyGems cache. Re-downloads on bundle install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "kubernetes_caches":
		f.Advisory = "Kubernetes client cache (kubeconfig cache, helm). Rebuilds on next command."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "terraform_caches":
		f.Advisory = "Terraform plugin cache. Re-downloads on terraform init."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "gcloud_caches":
		f.Advisory = "Google Cloud SDK logs and cache. Rebuilds on next gcloud command."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "firebase_caches":
		f.Advisory = "Firebase CLI cache. Rebuilds on next deploy."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "aws_cli_cache":
		f.Advisory = "AWS CLI cache. Rebuilds on next aws command."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "go_mod_cache":
		f.Advisory = "Go module download cache. Rebuilds on next go build."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "python_caches":
		f.Advisory = "Python pip/conda cache. Packages re-download on install."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "huggingface_cache":
		f.Advisory = "HuggingFace model cache. Models re-download on next use. May be large."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "ollama_models":
		f.Advisory = "Ollama model weights. Re-downloads with ollama pull."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "pytorch_cache":
		f.Advisory = "PyTorch hub cache. Re-downloads on next torch.hub.load()."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "tensorflow_cache":
		f.Advisory = "TensorFlow cache. Rebuilds on next import."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "onnx_cache":
		f.Advisory = "ONNX Runtime cache. Re-downloads on next inference."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "vllm_cache":
		f.Advisory = "vLLM model cache. Re-downloads from hub."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "jax_cache":
		f.Advisory = "JAX compilation cache. Recompiles on next run."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "stable_diffusion_models":
		f.Advisory = "Stable Diffusion model weights. May be large. Re-downloads from hub."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "langchain_cache":
		f.Advisory = "LangChain cache. Rebuilds on next chain execution."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "onedrive_cache", "google_drive_cache", "dropbox_cache", "icloud_cache":
		f.Advisory = "Cloud sync cache. Re-syncs from cloud on next access."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Logs & Reports ───────────────────────────────────────────
	case "system_logs":
		f.Advisory = "Old log files. No impact on running apps."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "nginx_logs":
		f.Advisory = "Nginx access/error logs. Rotates automatically."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "dev_log_files":
		f.Advisory = "Development server logs. No impact on running projects."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "crash_reports":
		f.Advisory = "Crash diagnostic reports. Keep if debugging a recent crash."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = false
	case "coverage_reports":
		f.Advisory = "Test coverage output. Regenerates on next test run."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "trash":
		f.Advisory = "Items already in Trash. Permanently removes them."
		f.Remediation = "Empty Trash"
		f.CanFix = true
	case "downloads_junk":
		f.Advisory = "Old files in Downloads. Review before cleaning."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Dev Artifacts (rebuildable but may take time) ─────────────
	case "node_modules":
		f.Advisory = "npm dependencies. Reinstalls with npm install (~30s–2min)."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "rust_targets":
		f.Advisory = "Rust compilation output. Rebuilds on next cargo build (~1–5min)."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "docker_desktop":
		f.Advisory = "Docker Desktop VM and cache data. Rebuilds on restart."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Build Output (review first) ──────────────────────────────
	case "build_output":
		f.Advisory = "Build output directory. May contain undeployed production builds."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true
	case "nextjs_cache":
		f.Advisory = "Next.js build cache. Rebuilds on next dev/build (~30s–2min)."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "turborepo_cache":
		f.Advisory = "Turborepo build cache. Rebuilds on next turbo run."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Docker ───────────────────────────────────────────────────
	case "docker_dangling_images":
		f.Advisory = "Untagged Docker images. No running containers use them."
		f.Remediation = "docker image prune"
		f.CanFix = true
	case "docker_buildkit_cache":
		f.Advisory = "Docker BuildKit layer cache. Rebuilds on next docker build."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Python Envs ──────────────────────────────────────────────
	case "python_venvs", "python_dot_venvs":
		f.Advisory = "Python virtual environment. Recreate with python -m venv. May have custom packages."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true

	// ── Virtualization ───────────────────────────────────────────
	case "parallels_full":
		f.Advisory = "Parallels Desktop data. May contain VMs with user data."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true
	case "vmware_fusion":
		f.Advisory = "VMware Fusion data. May contain VMs with user data."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true
	case "utm":
		f.Advisory = "UTM virtual machine data."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true
	case "virtualbox":
		f.Advisory = "VirtualBox data. May contain VMs."
		f.Remediation = "Move to Trash"
		f.CanFix = true
		f.Breaking = true

	// ── Git (Sirsi can fix these with git commands) ───────────────
	case "git_stale_branches":
		f.Advisory = "Local branch tracking a deleted remote. Safe to prune."
		f.Remediation = "git branch -D <branch>"
		f.CanFix = true
	case "git_merged_branches":
		f.Advisory = "Branch fully merged into main. Safe to delete."
		f.Remediation = "git branch -d <branch>"
		f.CanFix = true
	case "git_large_objects":
		f.Advisory = "Oversized .git directory. Sirsi can compact with git gc."
		f.Remediation = "git gc --aggressive --prune=now"
		f.CanFix = true
	case "git_orphaned_worktrees":
		f.Advisory = "Worktree whose directory was deleted. Safe to prune."
		f.Remediation = "git worktree prune"
		f.CanFix = true
	case "git_untracked_artifacts":
		f.Advisory = "Build artifacts not tracked by git. Safe to remove."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "git_rerere_cache":
		f.Advisory = "Old conflict resolution recordings. Safe to clear."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "git_reflog_bloat":
		f.Advisory = "Oversized reflog. Sirsi can expire old entries."
		f.Remediation = "git reflog expire --expire=30.days --all"
		f.CanFix = true

	// ── CI/CD ────────────────────────────────────────────────────
	case "github_actions_cache":
		f.Advisory = "Local GitHub Actions runner cache. Rebuilds on next workflow."
		f.Remediation = "Move to Trash"
		f.CanFix = true
	case "act_runner_cache":
		f.Advisory = "nektos/act local runner cache. Rebuilds on next act run."
		f.Remediation = "Move to Trash"
		f.CanFix = true

	// ── Repo Hygiene (informational — needs human judgment) ──────
	case "env_files":
		f.Advisory = "Contains secrets (API keys, passwords). Sirsi will add to .gitignore to prevent accidental commit."
		f.Remediation = "Add to .gitignore"
		f.CanFix = true
		f.Breaking = false
	case "stale_lock_files":
		f.Advisory = "Orphaned git lock file. May block git operations."
		f.Remediation = "Delete lock file"
		f.CanFix = true
	case "dead_symlinks":
		f.Advisory = "Broken symlink pointing to nonexistent target. Safe to remove."
		f.Remediation = "Delete symlink"
		f.CanFix = true
	case "oversized_repos":
		f.Advisory = "Repo exceeds 2 GB. Sirsi will compact .git with gc, repack, and prune loose objects."
		f.Remediation = "git gc --aggressive + repack + prune"
		f.CanFix = true
		f.Breaking = false

	// ── Ghost Residuals ──────────────────────────────────────────
	case "ka_ghost":
		if f.Severity == SeveritySafe {
			f.Advisory = "Dead app cache/log residual. The app is uninstalled. Safe to remove."
			f.Remediation = "Move to Trash"
			f.CanFix = true
		} else {
			f.Advisory = "Dead app preferences/data. The app is uninstalled but data may have value."
			f.Remediation = "Move to Trash (review first)"
			f.CanFix = true
			f.Breaking = false
		}

	default:
		// Fallback for any rule without specific advisory
		if f.Severity == SeveritySafe {
			f.Advisory = "Safe to clean. No impact on running applications."
			f.Remediation = "Move to Trash"
			f.CanFix = true
		} else if f.Severity == SeverityCaution {
			f.Advisory = "Review before cleaning. May affect development workflows."
			f.Remediation = "Move to Trash"
			f.CanFix = true
		} else {
			f.Advisory = "Flagged for review. Sirsi recommends manual inspection."
			f.Remediation = "Flag for review"
			f.CanFix = false
		}
	}
}
