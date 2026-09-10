package securityusage

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate/snapshots/snaptest"
)

func TestMixedSecurity_Java(t *testing.T) {
	t.Parallel()

	spec := specMixedSecurity

	genYaml := `java:
  groupID: com.example
  artifactID: mixed-security-sdk
  projectName: mixed-security-sdk
`

	expectedSnapshotFiles := []string{
		"docs/sdks/sdk/README.md",
	}

	expectedSnapshot := `--- docs/sdks/sdk/README.md ---
# SDK

## Overview

### Available Operations

* [globalSecurity](#globalsecurity)
* [option1Hoisted](#option1hoisted)
* [option2Hoisted](#option2hoisted)
* [option1NotAllowed](#option1notallowed)
* [opLevelMixedAuth](#oplevelmixedauth)

## globalSecurity

### Example Usage

<!-- UsageSnippet language="java" operationID="globalSecurity" method="get" path="/op0" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuthA2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.components.SecurityOption1;
import org.openapis.openapi.models.operations.GlobalSecurityResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .option1(SecurityOption1.builder()
                        .authA1("<value>")
                        .authA2(SchemeAuthA2.builder()
                            .username("")
                            .password("")
                            .build())
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

## option1Hoisted

### Example Usage

<!-- UsageSnippet language="java" operationID="option1Hoisted" method="get" path="/opA" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.SchemeAuthA2;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.components.SecurityOption1;
import org.openapis.openapi.models.operations.Option1HoistedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .option1(SecurityOption1.builder()
                        .authA1("<value>")
                        .authA2(SchemeAuthA2.builder()
                            .username("")
                            .password("")
                            .build())
                        .build())
                    .build())
            .build();

        Option1HoistedResponse res = sdk.option1Hoisted()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option1HoistedResponse](../../models/operations/Option1HoistedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## option2Hoisted

### Example Usage

<!-- UsageSnippet language="java" operationID="option2Hoisted" method="get" path="/opB" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.components.SecurityOption2;
import org.openapis.openapi.models.operations.Option2HoistedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .option2(SecurityOption2.builder()
                        .authB("<YOUR_JWT>")
                        .build())
                    .build())
            .build();

        Option2HoistedResponse res = sdk.option2Hoisted()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option2HoistedResponse](../../models/operations/Option2HoistedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## option1NotAllowed

### Example Usage

<!-- UsageSnippet language="java" operationID="option1NotAllowed" method="get" path="/opBC" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.components.Security;
import org.openapis.openapi.models.components.SecurityOption3;
import org.openapis.openapi.models.operations.Option1NotAllowedResponse;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
                .security(Security.builder()
                    .option3(SecurityOption3.builder()
                        .authC("<value>")
                        .build())
                    .build())
            .build();

        Option1NotAllowedResponse res = sdk.option1NotAllowed()
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Response

**[Option1NotAllowedResponse](../../models/operations/Option1NotAllowedResponse.md)**

### Errors

| Error Type                 | Status Code                | Content Type               |
| -------------------------- | -------------------------- | -------------------------- |
| models/errors/APIException | 4XX, 5XX                   | \*/\*                      |

## opLevelMixedAuth

### Example Usage

<!-- UsageSnippet language="java" operationID="opLevelMixedAuth" method="get" path="/not/hoisted" -->
` + "`" + `` + "`" + `` + "`" + `java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.operations.OpLevelMixedAuthResponse;
import org.openapis.openapi.models.operations.OpLevelMixedAuthSecurity;
import org.openapis.openapi.models.operations.OpLevelMixedAuthSecurityOption1;

public class Application {

    public static void main(String[] args) throws Exception {

        SDK sdk = SDK.builder()
            .build();

        OpLevelMixedAuthResponse res = sdk.opLevelMixedAuth()
                .security(OpLevelMixedAuthSecurity.builder()
                    .option1(OpLevelMixedAuthSecurityOption1.builder()
                        .authA1("<value>")
                        .authB("<YOUR_JWT>")
                        .build())
                    .build())
                .call();

        // handle response
    }
}
` + "`" + `` + "`" + `` + "`" + `

### Parameters

| Parameter                                                                                                              | Type                                                                                                                   | Required                                                                                                               | Description                                                                                                            |
| ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| ` + "`" + `security` + "`" + `                                                                                                             | [org.openapis.openapi.models.operations.OpLevelMixedAuthSecurity](../../models/operations/OpLevelMixedAuthSecurity.md) | :heavy_check_mark:                                                                                                     | The security requirements to use for the request.                                                                      |

### Response

**[OpLevelMixedAuthResponse](../../models/operations/OpLevelMixedAuthResponse.md)**

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
