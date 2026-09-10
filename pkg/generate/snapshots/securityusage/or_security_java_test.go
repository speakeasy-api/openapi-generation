package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestOrSecurity_Java(t *testing.T) {
	t.Parallel()

	spec := specOrSecurity

	genYaml := `java:
  groupId: com.example
  artifactId: multi-auth-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [auth1Hoisted](#auth1hoisted)
* [auth2Hoisted](#auth2hoisted)
* [auth2Preferred](#auth2preferred)
* [opLevelClientCredentials](#oplevelclientcredentials)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="java" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.GlobalSecurityResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth1(System.getenv().getOrDefault("AUTH1", ""))
                    .build())
            .build();

        GlobalSecurityResponse res = sdk.globalSecurity()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[GlobalSecurityResponse](../../models/operations/GlobalSecurityResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## auth1Hoisted

### Example Usage

<!-- UsageSnippet language="java" operationID="auth1Hoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.Auth1HoistedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth1(System.getenv().getOrDefault("AUTH1", ""))
                    .build())
            .build();

        Auth1HoistedResponse res = sdk.auth1Hoisted()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth1HoistedResponse](../../models/operations/Auth1HoistedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## auth2Hoisted

### Example Usage

<!-- UsageSnippet language="java" operationID="auth2Hoisted" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuth2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.Auth2HoistedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth2(SchemeAuth2.builder()
                        .username("")
                        .password("")
                        .build())
                    .build())
            .build();

        Auth2HoistedResponse res = sdk.auth2Hoisted()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth2HoistedResponse](../../models/operations/Auth2HoistedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## auth2Preferred

### Example Usage

<!-- UsageSnippet language="java" operationID="auth2Preferred" method="get" path="/op3" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuth2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.Auth2PreferredResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth2(SchemeAuth2.builder()
                        .username("")
                        .password("")
                        .build())
                    .build())
            .build();

        Auth2PreferredResponse res = sdk.auth2Preferred()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Auth2PreferredResponse](../../models/operations/Auth2PreferredResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## opLevelClientCredentials

### Example Usage

<!-- UsageSnippet language="java" operationID="opLevelClientCredentials" method="get" path="/op4" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.OpLevelClientCredentialsResponse;
import org.openapis.openapi.models.operations.OpLevelClientCredentialsSecurity;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        OpLevelClientCredentialsResponse res = sdk.opLevelClientCredentials()
                .security(OpLevelClientCredentialsSecurity.builder()
                    .clientID(System.getenv().getOrDefault("CLIENT_ID", ""))
                    .clientSecret(System.getenv().getOrDefault("CLIENT_SECRET", ""))
                    .build())
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                                              | Type                                                                                                                                   | Required                                                                                                                               | Description                                                                                                                            |
| -------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                                                             | [org.openapis.openapi.models.operations.OpLevelClientCredentialsSecurity](../../models/operations/OpLevelClientCredentialsSecurity.md) | :heavy_check_mark:                                                                                                                     | The security requirements to use for the request.                                                                                      |

### Response

**[OpLevelClientCredentialsResponse](../../models/operations/OpLevelClientCredentialsResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

` // end of snapshot

	snaptest.DoTestSnapshot(t, snaptest.Options{
		Spec:         spec,
		GenYaml:      genYaml,
		IncludeGlobs: expectedSnapshotFiles,
		Expected:     expectedSnapshot,
	})
}
