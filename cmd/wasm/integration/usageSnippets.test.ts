import { beforeAll, describe, expect, it } from "vitest";
import { getTestFileContents, getWasmFunction } from "./testHelpers";

describe("usage snippets", () => {
  let GenerateUsageSnippets: (...args: any[]) => Promise<string>;

  beforeAll(async () => {
    GenerateUsageSnippets = await getWasmFunction("GenerateUsageSnippets");
  });
  
	it("should generate usage snippets", async () => {
		const result = await GenerateUsageSnippets({
			schema: `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getPets
      summary: Test
      responses:
        '200':
          description: A test response
    post:
      operationId: createPet
      summary: Create a new pet
      responses:
        '200':
          description: A test response
      requestBody:
        content:
          application/json:
            schema:
              properties:
                name:
                  type: string
                age:
                  type: integer
                breed:
                  type: string
              required:
                - name
                - age
                - breed`,
			target: "typescriptv2",
			operationIds: ["getPets", "createPet"]
		});
		
    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(2);

    const getPets = parsed.find((snippet: any) => snippet.operationId === "getPets");
    expect(getPets).toBeDefined();
    const getPetsSnippet = getPets.content;
    expect(getPetsSnippet).toMatchInlineSnapshot(`
      "// Usage snippet provided for getPets (get /test)
      import { SDK } from "openapi";

      const sdk = new SDK({
        serverURL: "https://api.example.com",
      });

      async function run() {
        await sdk.getPets();


      }

      run();


      "
    `)

    const createPet = parsed.find((snippet: any) => snippet.operationId === "createPet");
    expect(createPet).toBeDefined();
    const createPetSnippet = createPet.content;
    expect(createPetSnippet).toMatchInlineSnapshot(`
      "// Usage snippet provided for createPet (post /test)
      import { SDK } from "openapi";

      const sdk = new SDK({
        serverURL: "https://api.example.com",
      });

      async function run() {
        await sdk.createPet();


      }

      run();


      "
    `)
  });

  it("should generate usage snippets for an enterprise feature (jsonl)", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/jsonl.yaml"),
      operationIds: ["getLogs"],
      target: "typescriptv2",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);

    const getLogs = parsed.find((snippet: any) => snippet.operationId === "getLogs");
    expect(getLogs).toBeDefined();
    const getLogsSnippet = getLogs.content;

    expect(getLogsSnippet).toMatchInlineSnapshot(`
      "// Usage snippet provided for getLogs (get /logs)
      import { SDK } from "openapi";

      const sdk = new SDK({
        serverURL: "https://api.example.com",
      });

      async function run() {
        const result = await sdk.getLogs();

        for await (const event of result) {
          // Handle the event
          console.log(event);
        }
      }

      run();


      "
    `)
  })

  it("should generate usage snippets with custom config", async () => {
    const schema = await getTestFileContents("testdata/sample-asset-api.json");
    const result = await GenerateUsageSnippets({
      schema: schema,
      operationIds: ["createAsset"],
      target: "typescriptv2",
      config: {
        "sdkClassName": "CustomSDK",
        "packageName": "custom"
      }
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);

    const createAsset = parsed.find((snippet: any) => snippet.operationId === "createAsset");
    expect(createAsset).toBeDefined();
    const createAssetSnippet = createAsset.content;
    expect(createAssetSnippet).toMatchInlineSnapshot(`
      "// Usage snippet provided for createAsset (post /assets)
      import { CustomSDK } from "custom";

      const customSDK = new CustomSDK({
        token: "EXAMPLE_ASSET_API_KEY",
      });

      async function run() {
        const result = await customSDK.create({
          name: "Sample image",
          source: "https://example.com/image.png",
          sizes: [
            {
              width: 320,
              height: 240,
            },
            {
              width: 640,
              height: 480,
            },
          ],
        });

        console.log(result);
      }

      run();


      "
    `)
  })

  it("should generate usage snippets for python", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "python",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "# Usage snippet provided for getPets (get /pets)
      from openapi import SDK


      with SDK() as sdk:

          res = sdk.get_pets()

          # Handle response
          print(res)



      "
    `);
  });

  it("should generate usage snippets for go", async () => {
    const result = await GenerateUsageSnippets({
      config: {
        modulePath: "example.com/openapi",
        packageName: "",
        sdkPackageName: "openapi",
      },
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "go",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "// Usage snippet provided for getPets (get /pets)
      package main

      import(
      	"context"
      	openapi "example.com/openapi/v2"
      	"log"
      )

      func main() {
          ctx := context.Background()

          s := openapi.New()

          res, err := s.GetPets(ctx)
          if err != nil {
              log.Fatal(err)
          }
          if res.Pets != nil {
              // handle response
          }
      }


      "
    `);
  });

  it.skip("should reject in the case of a panic", async () => {
    // Leaving this here as a demonstration that panics are recovered and returned as
    // promise rejections.
    // To test, you can add a panic anywhere in the snippet generationcode and it will be recovered
    // and returned as a rejection.
    expect(GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "go",
    })).rejects.toThrow("panic occurred: GetTargetFromTargetString is deprecated")
  })

  it("should generate usage snippets for java", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "javav2",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "// Usage snippet provided for getPets (get /pets)
      package hello.world;

      import java.lang.Exception;
      import org.openapis.openapi.SDK;
      import org.openapis.openapi.models.operations.GetPetsResponse;

      public class Application {

          public static void main(String[] args) throws Exception {

              SDK sdk = SDK.builder()
                  .build();

              GetPetsResponse res = sdk.getPets()
                      .call();

              if (res.pets().isPresent()) {
                  System.out.println(res.pets().get());
              }
          }
      }


      "
    `);
  });

  it("should generate usage snippets for csharp", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "csharp",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "// Usage snippet provided for getPets (get /pets)
      using Openapi;

      var sdk = new SDK();

      var res = await sdk.GetPetsAsync();

      // handle response


      "
    `);
  });

  it("should generate usage snippets for python", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds: ["getPets"],
      target: "pythonv2",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "# Usage snippet provided for getPets (get /pets)
      from openapi import SDK


      with SDK() as sdk:

          res = sdk.get_pets()

          # Handle response
          print(res)



      "
    `);
  })

  it("should generate usage snippets for php", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds:   ["getPets"],
      target: "php",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "# Usage snippet provided for getPets (get /pets)
      declare(strict_types=1);

      require 'vendor/autoload.php';

      use OpenAPI\\OpenAPI;

      $sdk = OpenAPI\\SDK::builder()->build();



      $response = $sdk->getPets(

      );

      if ($response->pets !== null) {
          // handle response
      }


      "
    `);
  })

  it("should generate usage snippets for mcp", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds:   ["getPets"],
      target: "mcp-typescript",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "
      {}


      "
    `);
  })

  it("should generate usage snippets for mcp with addPet", async () => {
    const result = await GenerateUsageSnippets({
      schema: await getTestFileContents("testdata/petstore.yaml"),
      operationIds:   ["addPet"],
      target: "mcp-typescript",
    })

    const parsed = JSON.parse(result);
    expect(parsed).toBeDefined();
    expect(parsed.length).toBe(1);
    expect(parsed[0].content).toMatchInlineSnapshot(`
      "
      {
        \"type\": \"object\",
        \"properties\": {
          \"request\": {
            \"type\": \"object\",
            \"properties\": {
              \"id\": {
                \"type\": \"integer\"
              },
              \"name\": {
                \"type\": \"string\"
              },
              \"species\": {
                \"type\": \"string\"
              },
              \"age\": {
                \"type\": \"integer\"
              }
            }
          }
        },
        \"required\": [
          \"request\"
        ]
      }


      "
    `);
  })
  
});
