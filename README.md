<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/license-MIT-green?style=for-the-badge" alt="MIT License" />
  <img src="https://img.shields.io/github/v/release/yv3000/SIIG?style=for-the-badge&color=blueviolet" alt="Latest Release" />
</p>

<p align="center">
  <img src="https://img.shields.io/badge/Windows-0078D6?style=for-the-badge&logo=windows&logoColor=white" alt="Windows" />
  <img src="https://img.shields.io/badge/macOS-000000?style=for-the-badge&logo=apple&logoColor=white" alt="macOS" />
  <img src="https://img.shields.io/badge/Linux-FCC624?style=for-the-badge&logo=linux&logoColor=black" alt="Linux" />
</p>

<h1 align="center">SIIG</h1>
<p align="center"><b>Zero-config CLI that auto-populates ignore files based on your project's tech stack.</b></p>
<p align="center">One command. No prompts. No config. Under 200ms. 122 stacks supported out of the box — and any stack beyond that is fetched automatically on demand.</p>

---

## What is SIIG?

**SIIG** scans your project directory, detects the tech stack, and automatically writes the right entries to `.gitignore`, `.vercelignore`, `.dockerignore`, and `.vscode/settings.json` — so you never accidentally push a `.env` file, API keys, build artifacts, or `node_modules` again.

```
$ siig
✓ Detected: Node.js, Next.js, Docker
✓ .gitignore    — 24 entries added
✓ .vercelignore — 4 entries added
✓ .dockerignore — 8 entries added
⚠ .vscode/settings.json — skipped (no .vscode dir found)
```

