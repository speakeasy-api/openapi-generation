/** Release file generation for CLI binary distribution.
 *
 * Generates:
 * - .goreleaser.yaml (via writeFile to avoid Go template escaping conflicts)
 * - .github/workflows/release.yaml (via .stmpl template)
 * - scripts/install.sh (via .stmpl template)
 * - scripts/install.ps1 (via .stmpl template)
 */

/** Returns true if release files should be generated. */
function shouldGenerateReleaseFiles(): boolean {
  return context.Global.Config.GenerateRelease !== false;
}

function getDistributionConfig(): Record<string, any> {
  return context.Global.Config.Distribution || {};
}

function getHomebrewConfig(): Record<string, any> {
  return getDistributionConfig().homebrew || {};
}

function getWingetConfig(): Record<string, any> {
  return getDistributionConfig().winget || {};
}

function getNFPMConfig(): Record<string, any> {
  return getDistributionConfig().nfpm || {};
}

/** Returns owner/repo for the published GitHub repository, derived from repoURL
 * (preferred) or packageName. Returns empty strings if unavailable. */
function getGitHubRepoParts(): { owner: string; repo: string } {
  const repoURL = String(context.Global.Config.RepoURL || "").trim();
  const packageName = String(context.Global.Config.PackageName || "").trim();

  const source = repoURL || packageName;
  const match = source.match(/github\.com\/([^/]+)\/([^/]+)/);
  if (match) {
    return { owner: match[1], repo: match[2].replace(/\.git$/, "") };
  }

  return { owner: "", repo: "" };
}

function getGitHubRepo(): string {
  const { owner, repo } = getGitHubRepoParts();
  if (owner && repo) {
    return `${owner}/${repo}`;
  }
  return "";
}

function getGitHubRepoURL(): string {
  const repo = getGitHubRepo();
  if (repo) {
    return `https://github.com/${repo}`;
  }
  return "";
}

function getReleaseJobs(): Job[] {
  return [
    {
      ID: "goreleaser-config",
      FileName: ".goreleaser.yaml",
      Context: null,
    },
    createTemplateFileJob(
      "release/release-workflow.yaml.stmpl",
      ".github/workflows/release.yaml",
      {},
    ),
    createTemplateFileJob("release/install.sh.stmpl", "scripts/install.sh", {}),
    createTemplateFileJob(
      "release/install.ps1.stmpl",
      "scripts/install.ps1",
      {},
    ),
  ];
}

function bashEnvDefault(name: string, defaultValue: string = ""): string {
  return `\${${envVarName(name)}:-${defaultValue}}`;
}

function bashEnvRef(name: string): string {
  return `\${${envVarName(name)}}`;
}

registerTemplateFunc("shouldGenerateReleaseFiles", shouldGenerateReleaseFiles);
registerTemplateFunc("getGitHubRepo", getGitHubRepo);
registerTemplateFunc("getGitHubRepoURL", getGitHubRepoURL);
registerTemplateFunc("bashEnvDefault", bashEnvDefault);
registerTemplateFunc("bashEnvRef", bashEnvRef);
registerTemplateFunc("envVarName", envVarName);
registerTemplateFunc("isHomebrewEnabled", isHomebrewEnabled);
registerTemplateFunc("getHomebrewTap", getHomebrewTap);
registerTemplateFunc("getHomebrewInstallCommand", getHomebrewInstallCommand);
registerTemplateFunc("isWingetEnabled", isWingetEnabled);
registerTemplateFunc("getWingetInstallCommand", getWingetInstallCommand);
registerTemplateFunc("isNFPMEnabled", isNFPMEnabled);
registerTemplateFunc("hasNFPMFormat", hasNFPMFormat);

