<!-- Start SDK Example Usage [usage] -->
```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.PostFile(ctx, examplealias.PostFileRequest{
		Upload: examplealias.File{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.File != nil {
		// handle response
	}
}

```

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New()

	example, fileErr := os.Open("example.file")
	if fileErr != nil {
		panic(fileErr)
	}

	res, err := s.Tag1.PostFileWithEncoding(ctx, examplealias.PostFileWithEncodingRequest{
		File: examplealias.PostFileWithEncodingFile{
			FileName: "example.file",
			Content:  example,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		// handle response
	}
}

```

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"example.com/openapi-go-sdk/optionalnullable"
	"example.com/openapi-go-sdk/types"
	"log"
	"math/big"
	"os"
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
			Str:        "example",
			Bool:       true,
			Integer:    999999,
			Int32:      1,
			Num:        1.1,
			Float32:    8499.3,
			Date:       types.MustDateFromString("2020-01-01"),
			DateTime:   types.MustTimeFromString("2020-01-01T00:00:00Z"),
			Anything:   "<value>",
			BoolOpt:    examplealias.Pointer(true),
			IntOptNull: examplealias.Pointer[int64](999999),
			NumOptNull: examplealias.Pointer[float64](1.1),
			IntEnum:    examplealias.IntEnumThird.ToPointer(),
			Int32Enum:  examplealias.Int32EnumSixtyNine,
			Bigint:     big.NewInt(593288),
			DecimalStr: types.MustNewDecimalFromString("7028.3"),
			Obj: examplealias.SimpleObject{
				Str: "example",
			},
			Map: map[string]examplealias.SimpleObject{},
			Arr: []examplealias.SimpleObject{},
			Any: examplealias.NewAny(
				"<value>",
			),
			NullableIntEnum:    optionalnullable.From(examplealias.Pointer(examplealias.NullableIntEnumThird)),
			NullableStringEnum: examplealias.NullableStringEnumSecond.ToPointer(),
			Color:              examplealias.ColorGreen.ToPointer(),
			Icon:               examplealias.IconTick,
			HeroWidth:          examplealias.HeroWidthFourHundredAndEighty.ToPointer(),
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

### A custom readme heading

A custom usage description

```go
package main

import (
	"context"
	examplealias "example.com/openapi-go-sdk"
	"log"
	"os"
)

func main() {
	ctx := context.Background()

	s := examplealias.New(
		examplealias.WithQueryParam1("some example query param"),
		examplealias.WithSecurity(examplealias.Security{
			MyAPIKey: &examplealias.MyAPIKey{
				MyAPIKey: os.Getenv("SPEAKEASY_MY_API_KEY"),
			},
		}),
	)

	res, err := s.Tag1.ListTest1(ctx, 100, examplealias.QueryParam2One, "some example header param")
	if err != nil {
		log.Fatal(err)
	}
	if res.Object != nil {
		for {
			// handle items

			res, err = res.Next()

			if err != nil {
				// handle error
			}

			if res == nil {
				break
			}
		}
	}
}

```
<!-- End SDK Example Usage [usage] -->