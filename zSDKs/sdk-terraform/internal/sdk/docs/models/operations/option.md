# Options

## Global Options

Global options are passed when initializing the SDK client and apply to all operations.

### WithServerURL

WithServerURL allows providing an alternative server URL.

```go
sdk.WithServerURL("https://api.example.com")
```

### WithTemplatedServerURL

WithTemplatedServerURL allows providing an alternative server URL with templated parameters.

```go
sdk.WithTemplatedServerURL("https://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})
```

### WithServerIndex

WithServerIndex allows the overriding of the default server by index.

```go
sdk.WithServerIndex(1)
```

### WithSubdomain

WithSubdomain allows setting the subdomain variable for url substitution.

```go
sdk.WithSubdomain(/* ... */)
```

### WithVersion

WithVersion allows setting the version variable for url substitution.

```go
sdk.WithVersion(/* ... */)
```

### WithHostName

WithHostName allows setting the HostName variable for url substitution.

```go
sdk.WithHostName(/* ... */)
```

### WithPort

WithPort allows setting the PORT variable for url substitution.

```go
sdk.WithPort(/* ... */)
```

### WithClient

WithClient allows the overriding of the default HTTP client used by the SDK.

```go
sdk.WithClient(httpClient)
```

### WithSecurity

WithSecurity configures the SDK to use the provided security details.

```go
sdk.WithSecurity(/* ... */)
```

### WithSecuritySource

WithSecuritySource configures the SDK to invoke the provided function on each method call to determine authentication.

```go
sdk.WithSecuritySource(/* ... */)
```

### WithQueryParam1

WithQueryParam1 allows setting the QueryParam1 parameter for all supported operations.

```go
sdk.WithQueryParam1(/* ... */)
```

### WithDeprecatedQueryParam1

WithDeprecatedQueryParam1 allows setting the DeprecatedQueryParam1 parameter for all supported operations.

```go
sdk.WithDeprecatedQueryParam1(/* ... */)
```

### WithDeprecatedQueryParam2

WithDeprecatedQueryParam2 allows setting the DeprecatedQueryParam2 parameter for all supported operations.

```go
sdk.WithDeprecatedQueryParam2(/* ... */)
```

### WithLoneQueryParam

WithLoneQueryParam allows setting the LoneQueryParam parameter for all supported operations.

```go
sdk.WithLoneQueryParam(/* ... */)
```

### WithRetryConfig

WithRetryConfig allows setting the default retry configuration used by the SDK for all supported operations.

```go
sdk.WithRetryConfig(retry.Config{
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
sdk.WithTimeout(30 * time.Second)
```

## Per-Method Options

Per-method options are passed as the last argument to individual methods and override any global settings for that request.

### WithServerURL

WithServerURL allows providing an alternative server URL for a single request.

```go
operations.WithServerURL("http://api.example.com")
```

### WithTemplatedServerURL

WithTemplatedServerURL allows providing an alternative server URL with templated parameters for a single request.

```go
operations.WithTemplatedServerURL("http://{host}:{port}", map[string]string{
    "host": "api.example.com",
    "port": "8080",
})
```

### WithRetries

WithRetries allows customizing the default retry configuration for a single request.

```go
operations.WithRetries(retry.Config{
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
operations.WithOperationTimeout(30 * time.Second)
```

### WithSetHeaders

WithSetHeaders allows setting custom headers on a per-request basis. If the request already contains headers matching the provided keys, they will be overwritten.

```go
operations.WithSetHeaders(map[string]string{
    "X-Cache-TTL": "60",
})
```

### WithURLOverride

WithURLOverride allows overriding the default URL for an operation.

```go
operations.WithURLOverride("/custom/path")
```

### WithAcceptHeaderOverride

WithAcceptHeaderOverride allows overriding the `Accept` header for operations that support multiple response content types.

```go
operations.WithAcceptHeaderOverride(operations.AcceptHeaderEnumWildcardRootWildcard)
```