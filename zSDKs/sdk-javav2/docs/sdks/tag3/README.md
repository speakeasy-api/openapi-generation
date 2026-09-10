# TestGroup.Tag3

## Overview

### Available Operations

* [postTest](#posttest) - Post Test2

## postTest

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="java" operationID="postTest2" method="post" path="/test2" -->
```java
package hello.world;

import java.lang.Exception;
import java.math.BigDecimal;
import java.math.BigInteger;
import java.time.LocalDate;
import java.time.OffsetDateTime;
import java.util.List;
import java.util.Map;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.errors.Test2ResponseException;
import org.openapis.openapi.models.operations.PostTest2Response;
import org.openapis.openapi.models.shared.Any;
import org.openapis.openapi.models.shared.Color;
import org.openapis.openapi.models.shared.ExhaustiveObject;
import org.openapis.openapi.models.shared.HeroWidth;
import org.openapis.openapi.models.shared.Icon;
import org.openapis.openapi.models.shared.Int32Enum;
import org.openapis.openapi.models.shared.IntEnum;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.NullableIntEnum;
import org.openapis.openapi.models.shared.NullableStringEnum;
import org.openapis.openapi.models.shared.Security;
import org.openapis.openapi.models.shared.SimpleObject;
import org.openapis.openapi.models.shared.Test2Request;
import org.openapis.openapi.models.shared.Type;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Test2ResponseException, Exception {

        SDK sdk = SDK.builder()
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .security(Security.builder()
                    .myApiKey(MyApiKey.builder()
                        .myApiKey("<value>")
                        .build())
                    .build())
            .build();

        PostTest2Response res = sdk.testGroup().tag3().postTest()
                .test2Request(Test2Request.builder()
                    .obj(ExhaustiveObject.builder()
                        .str("example")
                        .bool(true)
                        .integer(999999L)
                        .int32(1)
                        .num(1.1)
                        .float32(8499.3f)
                        .date(LocalDate.parse("2020-01-01"))
                        .dateTime(OffsetDateTime.parse("2020-01-01T00:00:00Z"))
                        .anything("<value>")
                        .int32Enum(Int32Enum.SIXTY_NINE)
                        .bigint(new BigInteger("593288"))
                        .decimalStr(new BigDecimal("7028.3"))
                        .obj(SimpleObject.builder()
                            .str("example")
                            .build())
                        .map(Map.ofEntries(
                        ))
                        .arr(List.of())
                        .any(Any.of("<value>"))
                        .nullableStringEnum(NullableStringEnum.SECOND)
                        .icon(Icon.TICK)
                        .boolOpt(true)
                        .intOptNull(999999L)
                        .numOptNull(1.1)
                        .intEnum(IntEnum.Third)
                        .nullableIntEnum(NullableIntEnum.Third)
                        .color(Color.GREEN)
                        .heroWidth(HeroWidth.FOUR_HUNDRED_AND_EIGHTY)
                        .build())
                    .type(Type.SuperType1)
                    .build())
                .call();

        if (res.body().isPresent()) {
            System.out.println(res.body().get());
        }
    }
}
```

### Parameters

| Parameter                                                                                                               | Type                                                                                                                    | Required                                                                                                                | Description                                                                                                             | Example                                                                                                                 |
| ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `deprecatedQueryParam1`                                                                                                 | @Nullable *String*                                                                                                      | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `deprecatedQueryParam2`                                                                                                 | @Nullable *String*                                                                                                      | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `test2Request`                                                                                                          | [Test2Request](../../models/shared/Test2Request.md)                                                                     | :heavy_check_mark:                                                                                                      | N/A                                                                                                                     |                                                                                                                         |
| `serverURL`                                                                                                             | *String*                                                                                                                | :heavy_minus_sign:                                                                                                      | An optional server URL to use.                                                                                          | http://localhost:8080                                                                                                   |

### Response

**[PostTest2Response](../../models/operations/PostTest2Response.md)**

### Errors

| Error Type                                | Status Code                               | Content Type                              |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| models/errors/BadRequestResponseException | 400                                       | application/json                          |
| models/errors/Error                       | 404                                       | application/json                          |
| models/errors/Test2ResponseException      | 500                                       | application/json                          |
| models/errors/SDKException                | 4XX, 5XX                                  | \*/\*                                     |