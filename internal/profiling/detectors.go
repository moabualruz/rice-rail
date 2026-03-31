package profiling

// FilePattern maps a file glob to what it detects.
type FilePattern struct {
	Glob     string
	Detector string // language, package_manager, build_system, tool, ci, framework
	Name     string
	Category string // for tools: linter, formatter, etc.
}

// LanguagePatterns maps file extensions to languages.
var LanguagePatterns = map[string]string{
	".go":     "go",
	".rs":     "rust",
	".ts":     "typescript",
	".tsx":    "typescript",
	".js":     "javascript",
	".jsx":    "javascript",
	".py":     "python",
	".java":   "java",
	".kt":     "kotlin",
	".cs":     "csharp",
	".rb":     "ruby",
	".php":    "php",
	".swift":  "swift",
	".c":      "c",
	".cpp":    "cpp",
	".h":      "c",
	".hpp":    "cpp",
	".zig":    "zig",
	".lua":    "lua",
	".ex":     "elixir",
	".exs":    "elixir",
	".dart":   "dart",
	".vue":    "vue",
	".svelte": "svelte",
	".slint":  "slint",
}

// ManifestPatterns maps manifest files to package managers.
var ManifestPatterns = []FilePattern{
	{Glob: "go.mod", Detector: "package_manager", Name: "go modules"},
	{Glob: "Cargo.toml", Detector: "package_manager", Name: "cargo"},
	{Glob: "package.json", Detector: "package_manager", Name: "npm/yarn/pnpm"},
	{Glob: "pyproject.toml", Detector: "package_manager", Name: "python (pyproject)"},
	{Glob: "setup.py", Detector: "package_manager", Name: "python (setuptools)"},
	{Glob: "requirements.txt", Detector: "package_manager", Name: "pip"},
	{Glob: "Pipfile", Detector: "package_manager", Name: "pipenv"},
	{Glob: "poetry.lock", Detector: "package_manager", Name: "poetry"},
	{Glob: "uv.lock", Detector: "package_manager", Name: "uv"},
	{Glob: "pom.xml", Detector: "package_manager", Name: "maven"},
	{Glob: "build.gradle", Detector: "package_manager", Name: "gradle"},
	{Glob: "build.gradle.kts", Detector: "package_manager", Name: "gradle (kotlin)"},
	{Glob: "*.csproj", Detector: "package_manager", Name: "dotnet"},
	{Glob: "*.sln", Detector: "package_manager", Name: "dotnet"},
	{Glob: "Gemfile", Detector: "package_manager", Name: "bundler"},
	{Glob: "composer.json", Detector: "package_manager", Name: "composer"},
	{Glob: "pubspec.yaml", Detector: "package_manager", Name: "dart/flutter"},
	{Glob: "mix.exs", Detector: "package_manager", Name: "mix"},
	{Glob: "Package.swift", Detector: "package_manager", Name: "swift pm"},
}

// BuildSystemPatterns maps build files to build systems.
var BuildSystemPatterns = []FilePattern{
	{Glob: "Makefile", Detector: "build_system", Name: "make"},
	{Glob: "justfile", Detector: "build_system", Name: "just"},
	{Glob: "Taskfile.yml", Detector: "build_system", Name: "taskfile"},
	{Glob: "CMakeLists.txt", Detector: "build_system", Name: "cmake"},
	{Glob: "meson.build", Detector: "build_system", Name: "meson"},
	{Glob: "BUILD", Detector: "build_system", Name: "bazel"},
	{Glob: "BUILD.bazel", Detector: "build_system", Name: "bazel"},
	{Glob: "WORKSPACE", Detector: "build_system", Name: "bazel"},
	{Glob: "Dockerfile", Detector: "build_system", Name: "docker"},
	{Glob: "docker-compose.yml", Detector: "build_system", Name: "docker compose"},
	{Glob: "docker-compose.yaml", Detector: "build_system", Name: "docker compose"},
	{Glob: "Earthfile", Detector: "build_system", Name: "earthly"},
	{Glob: "Tiltfile", Detector: "build_system", Name: "tilt"},
	{Glob: "nx.json", Detector: "build_system", Name: "nx"},
	{Glob: "turbo.json", Detector: "build_system", Name: "turborepo"},
	{Glob: "lerna.json", Detector: "build_system", Name: "lerna"},
}

