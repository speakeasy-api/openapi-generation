<!-- Start SDK Example Usage [usage] -->
```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.PostFileRequest;
import org.openapis.openapi.models.operations.PostFileResponse;
import org.openapis.openapi.models.shared.File;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();

        PostFileRequest req = PostFileRequest.builder()
                .upload(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileResponse res = sdk.postFile().request(req).call();
    }
}

```

```java
package hello.world;

import java.lang.Exception;
import java.nio.file.Paths;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.File;
import org.openapis.openapi.models.operations.PostFileWithEncodingRequest;
import org.openapis.openapi.models.operations.PostFileWithEncodingResponse;
import org.openapis.openapi.utils.Blob;

public class Application {

    public static void main(String[] args) throws Error, Exception {

        SDK sdk = SDK.builder().build();

        PostFileWithEncodingRequest req = PostFileWithEncodingRequest.builder()
                .file(File.builder()
                        .fileName("example.file")
                        .content(Blob.from(Paths.get("example.file")))
                        .build())
                .build();

        PostFileWithEncodingResponse res =
                sdk.tag1().postFileWithEncoding().request(req).call();
    }
}

```

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

    public static void main(String[] args)
            throws BadRequestResponseException, Error, Test2ResponseException, Exception {

        SDK sdk = SDK.builder()
                .deprecatedQueryParam1("some example query param")
                .deprecatedQueryParam2("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        PostTest2Response res = sdk.testGroup()
                .tag2()
                .postTest()
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
                                .obj(SimpleObject.builder().str("example").build())
                                .map(Map.ofEntries())
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

### A custom readme heading

A custom usage description

```java
package hello.world;

import java.lang.Exception;
import org.openapis.openapi.SDK;
import org.openapis.openapi.models.errors.BadRequestResponseException;
import org.openapis.openapi.models.errors.Error;
import org.openapis.openapi.models.operations.QueryParam2;
import org.openapis.openapi.models.operations.ResultArray;
import org.openapis.openapi.models.shared.MyApiKey;
import org.openapis.openapi.models.shared.Security;

public class Application {

    public static void main(String[] args) throws BadRequestResponseException, Error, Exception {

        SDK sdk = SDK.builder()
                .queryParam1("some example query param")
                .security(Security.builder()
                        .myApiKey(MyApiKey.builder().myApiKey("<value>").build())
                        .build())
                .build();

        sdk.tag1()
                .listTest1()
                .page(100L)
                .queryParam2(QueryParam2.ONE)
                .headerParam1("some example header param")
                .callAsStreamUnwrapped()
                .forEach((ResultArray item) -> {
                    // handle item
                });
    }
}

```
<!-- End SDK Example Usage [usage] -->