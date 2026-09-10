# Options

## Global Options

Global options are passed when initializing the SDK client and apply to all operations.

### WithServerURL

WithServerURL allows providing an alternative server URL.

```go
examplealias.WithServerURL("https://api.example.com")
```

### WithTemplatedServerURL

WithTemplatedServerURL allows providing an alternative server URL with templated parameters.

```go
examplealias.WithTemplatedServerURL("https://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})
```

### WithServerIndex

WithServerIndex allows the overriding of the default server by index.

```go
examplealias.WithServerIndex(1)
```

### WithSubdomain

WithSubdomain allows setting the subdomain variable for url substitution.

```go
examplealias.WithSubdomain(/* ... */)
```

### WithVersion

WithVersion allows setting the version variable for url substitution.

```go
examplealias.WithVersion(/* ... */)
```

### WithHostName

WithHostName allows setting the HostName variable for url substitution.

```go
examplealias.WithHostName(/* ... */)
```

### WithPort

WithPort allows setting the PORT variable for url substitution.

```go
examplealias.WithPort(/* ... */)
```

### WithClient

WithClient allows the overriding of the default HTTP client used by the SDK.

```go
examplealias.WithClient(httpClient)
```

### WithSecurity

WithSecurity configures the SDK to use the provided security details.

```go
examplealias.WithSecurity(/* ... */)
```

### WithSecuritySource

WithSecuritySource configures the SDK to invoke the provided function on each method call to determine authentication.

```go
examplealias.WithSecuritySource(/* ... */)
```

### WithQueryParam1

WithQueryParam1 allows setting the QueryParam1 parameter for all supported operations.

```go
examplealias.WithQueryParam1(/* ... */)
```

### WithDeprecatedQueryParam1

WithDeprecatedQueryParam1 allows setting the DeprecatedQueryParam1 parameter for all supported operations.

```go
examplealias.WithDeprecatedQueryParam1(/* ... */)
```

### WithDeprecatedQueryParam2

WithDeprecatedQueryParam2 allows setting the DeprecatedQueryParam2 parameter for all supported operations.

```go
examplealias.WithDeprecatedQueryParam2(/* ... */)
```

### WithLoneQueryParam

WithLoneQueryParam allows setting the LoneQueryParam parameter for all supported operations.

```go
examplealias.WithLoneQueryParam(/* ... */)
```

### WithRetryConfig

WithRetryConfig allows setting the default retry configuration used by the SDK for all supported operations.

```go
examplealias.WithRetryConfig(retry.Config{
    Strategy: "backoff",
    Backoff: retry.BackoffStrategy{
        InitialInterval: 500 * time.Millisecond,
        MaxInterval: 60 * time.Second,
        Exponent: 1.5,
        MaxElapsedTime: 5 * time.Minute,
    },
    RetryConnectionErrors: true,
})
```

### WithTimeout

WithTimeout sets the default request timeout for all operations.

```go
examplealias.WithTimeout(30 * time.Second)
```

## Per-Method Options

Per-method options are passed as the last argument to individual methods and override any global settings for that request.

### WithMethodServerURL

WithMethodServerURL allows providing an alternative server URL for a single request.

```go
examplealias.WithMethodServerURL("http://api.example.com")
```

### WithMethodTemplatedServerURL

WithMethodTemplatedServerURL allows providing an alternative server URL with templated parameters for a single request.

```go
examplealias.WithMethodTemplatedServerURL("http://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})
```

### WithRetries

WithRetries allows customizing the default retry configuration for a single request.

```go
examplealias.WithRetries(retry.Config{
    Strategy: "backoff",
    Backoff: retry.BackoffStrategy{
        InitialInterval: 500 * time.Millisecond,
        MaxInterval: 60 * time.Second,
        Exponent: 1.5,
        MaxElapsedTime: 5 * time.Minute,
    },
    RetryConnectionErrors: true,
})
```

### WithOperationTimeout

WithOperationTimeout allows setting the request timeout for a single request.

```go
examplealias.WithOperationTimeout(30 * time.Second)
```

### WithSetHeaders

WithSetHeaders allows setting custom headers on a per-request basis. If the request already contains headers matching the provided keys, they will be overwritten.

```go
examplealias.WithSetHeaders(map[string]string{
    "X-Cache-TTL": "60",
})
```

### WithURLOverride

WithURLOverride allows overriding the default URL for an operation.

```go
examplealias.WithURLOverride("/custom/path")
```

### WithAcceptHeaderOverride

WithAcceptHeaderOverride allows overriding the `Accept` header for operations that support multiple response content types.

```go
examplealias.WithAcceptHeaderOverride(examplealias.AcceptHeaderEnumWildcardRootWildcard)
```