// CIPatterns maps CI config files to providers.
var CIPatterns = []FilePattern{
	{Glob: ".github/workflows/*.yml", Detector: "ci", Name: "github actions"},
	{Glob: ".github/workflows/*.yaml", Detector: "ci", Name: "github actions"},
	{Glob: ".gitlab-ci.yml", Detector: "ci", Name: "gitlab ci"},
	{Glob: "Jenkinsfile", Detector: "ci", Name: "jenkins"},
	{Glob: ".circleci/config.yml", Detector: "ci", Name: "circleci"},
	{Glob: ".travis.yml", Detector: "ci", Name: "travis"},
	{Glob: "azure-pipelines.yml", Detector: "ci", Name: "azure devops"},
	{Glob: "bitbucket-pipelines.yml", Detector: "ci", Name: "bitbucket pipelines"},
	{Glob: ".buildkite/pipeline.yml", Detector: "ci", Name: "buildkite"},
}

// ToolConfigPatterns maps tool config files to development tools.
var ToolConfigPatterns = []FilePattern{
	// Go
	{Glob: ".golangci.yml", Detector: "tool", Name: "golangci-lint", Category: "linter"},
	{Glob: ".golangci.yaml", Detector: "tool", Name: "golangci-lint", Category: "linter"},

	// JavaScript/TypeScript
	{Glob: ".eslintrc*", Detector: "tool", Name: "eslint", Category: "linter"},
	{Glob: "eslint.config.*", Detector: "tool", Name: "eslint", Category: "linter"},
	{Glob: ".prettierrc*", Detector: "tool", Name: "prettier", Category: "formatter"},
	{Glob: "prettier.config.*", Detector: "tool", Name: "prettier", Category: "formatter"},
	{Glob: "tsconfig.json", Detector: "tool", Name: "typescript", Category: "typechecker"},
	{Glob: "biome.json", Detector: "tool", Name: "biome", Category: "linter"},
	{Glob: "biome.jsonc", Detector: "tool", Name: "biome", Category: "linter"},
	{Glob: "jest.config.*", Detector: "tool", Name: "jest", Category: "test_runner"},
	{Glob: "vitest.config.*", Detector: "tool", Name: "vitest", Category: "test_runner"},
	{Glob: "playwright.config.*", Detector: "tool", Name: "playwright", Category: "test_runner"},
	{Glob: "cypress.config.*", Detector: "tool", Name: "cypress", Category: "test_runner"},
	{Glob: ".dependency-cruiser*", Detector: "tool", Name: "dependency-cruiser", Category: "rule_engine"},

	// Python
	{Glob: "ruff.toml", Detector: "tool", Name: "ruff", Category: "linter"},
	{Glob: ".ruff.toml", Detector: "tool", Name: "ruff", Category: "linter"},
	{Glob: ".flake8", Detector: "tool", Name: "flake8", Category: "linter"},
	{Glob: "mypy.ini", Detector: "tool", Name: "mypy", Category: "typechecker"},
	{Glob: ".mypy.ini", Detector: "tool", Name: "mypy", Category: "typechecker"},
	{Glob: "pyrightconfig.json", Detector: "tool", Name: "pyright", Category: "typechecker"},
	{Glob: "pytest.ini", Detector: "tool", Name: "pytest", Category: "test_runner"},
	{Glob: "conftest.py", Detector: "tool", Name: "pytest", Category: "test_runner"},
	{Glob: "setup.cfg", Detector: "tool", Name: "setuptools", Category: "build_system"},
	{Glob: "tox.ini", Detector: "tool", Name: "tox", Category: "test_runner"},

	// Rust
	{Glob: "clippy.toml", Detector: "tool", Name: "clippy", Category: "linter"},
	{Glob: "rustfmt.toml", Detector: "tool", Name: "rustfmt", Category: "formatter"},
	{Glob: ".rustfmt.toml", Detector: "tool", Name: "rustfmt", Category: "formatter"},

	// Java/Kotlin
	{Glob: "checkstyle.xml", Detector: "tool", Name: "checkstyle", Category: "linter"},
	{Glob: "spotbugs*.xml", Detector: "tool", Name: "spotbugs", Category: "linter"},
	{Glob: "detekt.yml", Detector: "tool", Name: "detekt", Category: "linter"},
	{Glob: "ktlint*", Detector: "tool", Name: "ktlint", Category: "formatter"},

	// C#
	{Glob: ".editorconfig", Detector: "tool", Name: "editorconfig", Category: "formatter"},

	// Cross-language
	{Glob: ".semgrep.yml", Detector: "tool", Name: "semgrep", Category: "rule_engine"},
	{Glob: ".semgrep/", Detector: "tool", Name: "semgrep", Category: "rule_engine"},
	{Glob: "sgconfig.yml", Detector: "tool", Name: "ast-grep", Category: "rule_engine"},
	{Glob: ".comby", Detector: "tool", Name: "comby", Category: "codemod"},
	{Glob: ".pre-commit-config.yaml", Detector: "tool", Name: "pre-commit", Category: "linter"},
	{Glob: ".husky/", Detector: "tool", Name: "husky", Category: "linter"},
	{Glob: "lefthook.yml", Detector: "tool", Name: "lefthook", Category: "linter"},

	// Security
	{Glob: ".trivyignore", Detector: "tool", Name: "trivy", Category: "security"},
	{Glob: ".snyk", Detector: "tool", Name: "snyk", Category: "security"},
}

