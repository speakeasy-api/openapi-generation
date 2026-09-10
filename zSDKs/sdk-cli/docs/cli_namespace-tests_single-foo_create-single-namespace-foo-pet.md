## cli namespace-tests single-foo create-single-namespace-foo-pet

Create Single Namespace Foo Pet

### Synopsis

This endpoint tests creating a component in the foo namespace.
No import aliasing should be needed since there's no conflict within this group.

```
cli namespace-tests single-foo create-single-namespace-foo-pet [flags]
```

### Examples

```
  cli single-foo create-single-namespace-foo-pet --id pet-foo-123 --name Fluffy --species cat
```

### Options

```
      --body string         Request body as JSON (alternative to individual flags). Can also be provided via stdin; @path reads a file, @- reads stdin to EOF. Use --schema to print the exact JSON Schema.
  -e, --errors string       A field name that often collides
  -h, --help                help for create-single-namespace-foo-pet
  -i, --id string           [required]
  -m, --models string       A field name that often collides
  -n, --name string         [required]
      --operations string   A field name that often collides
  -r, --request string      A field name that often collides
      --schema              Print the exact JSON Schema of the request body and exit
  -s, --species string      The species of the pet (e.g., dog, cat) [required]
  -u, --utils string        A field name that often collides
```

### Options inherited from parent commands

```
      --agent-mode                       Enable structured errors and default TOON output for AI coding agents.
      --app-id string                    Custom authentication credential
      --bearer-auth string               HTTP Bearer
      --client-id string                 Client Credentials flow. client identifier
      --client-secret string             Client Credentials flow. client secret
      --color string                     Control colored output: auto (color when output is a TTY), always, or never. Respects NO_COLOR and FORCE_COLOR env vars. (default "auto")
  -d, --debug                            Log request and response diagnostics to stderr
      --deprecated-query-param1 string   A deprecated description (env: CLI_DEPRECATED_QUERY_PARAM1)
      --deprecated-query-param2 string   Global deprecated-query-param2 parameter (env: CLI_DEPRECATED_QUERY_PARAM2)
      --dry-run                          Preview API requests without sending them (no network, no OS keychain). Human preview on stderr; with -o json or --jq, one JSON object per request on stdout. Local mutation commands (auth login, auth logout and configure) make no request: they skip prompts and writes and report a no-op (stderr, or one JSON object on stdout in the machine form)
  -H, --header stringArray               Set a custom HTTP request header (format: "Key: Value"). Can be specified multiple times.
      --host-name string                 Server template variable: HostName
      --include-headers                  Include HTTP response headers in the output
      --interactive                      Prompt for missing inputs and open guided configure/auth forms (forms fall back to line prompts on stdin off-TTY)
  -q, --jq string                        Filter and transform output using a jq expression (e.g., '.name', '.items[] | .id')
      --lone-query-param string          Global lone-query-param parameter (env: CLI_LONE_QUERY_PARAM)
      --mobile-auth string               OAuth2 Password flow, a flow that is too
                                         long to describe in a single line.
      --my-api-key string                API Key
      --no-interactive                   Disable all interactive features (auto-prompting, explorer auto-launch, TUI forms)
      --no-retries                       Disable automatic retries (default: retries enabled with exponential backoff)
      --oauth2 string                    OAuth2 Authorization
  -o, --output-format string             Specify the output format. Options: pretty, json, yaml, table, toon. (default "pretty")
      --password string                  HTTP Basic password
      --port string                      Server template variable: PORT
      --query-param1 string              A long winded, multi-line description (env: CLI_QUERY_PARAM1)
      --raw-output                       Write --jq string results as raw text instead of JSON strings (like jq -r); non-string results stay JSON
      --retry-config string              Full retry config as JSON. Schema: {"strategy":"backoff","backoff":{"initialInterval":500,"maxInterval":10000,"exponent":1.5,"maxElapsedTime":30000},"retryConnectionErrors":false}. Times are in milliseconds.
      --retry-connection-errors          Retry on connection errors (EOF, reset, etc.)
      --retry-max-elapsed-time string    Maximum total time for retries (e.g., 30s, 5m). Default: 30s
      --secret string                    Custom authentication credential
      --server string                    Select a server by index (for indexed servers) or name (for named servers)
      --server-url string                Override the default server URL
      --subdomain string                 Server template variable: subdomain
      --timeout string                   HTTP request timeout (e.g., 30s, 5m, 100ms)
      --token-url string                 Client Credentials flow. token URL
      --usage                            Print the CLI Usage schema in KDL format
      --username string                  HTTP Basic username
      --version string                   Server template variable: version
```

### SEE ALSO

* [cli namespace-tests single-foo](cli_namespace-tests_single-foo.md)	 - Operations for single-foo

### Machine interface

* `cli namespace-tests single-foo create-single-namespace-foo-pet --usage` — this command's flags, defaults and env vars as machine-readable KDL
* `cli namespace-tests single-foo create-single-namespace-foo-pet --schema` — the exact JSON Schema of the request body (all `$ref`s bundled)
* `cli namespace-tests single-foo create-single-namespace-foo-pet --dry-run` — preview the request without OS-keychain access or a network call (human preview on stderr)
* `--dry-run --output-format json` (or a caller-explicit `--jq`) writes one preview object per request as NDJSON on stdout; jq is not applied to previews
* `--output-format json` or `--jq <expr>` for machine-readable live output; in agent mode errors are a JSON envelope on stderr

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
