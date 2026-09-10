package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestAndSecurity_Java(t *testing.T) {
	t.Parallel()

	spec := specAndSecurity

	genYaml := `java:
  groupID: com.example
  artifactID: and-security-sdk
  projectName: and-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [andAuthHoisted](#andauthhoisted)
* [opLevelAndAuth](#oplevelandauth)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="java" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuth2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.GlobalSecurityResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth1(System.getenv().getOrDefault("AUTH1", ""))
                    .auth2(SchemeAuth2.builder()
                        .username("")
                        .password("")
                        .build())
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

## andAuthHoisted

### Example Usage

<!-- UsageSnippet language="java" operationID="andAuthHoisted" method="get" path="/op1" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuth2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.operations.AndAuthHoistedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .auth1(System.getenv().getOrDefault("AUTH1", ""))
                    .auth2(SchemeAuth2.builder()
                        .username("")
                        .password("")
                        .build())
                    .build())
            .build();

        AndAuthHoistedResponse res = sdk.andAuthHoisted()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[AndAuthHoistedResponse](../../models/operations/AndAuthHoistedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## opLevelAndAuth

### Example Usage

<!-- UsageSnippet language="java" operationID="opLevelAndAuth" method="get" path="/op2" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuth2;
import org.openapis.openapi.models.operations.OpLevelAndAuthResponse;
import org.openapis.openapi.models.operations.OpLevelAndAuthSecurity;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        OpLevelAndAuthResponse res = sdk.opLevelAndAuth()
                .security(OpLevelAndAuthSecurity.builder()
                    .auth2(SchemeAuth2.builder()
                        .username("")
                        .password("")
                        .build())
                    .auth3(System.getenv().getOrDefault("AUTH3", ""))
                    .build())
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                          | Type                                                                                                               | Required                                                                                                           | Description                                                                                                        |
| ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| ` + "`" + `security` + "`" + `                                                                                                         | [org.openapis.openapi.models.operations.OpLevelAndAuthSecurity](../../models/operations/OpLevelAndAuthSecurity.md) | :heavy_check_mark:                                                                                                 | The security requirements to use for the request.                                                                  |

### Response

**[OpLevelAndAuthResponse](../../models/operations/OpLevelAndAuthResponse.md)**

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
