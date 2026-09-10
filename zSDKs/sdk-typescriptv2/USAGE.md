<!-- Start SDK Example Usage [usage] -->
```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.postFile({
    upload: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

```typescript
import { openAsBlob } from "node:fs";
import { SDK } from "openapi";

const sdk = new SDK();

async function run() {
  const result = await sdk.tag1.postFileWithEncoding({
    file: await openAsBlob("example.file"),
  });

  console.log(result);
}

run();

```

```typescript
import {
  HeroWidth,
  Icon,
  Int32Enum,
  IntEnum,
  NullableIntEnum,
  NullableStringEnum,
  SDK,
  Type,
} from "openapi";
import { Decimal } from "openapi/types";

const sdk = new SDK({
  deprecatedQueryParam1: "some example query param",
  deprecatedQueryParam2: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.testGroup.tag2.postTest({
    obj: {
      str: "example",
      bool: true,
      integer: 999999,
      int32: 1,
      num: 1.1,
      float32: 8499.3,
      date: new Date("2020-01-01"),
      dateTime: new Date("2020-01-01T00:00:00Z"),
      anything: "<value>",
      boolOpt: true,
      intOptNull: 999999,
      numOptNull: 1.1,
      intEnum: IntEnum.Third,
      int32Enum: Int32Enum.SixtyNine,
      bigint: 593288,
      decimalStr: new Decimal("7028.3"),
      obj: {
        str: "example",
      },
      map: {},
      arr: [],
      any: "<value>",
      nullableIntEnum: NullableIntEnum.Third,
      nullableStringEnum: NullableStringEnum.Second,
      color: "green",
      icon: Icon.Tick,
      heroWidth: HeroWidth.FourHundredAndEighty,
    },
    type: Type.SuperType1,
  });

  console.log(result);
}

run();

```

### A custom readme heading

A custom usage description

```typescript
import { QueryParam2, SDK } from "openapi";

const sdk = new SDK({
  queryParam1: "some example query param",
  security: {
    myApiKey: {
      myApiKey: process.env["SPEAKEASY_MY_API_KEY"] ?? "",
    },
  },
});

async function run() {
  const result = await sdk.tag1.listTest1(
    100,
    QueryParam2.One,
    "some example header param",
  );

  for await (const page of result) {
    console.log(page);
  }
}

run();

```
<!-- End SDK Example Usage [usage] -->