function envVarName(key: string): string {
  const prefix = String(context.Global.Config.EnvVarPrefix || "")
    .replace(/[^A-Za-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .toUpperCase();
  const suffix = String(key || "")
    .replace(/[^A-Za-z0-9]+/g, "_")
    .replace(/^_+|_+$/g, "")
    .toUpperCase();

  if (!prefix) {
    return suffix;
  }
  if (!suffix) {
    return prefix;
  }

  return `${prefix}_${suffix}`;
}

function getNFPMFormats(): string[] {
  const nfpm = getNFPMConfig();
  const raw = String(nfpm.formats || "deb,rpm").trim();
  if (!raw) {
    return ["deb", "rpm"];
  }
  return raw
    .split(",")
    .map((f) => f.trim())
    .filter((f) => f.length > 0);
}

function isHomebrewEnabled(): boolean {
  return getHomebrewConfig().enabled === true;
}

function getHomebrewTap(): string {
  return String(getHomebrewConfig().tap || "").trim();
}

function yamlDoubleQuoted(value: string): string {
  const escaped = String(value || "")
    .replace(/\\/g, "\\\\")
    .replace(/\r/g, "\\r")
    .replace(/\n/g, "\\n")
    .replace(/\t/g, "\\t")
    .replace(/"/g, '\\"');
  return `"${escaped}"`;
}

function getHomebrewInstallCommand(): string {
  const tap = getHomebrewTap();
  const parts = tap.split("/");
  if (parts.length !== 2) {
    return `brew install ${sanitizeCliName()}`;
  }

  const tapName = parts[1].replace(/^homebrew-/, "");
  return `brew install ${parts[0]}/${tapName}/${sanitizeCliName()}`;
}

function isWingetEnabled(): boolean {
  return getWingetConfig().enabled === true;
}

function getWingetInstallCommand(): string {
  const winget = getWingetConfig();
  const packageIdentifier = String(winget.packageIdentifier || "").trim();
  if (packageIdentifier) {
    return `winget install ${packageIdentifier}`;
  }

  return `winget install ${sanitizeCliName()}`;
}

function isNFPMEnabled(): boolean {
  return getNFPMConfig().enabled === true;
}

function hasNFPMFormat(format: string): boolean {
  return getNFPMFormats().includes(format);
}

function getReleaseSection(repo: { owner: string; repo: string }): string {
  const lines = [`release:`];

  if (repo.owner && repo.repo) {
    lines.push(
      `  github:`,
      `    owner: ${repo.owner}`,
      `    name: ${repo.repo}`,
    );
  }

  lines.push(`  draft: false`, `  prerelease: auto`, `  mode: append`);

  return `\n${lines.join("\n")}\n`;
}

function getHomebrewSection(): string {
  if (!isHomebrewEnabled()) {
    return "";
  }

  const description = templateRootLong().replace(/^"|"$/g, "");
  const tap = getHomebrewTap();
  const [owner, name] = tap.split("/");

  return `
brews:
  - name: ${sanitizeCliName()}
    repository:
      owner: ${owner}
      name: ${name}
      branch: main
    token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    homepage: ${yamlDoubleQuoted(getGitHubRepoURL())}
    description: ${yamlDoubleQuoted(description)}
    install: |
      bin.install "${sanitizeCliName()}"
`;
}

function getWingetSection(): string {
  if (!isWingetEnabled()) {
    return "";
  }

  const winget = getWingetConfig();
  const description = templateRootLong().replace(/^"|"$/g, "");
  const publisher = String(winget.publisher || "").trim();
  const publisherURL = String(winget.publisherUrl || "").trim();
  const packageIdentifier = String(winget.packageIdentifier || "").trim();
  const repositoryOwner = String(winget.repositoryOwner || "").trim();

  return `
winget:
  - name: ${yamlDoubleQuoted(sanitizeCliName())}
    publisher: ${yamlDoubleQuoted(publisher)}
    publisher_url: ${yamlDoubleQuoted(publisherURL)}
    package_identifier: ${yamlDoubleQuoted(packageIdentifier)}
    license: ${yamlDoubleQuoted(String(winget.license || "").trim())}
    homepage: ${yamlDoubleQuoted(getGitHubRepoURL())}
    short_description: ${yamlDoubleQuoted(description)}
    repository:
      owner: ${repositoryOwner}
      name: winget-pkgs
      branch: "${sanitizeCliName()}-{{ .Version }}"
      token: "{{ .Env.WINGET_GITHUB_TOKEN }}"
    pull_request:
      enabled: true
      draft: true
      base:
        owner: microsoft
        name: winget-pkgs
        branch: master
`;
}

function getNFPMSection(): string {
  if (!isNFPMEnabled()) {
    return "";
  }

  const nfpm = getNFPMConfig();
  const homepage = getGitHubRepoURL();
  const description = templateRootLong().replace(/^"|"$/g, "");
  const formats = getNFPMFormats()
    .map((format) => `      - ${format}`)
    .join("\n");

  return `
nfpms:
  - package_name: ${sanitizeCliName()}
    file_name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Arch }}"
    homepage: ${yamlDoubleQuoted(homepage)}
    description: ${yamlDoubleQuoted(description)}
    maintainer: ${yamlDoubleQuoted(String(nfpm.maintainer || "").trim())}
    license: ${yamlDoubleQuoted(String(nfpm.license || "").trim())}
    formats:
${formats}
`;
}

function shouldSignChecksums(): boolean {
  return isHomebrewEnabled() || isWingetEnabled() || isNFPMEnabled();
}

function getSigningSection(): string {
  if (!shouldSignChecksums()) {
    return "";
  }

  return `
signs:
  - artifacts: checksum
    args:
      - "--batch"
      - "--local-user"
      - "{{ .Env.GPG_FINGERPRINT }}"
      - "--output"
      - "\${signature}"
      - "--detach-sign"
      - "\${artifact}"
`;
}

/** Writes the .goreleaser.yaml configuration file. */
function writeGoreleaserConfig() {
  const cliName = sanitizeCliName();
  const repo = getGitHubRepoParts();

  // Build the goreleaser name_template using goreleaser's own template syntax.
  // These are NOT Go text/template vars — they're goreleaser runtime vars that
  // must be written literally to the output file.
  const nameTemplate = [
    `{{ .ProjectName }}_`,
    `{{- title .Os }}_`,
    `{{- if eq .Arch "amd64" }}x86_64`,
    `{{- else if eq .Arch "386" }}i386`,
    `{{- else }}{{ .Arch }}{{ end }}`,
    `{{- if .Arm }}v{{ .Arm }}{{ end }}`,
  ].join("\n      ");

  const config = `# yaml-language-server: $schema=https://goreleaser.com/static/schema.json
version: 2

before:
  hooks:
    - go mod tidy

builds:
  - id: ${cliName}
    main: ./cmd/${cliName}
    binary: ${cliName}
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - windows
      - darwin
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.buildTime={{.Date}}

archives:
  - id: ${cliName}
    formats: [tar.gz]
    name_template: >-
      ${nameTemplate}
    format_overrides:
      - goos: windows
        formats: [zip]
    files:
      - README.md
      - LICENSE*
${getHomebrewSection()}${getWingetSection()}${getNFPMSection()}${getReleaseSection(
    repo,
  )}
checksum:
  name_template: "checksums.txt"
${getSigningSection()}`;

  writeFile(".goreleaser.yaml", config, 0, false);
}
