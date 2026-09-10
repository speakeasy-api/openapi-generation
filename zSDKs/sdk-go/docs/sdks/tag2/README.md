# TestGroup.Tag2

## Overview

### Available Operations

* [PostTest](#posttest) - Post Test2

## PostTest

This is a test endpoint.
It has a description.

### Example Usage

<!-- UsageSnippet language="go" operationID="postTest2" method="post" path="/test2" -->
```go
package main

import(
	"context"
	"os"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/types"
	"math/big"
	"example.com/openapi-go-sdk/optionalnullable"
	"log"
)

func main() {
    ctx := context.Background()

    s := examplealias.New(
        examplealias.WithDeprecatedQueryParam1("some example query param"),
        examplealias.WithDeprecatedQueryParam2("some example query param"),
        examplealias.WithSecurity(examplealias.Security{
            MyAPIKey: &examplealias.MyAPIKey{
                MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
            },
        }),
    )

    res, err := s.TestGroup.Tag2.PostTest(ctx, examplealias.Test2Request{
        Obj: examplealias.ExhaustiveObject{
            Str: "example",
            Bool: true,
            Integer: 999999,
            Int32: 1,
            Num: 1.1,
            Float32: 8499.3,
            Date: types.MustDateFromString("2020-01-01"),
            DateTime: types.MustTimeFromString("2020-01-01T00:00:00Z"),
            Anything: "<value>",
            BoolOpt: examplealias.Pointer(true),
            IntOptNull: examplealias.Pointer[int64](999999),
            NumOptNull: examplealias.Pointer[float64](1.1),
            IntEnum: examplealias.IntEnumThird.ToPointer(),
            Int32Enum: examplealias.Int32EnumSixtyNine,
            Bigint: big.NewInt(593288),
            DecimalStr: types.MustNewDecimalFromString("7028.3"),
            Obj: examplealias.SimpleObject{
                Str: "example",
            },
            Map: map[string]examplealias.SimpleObject{

            },
            Arr: []examplealias.SimpleObject{},
            Any: examplealias.NewAny(
                "<value>",
            ),
            NullableIntEnum: optionalnullable.From(examplealias.Pointer(examplealias.NullableIntEnumThird)),
            NullableStringEnum: examplealias.NullableStringEnumSecond.ToPointer(),
            Color: examplealias.ColorGreen.ToPointer(),
            Icon: examplealias.IconTick,
            HeroWidth: examplealias.HeroWidthFourHundredAndEighty.ToPointer(),
        },
        Type: examplealias.TypeSuperType1.ToPointer(),
    })
    if err != nil {
        log.Fatal(err)
    }
    if res.Body != nil {
        // handle response
    }
}
```

### Parameters

| Parameter                                                                                                               | Type                                                                                                                    | Required                                                                                                                | Description                                                                                                             | Example                                                                                                                 |
| ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `ctx`                                                                                                                   | [context.Context](https://pkg.go.dev/context#Context)                                                                   | :heavy_check_mark:                                                                                                      | The context to use for the request.                                                                                     |                                                                                                                         |
| `test2Request`                                                                                                          | [Test2Request](../../test2request.md)                                                                                   | :heavy_check_mark:                                                                                                      | N/A                                                                                                                     |                                                                                                                         |
| `deprecatedQueryParam1`                                                                                                 | `*string`                                                                                                               | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `deprecatedQueryParam2`                                                                                                 | `*string`                                                                                                               | :heavy_minus_sign:                                                                                                      | : warning: ** DEPRECATED **: This will be removed in a future release, please migrate away from it as soon as possible. | some example query param                                                                                                |
| `opts`                                                                                                                  | [][examplealias.Option](../../option.md)                                                                                | :heavy_minus_sign:                                                                                                      | The options for this request.                                                                                           |                                                                                                                         |

### Response

**[*PostTest2Response](../../posttest2response.md), error**

### Errors

| Error Type                           | Status Code                          | Content Type                         |
| ------------------------------------ | ------------------------------------ | ------------------------------------ |
| examplealias.BadRequestResponseError | 400                                  | application/json                     |
| examplealias.ErrorsError             | 404                                  | application/json                     |
| examplealias.Test2ResponseError      | 500                                  | application/json                     |
| examplealias.SDKError                | 4XX, 5XX                             | \*/\*                                |