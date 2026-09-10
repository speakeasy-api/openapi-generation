<!-- Start SDK Example Usage [usage] -->
```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileRequest req = new PostFileRequest() {
    Upload = new Speakeasy.OpenAPI.File() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.PostFileAsync(req);

// handle response
```

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK();

PostFileWithEncodingRequest req = new PostFileWithEncodingRequest() {
    File = new PostFileWithEncodingFile() {
        FileName = "example.file",
        Content = System.IO.File.ReadAllBytes("example.file"),
    },
};

var res = await sdk.Tag1.PostFileWithEncodingAsync(req);

// handle response
```

```csharp
using NodaTime;
using Speakeasy.OpenAPI;
using System;
using System.Collections.Generic;
using System.Numerics;

var sdk = new SDK(
    deprecatedQueryParam1: "some example query param",
    deprecatedQueryParam2: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

var res = await sdk.TestGroup.Tag2.PostTestAsync(test2Request: new Test2Request() {
    Obj = new ExhaustiveObject() {
        Str = "example",
        Bool = true,
        Integer = 999999,
        Int32 = 1,
        Num = 1.1D,
        Float32 = 8499.3F,
        Date = LocalDate.FromDateTime(System.DateTime.Parse("2020-01-01")),
        DateTime = System.DateTime.Parse("2020-01-01T00:00:00Z").ToUniversalTime(),
        Anything = "<value>",
        BoolOpt = true,
        IntOptNull = 999999,
        NumOptNull = 1.1D,
        IntEnum = IntEnum.Third,
        Int32Enum = Int32Enum.SixtyNine,
        Bigint = BigInteger.Parse("702830"),
        DecimalStr = 3858.6M,
        Obj = new SimpleObject() {
            Str = "example",
        },
        Map = new Dictionary<string, SimpleObject>() {
            { "key", new SimpleObject() {
                 Str = "example",
             } },
        },
        Arr = new List<SimpleObject>() {
            new SimpleObject() {
                Str = "example",
            },
            new SimpleObject() {
                Str = "example",
            },
        },
        Any = Any.CreateSimpleObject(new SimpleObject() {
            Str = "example",
        }),
        NullableIntEnum = NullableIntEnum.Third,
        NullableStringEnum = NullableStringEnum.Second,
        Color = Color.Green,
        Icon = Icon.Tick,
        HeroWidth = HeroWidth.FourHundredAndEighty,
    },
    Type = Speakeasy.OpenAPI.Type.SuperType1,
});

// handle response
```

### A custom readme heading

A custom usage description

```csharp
using Speakeasy.OpenAPI;

var sdk = new SDK(
    queryParam1: "some example query param",
    security: new Security() {
        MyApiKey = new MyApiKey() {
            MyApiKeyValue = "<YOUR_API_KEY_HERE>",
        },
    }
);

ListTest1Response? res = await sdk.Tag1.ListTest1Async(
    page: 100,
    queryParam2: QueryParam2.One,
    headerParam1: "some example header param"
);

while (res != null)
{
    // handle items

    res = await res.Next!();
}
```
<!-- End SDK Example Usage [usage] -->