// MonorepoIndicators are files/dirs that suggest monorepo topology.
var MonorepoIndicators = []FilePattern{
	{Glob: "nx.json", Detector: "monorepo", Name: "nx"},
	{Glob: "turbo.json", Detector: "monorepo", Name: "turborepo"},
	{Glob: "lerna.json", Detector: "monorepo", Name: "lerna"},
	{Glob: "pnpm-workspace.yaml", Detector: "monorepo", Name: "pnpm workspaces"},
	{Glob: "WORKSPACE", Detector: "monorepo", Name: "bazel"},
	{Glob: "BUILD.bazel", Detector: "monorepo", Name: "bazel"},
}

// ArchitecturePatterns maps folder names to architecture hints.
var ArchitecturePatterns = map[string]string{
	"cmd":            "go-style multi-binary layout",
	"internal":       "go-style internal packages",
	"pkg":            "go-style public packages",
	"src":            "standard source directory",
	"lib":            "library code",
	"app":            "application entry points",
	"apps":           "monorepo applications",
	"packages":       "monorepo packages",
	"services":       "microservices or service layer",
	"domain":         "domain-driven design",
	"entities":       "DDD entities layer",
	"usecases":       "clean architecture use cases",
	"adapters":       "hexagonal/ports-and-adapters",
	"ports":          "hexagonal ports",
	"infrastructure": "infrastructure/adapter layer",
	"presentation":   "presentation/UI layer",
	"api":            "API layer",
	"handlers":       "request handlers",
	"controllers":    "MVC controllers",
	"models":         "data models",
	"views":          "MVC views or templates",
	"components":     "UI components",
	"pages":          "page-based routing",
	"routes":         "route definitions",
	"middleware":     "middleware layer",
	"migrations":     "database migrations",
	"tests":          "test directory",
	"e2e":            "end-to-end tests",
	"fixtures":       "test fixtures",
	"proto":          "protobuf/gRPC definitions",
	"schemas":        "schema definitions",
}
