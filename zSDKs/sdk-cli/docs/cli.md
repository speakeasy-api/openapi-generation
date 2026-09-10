## cli

SDK Review: A test document for reviewing the SDK.

### Synopsis

This document will show case as many of our features as possible in as little operations/models as possible.
This will then generate a SDK that we can more easily review than the test SDKs based on uber.yaml spec.

```
cli [flags]
```

### Options

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
  -h, --help                             help for cli
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

* [cli archive](cli_archive.md)	 - Archive a user account
* [cli auth](cli_auth.md)	 - Manage authentication credentials
* [cli binary-and-string-upload](cli_binary-and-string-upload.md)	 - Binary And String Upload
* [cli chat](cli_chat.md)	 - Chat
* [cli configure](cli_configure.md)	 - Configure authentication, global parameters, and preferences
* [cli create-user](cli_create-user.md)	 - Create User
* [cli create-with-union](cli_create-with-union.md)	 - Create with discriminated union request body
* [cli delete-user](cli_delete-user.md)	 - Delete User
* [cli enroll](cli_enroll.md)	 - Enroll a user by email
* [cli explore](cli_explore.md)	 - Interactively browse and run commands
* [cli get-asset](cli_get-asset.md)	 - Get Asset
* [cli get-binary-default-response](cli_get-binary-default-response.md)	 - Get Binary Default Response
* [cli get-duplicate-export-collision](cli_get-duplicate-export-collision.md)	 - Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308
* [cli get-empty-object-error](cli_get-empty-object-error.md)	 - Get Empty Object Error
* [cli get-error-in-union](cli_get-error-in-union.md)	 - Get Error In Union
* [cli get-error-only-example](cli_get-error-only-example.md)	 - Operation with example only on error response
* [cli get-fully-flattened-request](cli_get-fully-flattened-request.md)	 - Get Fully Flattened Request
* [cli get-named-primitive-union](cli_get-named-primitive-union.md)	 - Test named primitive union options using title and x-speakeasy-name-override
* [cli get-nested-integer-string](cli_get-nested-integer-string.md)	 - Test nested struct with integer:string tag
* [cli get-polymorphism](cli_get-polymorphism.md)	 - Get Polymorphism
* [cli get-request-body-flattened-away](cli_get-request-body-flattened-away.md)	 - Get Request Body Flattened Away
* [cli get-union-errors](cli_get-union-errors.md)	 - Get Union Errors
* [cli get-user](cli_get-user.md)	 - Get User
* [cli group](cli_group.md)	 - Operations for group
* [cli invite](cli_invite.md)	 - Invite a user by email
* [cli login](cli_login.md)	 - Login
* [cli namespace-tests](cli_namespace-tests.md)	 - Operations for namespace-tests
* [cli observe](cli_observe.md)	 - Produce an asset and print its terminal status
* [cli obsolete](cli_obsolete.md)	 - A subSDK in which all operations are deprecated
* [cli operation-with-leading-and-trailing-underscores](cli_operation-with-leading-and-trailing-underscores.md)	 - Operation With Leading And Trailing Underscores
* [cli parentheses-in-path-allowed](cli_parentheses-in-path-allowed.md)	 - A string with {{ double braces }} and { single braces }
and \{\{ escaped curlies \}\} and `backticks`.
and \`escaped backticks\` and double slashes\\
and 'single quotes' and "double quotes".
and  \'escaped single quotes\' and \"escaped double quotes\".
* [cli post-file](cli_post-file.md)	 - Post File
* [cli produce](cli_produce.md)	 - Produce an image asset and wait for completion
* [cli render](cli_render.md)	 - Render an image asset to a file
* [cli render-asset](cli_render-asset.md)	 - Render Asset
* [cli say](cli_say.md)	 - Stream a chat reply
* [cli tag1](cli_tag1.md)	 - The first tag
* [cli test-endpoint](cli_test-endpoint.md)	 - Test Endpoint
* [cli test-enum-formats](cli_test-enum-formats.md)	 - Test x-speakeasy-enums in different formats
* [cli test-group](cli_test-group.md)	 - Operations for test-group
* [cli update-user](cli_update-user.md)	 - Update User
* [cli url-validation-stress-test](cli_url-validation-stress-test.md)	 - Url Validation Stress Test
* [cli validate](cli_validate.md)	 - Validate
* [cli version](cli_version.md)	 - Print the CLI version
* [cli whoami](cli_whoami.md)	 - Display current authentication and global parameter configuration

Exit codes: 0 ok · 1 runtime · 2 usage · 3 authentication/authorization