- You run `siig` once. It figures out everything else.
- Works on any project — new or existing, any language, any framework.
- **122 stacks embedded** — works fully offline, no internet needed for common stacks.
- **Unknown stack?** SIIG silently fetches it from [siig-rules](https://github.com/yv3000/siig-rules), caches it locally for 7 days, and applies it — no manual input needed. The lookup uses the unrecognised file's own name as the stack id, so `gleam.toml` resolves the `gleam` rule.
- **Idempotent** — running it twice produces the same result. Safe to run anytime.
- **Never touches source code** — only reads file/dir names, only writes to ignore files.

> **Think of it as a one-time setup that actually knows what to ignore.** Instead of copy-pasting from gitignore.io or forgetting half the entries, SIIG just handles it — correctly, instantly, for whatever stack you're running.

---

## Install

SIIG is a single binary. No runtime required. No Node.js, no Python — just download and run.

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/yv3000/SIIG/main/scripts/install.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/SIIG/main/scripts/install.sh | sh
```

The installer will:
1. Detect your OS and architecture (amd64 / arm64)
2. Download the correct binary from GitHub Releases
3. Verify the SHA256 checksum before installing
4. Place `siig` in `~/.siig/bin/` (or `%USERPROFILE%\.siig\bin\` on Windows)
5. Add it to your PATH automatically

---

## Uninstall

### Windows

```powershell
irm https://raw.githubusercontent.com/yv3000/SIIG/main/scripts/uninstall.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/yv3000/SIIG/main/scripts/uninstall.sh | sh
```

Removes the binary, cleans up the PATH entry and deletes `~/.siig` (the fetched-rule cache). Nothing else is touched — SIIG never modifies anything outside your project directory during normal use.

---

## Proof it works

```bash
# Mixed stack project — Node + Python + Docker + Terraform
$ siig --verbose

✓ Detected: Node.js, Python, Docker, Terraform

Entries by Target:
  gitignore: .env, .env.*, node_modules/, __pycache__/, .terraform/, *.tfstate ...
  dockerignore: node_modules/, __pycache__/, .env, *.pyc ...

✓ .gitignore    — 74 entries added
✓ .dockerignore — 11 entries added
⚠ .vercelignore — skipped (not applicable)
⚠ .npmignore    — skipped (not applicable)

# Run it again — nothing changes
$ siig
✓ Detected: Node.js, Python, Docker, Terraform
✓ Everything already up to date — no changes needed

# Unknown stack auto-fetched from siig-rules
$ siig --verbose
✓ Detected: Node.js, Gleam (online)
✓ .gitignore    — 18 entries added
```

---

## Usage

```bash
siig                          # Auto-detect stack and populate ignore files
siig --dry-run                # Preview what would be added — nothing written
siig --check                  # Exit 1 if ignore files are missing entries (for CI)
siig --json                   # One JSON object on stdout instead of the human summary
siig --verbose                # Show every rule applied, every entry added
siig --targets git,docker     # Only populate specific targets
siig --offline                # Skip network — use embedded rules and cache only
siig --no-cache               # Ignore rules cached from earlier runs, fetch fresh
siig --clear-cache            # Delete the locally cached remote rules
siig --version                # Print version
siig --help                   # Print help
```

### Diagnostics

Things the user cannot see happening — a path that could not be read, a rule cache entry that failed to write — are reported on **stderr** and never written into an ignore file. `--verbose` emits them as JSON records so a script can read them; the summary on stdout stays human-readable.

### Targets

| Flag value | File written |
|---|---|
| `git` | `.gitignore` |
| `vercel` | `.vercelignore` |
| `docker` | `.dockerignore` |
| `npm` | `.npmignore` |
| `vscode` | `.vscode/settings.json` |

Example: `siig --targets git,vercel,docker`

---

## Output Targets

| File | Written When |
|---|---|
| `.gitignore` | Always |
| `.vercelignore` | `vercel.json` or Next.js detected |
| `.dockerignore` | `Dockerfile` or `docker-compose.yml` detected |
| `.npmignore` | Library `package.json` detected (`"main"` or `"exports"` field) |
| `.vscode/settings.json` | `.vscode/` directory exists in project |

---

## Supported Stacks (122)

The category tables below are highlights. The complete list, generated from `rules/db/`, is at the end of this section.

### Languages

| Stack | Detected Via |
|---|---|
| C | `*.c`, `Makefile`, `CMakeLists.txt` |
| C++ | `*.cpp`, `CMakeLists.txt`, `vcpkg.json` |
| C# | `*.csproj`, `*.sln` |
| .NET | `*.csproj`, `*.sln` |
| F# | `*.fsproj`, `*.fs` |
| Go | `go.mod` |
| Haskell | `*.cabal`, `stack.yaml` |
| Java | `pom.xml`, `build.gradle` |
| Julia | `Project.toml`, `*.jl` |
| Kotlin (JVM) | `build.gradle.kts`, `settings.gradle.kts` |
| Lua | `*.lua`, `*.rockspec` |
| MATLAB | `*.m`, `*.mlx` |
| Nix | `flake.nix`, `default.nix` |
| Objective-C | `*.m`, `Podfile`, `*.xcodeproj` |
| Perl | `*.pl`, `cpanfile` |
| PHP | `composer.json` |
| Python | `requirements.txt`, `pyproject.toml`, `Pipfile` |
| R | `DESCRIPTION`, `*.Rproj`, `renv.lock` |
| Ruby | `Gemfile` |
| Rust | `Cargo.toml` |
| Scala | `build.sbt`, `*.scala` |
| Swift | `Package.swift`, `*.xcodeproj` |
| Zig | `build.zig` |
| Clojure | `project.clj`, `deps.edn` |
| Elixir | `mix.exs` |
| Erlang | `rebar.config`, `*.erl` |
| Groovy | `*.groovy`, `Jenkinsfile` |
| Gleam | `gleam.toml` |
| Nim | `nim.cfg`, `*.nimble`, `config.nims` |
| OCaml | `dune-project`, `*.opam` |
| Crystal | `shard.yml` |
| Fortran | `fpm.toml`, `*.f90` |
| D | `dub.sdl`, `dub.json` |
| PureScript | `spago.dhall`, `packages.dhall` |
| Lean 4 | `lean-toolchain`, `lakefile.lean` |
| Coq | `_CoqProject` |

### Frontend Frameworks

| Stack | Detected Via |
|---|---|
| Angular | `angular.json` |
| Astro | `astro.config.*` |
| Gatsby | `gatsby-config.*` |
| Next.js | `next.config.*` |
| Nuxt.js | `nuxt.config.*` |
| Qwik | `vite.config.ts` + Qwik |
| React | `vite.config.*` + React |
| Remix | `remix.config.*` |
| SolidJS | `solid.config.*` |
| Svelte / SvelteKit | `svelte.config.*`, `*.svelte` |
| Vue.js | `vue.config.*`, `*.vue` |
| Storybook | `.storybook/main.*` |

### Backend Frameworks

| Stack | Detected Via |
|---|---|
| Django | `manage.py`, `wsgi.py` |
| Express.js | `server.js`, `app.js` |
| FastAPI | `main.py`, `app.py` |
| Flask | `wsgi.py`, `instance/` |
| Laravel | `artisan`, `composer.json` |
| NestJS | `nest-cli.json` |
| Spring Boot | `pom.xml`, `application.properties` |
| Symfony | `symfony.lock` |

### Mobile

| Stack | Detected Via |
|---|---|
| Android | `AndroidManifest.xml`, `build.gradle` |
| Capacitor (Ionic) | `capacitor.config.*` |
| Expo | `app.json`, `eas.json` |
| Flutter | `pubspec.yaml` |
| iOS / Xcode | `*.xcodeproj`, `Info.plist` |
| React Native | `metro.config.*` |
| Tauri | `src-tauri/tauri.conf.json` |

### DevOps & Cloud

| Stack | Detected Via |
|---|---|
| Ansible | `ansible.cfg`, `site.yml` |
| Apache | `.htaccess`, `httpd.conf` |
| AWS CDK | `cdk.json` |
| Docker | `Dockerfile`, `docker-compose.yml` |
| Fly.io | `fly.toml` |
| GitHub Actions | `.github/workflows/` |
| Kubernetes / Helm | `Chart.yaml`, `kustomization.yaml` |
| Netlify | `netlify.toml` |
| Nginx | `nginx.conf` |
| Packer | `*.pkr.hcl` |
| Pulumi | `Pulumi.yaml` |
| Railway | `railway.toml` |
| Terraform | `*.tf`, `*.tfvars` |
| Vagrant | `Vagrantfile` |
| Vercel | `vercel.json` |

### Databases & ORMs

| Stack | Detected Via |
|---|---|
| Drizzle ORM | `drizzle.config.*` |
| MongoDB | `mongod.conf`, `.mongorc.js` |
| MySQL | `my.cnf`, `*.sql` |
| PostgreSQL | `postgresql.conf`, `schema.sql` |
| Prisma | `schema.prisma` |
| Redis | `redis.conf`, `dump.rdb` |
| SQLite | `*.db`, `*.sqlite` |
| Supabase | `supabase/` |

### Game Dev

| Stack | Detected Via |
|---|---|
| Godot | `project.godot`, `*.gd` |
| Unity | `ProjectSettings/`, `Assets/` |
| Unreal Engine | `*.uproject` |

### Build Tools & Monorepos

| Stack | Detected Via |
|---|---|
| Bazel | `WORKSPACE`, `BUILD.bazel` |
| Gradle | `build.gradle`, `gradlew` |
| Lerna | `lerna.json` |
| Maven | `pom.xml`, `.mvn/` |
| Nx | `nx.json` |
| Turborepo | `turbo.json` |

### Runtimes & Package Managers

| Stack | Detected Via |
|---|---|
| Bun | `bun.lockb`, `bunfig.toml` |
| Conda | `environment.yml` |
| Deno | `deno.json`, `deno.lock` |
| Node.js | `package.json` |
| pnpm | `pnpm-workspace.yaml` |
| Yarn | `.yarnrc.yml`, `yarn.lock` |

### Other Tools & Frameworks

| Stack | Detected Via |
|---|---|
| Dart | `pubspec.yaml`, `*.dart` |
| Electron | `electron-builder.yml` |
| Firebase | `firebase.json` |
| GraphQL | `schema.graphql`, `codegen.yml` |
| Jupyter / Data Science | `*.ipynb`, `environment.yml` |
| Kotlin (Android) | `build.gradle`, `AndroidManifest.xml` |
| PyTorch / ML | `train.py`, `checkpoints/` |
| Shopify | `shopify.app.toml`, `theme.liquid` |
| Stripe | `stripe.config.*` |
| T3 Stack | `prisma/schema.prisma` + tRPC |
| Tamagui | `tamagui.config.*` |
| Testing (Jest/Vitest/Cypress) | `jest.config.*`, `playwright.config.*` |
| tRPC | `trpc.ts`, `router.ts` |
| Vite | `vite.config.*` |
| VS Code Extension | `.vscodeignore` |
| Wails (Go Desktop) | `wails.json` |
| WordPress | `wp-config.php` |
| LaTeX | `latexmkrc`, `*.tex` |
| PlatformIO | `platformio.ini` |
| Solidity | `foundry.toml`, `hardhat.config.*` |

### All 122 rule files

`global` is the always-applied baseline rather than a detected stack.

`android` · `angular` · `ansible` · `apache` · `astro` · `aws-cdk` · `bazel` · `bun` · `c` · `capacitor` · `clojure` · `conda` · `coq` · `cpp` · `crystal` · `csharp` · `dart` · `deno` · `django` · `dlang` · `docker` · `dotnet` · `drizzle` · `electron` · `elixir` · `erlang` · `expo` · `express` · `fastapi` · `firebase` · `flask` · `flutter` · `flyio` · `fortran` · `fsharp` · `gatsby` · `github-actions` · `gleam` · `global` · `go` · `godot` · `gradle` · `graphql` · `groovy` · `haskell` · `ios` · `java` · `julia` · `jupyter` · `kotlin` · `kotlin-jvm` · `kubernetes` · `laravel` · `latex` · `lean` · `lerna` · `lua` · `matlab` · `maven` · `mongodb` · `mysql` · `nestjs` · `netlify` · `nextjs` · `nginx` · `nim` · `nix` · `node` · `nuxt` · `nx` · `objectivec` · `ocaml` · `packer` · `perl` · `php` · `platformio` · `pnpm` · `postgres` · `prisma` · `pulumi` · `purescript` · `python` · `pytorch` · `qwik` · `r` · `railway` · `react` · `react-native` · `redis` · `remix` · `ruby` · `rust` · `scala` · `shopify` · `solid` · `solidity` · `spring` · `sqlite` · `storybook` · `stripe` · `supabase` · `svelte` · `swift` · `symfony` · `t3` · `tamagui` · `tauri` · `terraform` · `testing` · `trpc` · `turborepo` · `unity` · `unreal` · `vagrant` · `vercel` · `vite` · `vscode-extension` · `vue` · `wails` · `wordpress` · `yarn` · `zig`

> Don't see your stack? SIIG will attempt to fetch it automatically from [siig-rules](https://github.com/yv3000/siig-rules). You can also [contribute rules](https://github.com/yv3000/siig-rules/blob/main/CONTRIBUTING.md) for any stack not listed here.

---

## How It Works

### Architecture

```
siig (single binary, ~5.5MB)
├── rules/db/         ← 122 JSON rule files, embedded at compile time
├── scanner/          ← walks project dir (max depth 3), collects file/dir names
├── matcher/          ← merges rules per target; fetches unknown stacks online
├── fetcher/          ← fetches a single stack JSON from siig-rules on demand
├── cache/            ← stores fetched rules in ~/.siig/cache/ for 7 days
├── writer/           ← dedupes against existing ignore files, appends new entries
├── output/           ← CLI reporting (results, dry-run, verbose)
├── internal/version/ ← version stamped at release time
└── cmd/              ← cobra entry point
```

### The Flow

```
User runs: siig
      |
      v
  scanner   →  walks project root (max depth 3, skips node_modules/.git)
      |            collects all file names and dir names found
      v
  matcher   →  for each of 122 embedded stack rules:
      |            checks if detect.files or detect.dirs exist in scan result
      |            if match → apply that stack's rules
      |            global rules always applied regardless of stack
      |
      |         for each unrecognized file in project root:
      |            check ~/.siig/cache/<name>.json  (cache hit → apply instantly)
      |            if cache miss → fetch siig-rules/db/<name>.json  (~100-300ms)
      |            if found → cache it → apply rules → show as "Stack (online)"
      |            if 404 → silently skip
      v
  writer    →  reads existing ignore files
      |            builds a Set of already-present entries
      |            appends only entries not already there
      |            adds "# Added by SIIG vX.X.X" comment block
      v
  output    →  prints summary of what was detected and written
```

### Rules DB

Each stack is a single JSON file in `rules/db/`. Example — `node.json`:

```json
{
  "id": "node",
  "name": "Node.js",
  "detect": {
    "files": ["package.json", "package-lock.json", "yarn.lock"],
    "dirs": ["node_modules"]
  },
  "rules": {
    "gitignore": ["node_modules/", ".env", ".env.local", "dist/", "*.log"],
    "vercelignore": ["node_modules/", ".env"],
    "dockerignore": ["node_modules/", ".env", "*.log"],
    "vscode": { "files.exclude": { "node_modules": true } }
  }
}
```

All 122 JSON files are embedded into the binary at compile time using Go's `//go:embed` — **no external files needed at runtime for any of the 122 supported stacks.**

### Cache

Fetched rules are stored in `~/.siig/cache/` and reused for 7 days before re-fetching. Use `--no-cache` to force a fresh fetch, or `--offline` to skip all network calls entirely.

---

## Security

- **Hybrid offline-first** — 122 stacks work with zero network. Network is only used as a fallback for stacks not in the embedded DB, and only for reading public JSON files from GitHub.
- **Fetched rules are validated** — the response is capped at 1 MiB, parsed with unknown JSON fields rejected, and then checked: the rule must answer with the stack id that was requested, and no entry may be empty, oversized or contain a newline (which would let a remote file inject arbitrary lines into your ignore files). The same check applies when a rule is read back from the local cache.
- **No telemetry** — no data collection of any kind.
- **Read-only scan** — SIIG reads file and directory *names*. The only content it opens is root `package.json` (to see if `main`/`exports` mean a library) and existing ignore / `.vscode/settings.json` files it may update. Source, secrets, and `.env` values are never read.
- **Write-safe** — only appends to ignore files. Never modifies source code, never deletes anything.
- **Idempotent** — every entry is checked against existing content before writing. Running twice is identical to running once.
- **Checksum verified install** — install scripts verify SHA256 of the downloaded binary before placing it on your system.
- **No elevated permissions** — runs entirely as the current user. No admin/sudo required after install.

---

## Platform Support

| Platform | Architecture | Status |
|---|---|---|
| Windows 10/11 | x64 (amd64) | ✅ Fully supported |
| Windows 11 | ARM64 | ✅ Fully supported |
| macOS (Intel) | x64 (amd64) | ✅ Fully supported |
| macOS (Apple Silicon) | ARM64 (arm64) | ✅ Fully supported |
| Linux | x64 (amd64) | ✅ Fully supported |
| Linux | ARM64 | ✅ Fully supported |

---

## Build from Source

```bash
git clone https://github.com/yv3000/SIIG.git
cd SIIG
go test ./...
go vet ./...
go build -o siig .
```

Tests live next to each package (`matcher/matcher_test.go`, `writer/writer_test.go`, `scanner/scanner_test.go`, `cache/cache_test.go`) and use temp directories, so they never touch your real home or projects.

CI runs on every push and pull request:

| Job | What it enforces |
|---|---|
| `test` | `go test ./... -covermode=atomic` and `go vet ./...` on Go 1.21 **and** current stable; total coverage must stay at or above **80%** |
| `lint` | golangci-lint v2 (standard linters + `gofmt`) |
| `audit` | `govulncheck ./...` for called vulnerabilities |

Reproduce the same checks locally:

```bash
go test ./... -covermode=atomic -coverprofile=coverage.out
go tool cover -func=coverage.out | tail -1   # total must be >= 80%
go vet ./...
golangci-lint run ./...
```

Cross-compile for any platform:

```bash
GOOS=windows GOARCH=amd64 go build -o siig-windows-amd64.exe .
GOOS=darwin  GOARCH=arm64 go build -o siig-darwin-arm64 .
GOOS=linux   GOARCH=amd64 go build -o siig-linux-amd64 .
```

**Requirements:** Go 1.21+

---

## FAQ

**Will SIIG overwrite my existing `.gitignore`?**
No. SIIG only appends entries that aren't already present. Your existing content is never modified or removed.

**Is it safe to run on an existing project?**
Yes. SIIG never touches source code — only ignore files. Run `siig --dry-run` first if you want to preview exactly what would be added.

**What if my stack isn't in the embedded 122?**
SIIG automatically tries to fetch it from [siig-rules](https://github.com/yv3000/siig-rules) on demand. If found, it's cached locally for 7 days and applied — no action needed from you. If it doesn't exist there either, it's silently skipped. The request is built from the unrecognised file's name, so a project with `gleam.toml` asks for the `gleam` rule; a stack whose marker file is named differently (PureScript's `spago.dhall`) needs the rule to be embedded, which happens through the sync PR and the next release. You can also [contribute rules](https://github.com/yv3000/siig-rules/blob/main/CONTRIBUTING.md) for it.

**Does SIIG read my `.env` files or source code?**
No. Scanning uses file and directory names. It may read root `package.json` for library detection and the ignore / VS Code settings files it writes. `.env` and source are never read.

**How do I use SIIG without internet?**
Run `siig --offline`. This skips all network requests and uses only the 122 embedded stacks plus anything already in the local cache.

**How do I force a fresh fetch and ignore the cache?**
Run `siig --no-cache`. Cached rules are skipped for that run and every unknown stack is fetched fresh — nothing on disk is deleted. To actually remove the cached rules, run `siig --clear-cache`.

**Can I run SIIG in CI?**
Yes — `siig` exits with code 0 on success, non-zero on error. Use `siig --check` to fail the job when the ignore files are missing entries (it writes nothing), and `siig --offline` if you don't want any outbound network calls.

Add `--json` to get one machine-readable object on stdout instead of the summary, so a job can post or archive the result without scraping text:

```json
{
  "version": "2.0.2",
  "stacks": ["Docker", "Go"],
  "onlineStacks": [],
  "targets": [
    {"target": "gitignore", "file": ".gitignore", "entriesAdded": 23, "skipped": false, "reason": ""},
    {"target": "npmignore", "file": ".npmignore", "entriesAdded": 0, "skipped": true, "reason": "not applicable for this project"}
  ],
  "scanErrors": []
}
```

`file` is always relative to the project — SIIG never puts your directory layout into a CI log. `--json` writes what it reports unless you also pass `--dry-run`, and `--json --check` still exits 1 while entries are missing (the error goes to stderr, the JSON to stdout).

**How do I add a custom stack?**
Open a PR to [siig-rules](https://github.com/yv3000/siig-rules) adding a JSON file to `db/` following the schema in [CONTRIBUTING.md](https://github.com/yv3000/siig-rules/blob/main/CONTRIBUTING.md).

---

## License

MIT License

Copyright (c) 2026 THE YV

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

---

<p align="center">
  <sub>YV 🖤 ~ I EXPECT NOTHING FROM YOU...</sub>
</p>