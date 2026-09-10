import { describe, it, expect, beforeAll } from "vitest";
import { getWasmFunction } from "./testHelpers";
import { load as parseYAML } from "js-yaml";

const original = `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
`

const modified = `openapi: 3.1.0
info:
  title: Test API
  version: 1.0.0
paths:
  /test:
    get:
      operationId: getTest
      responses:
        '200':
          description: OK
        '400':
          description: Bad Request
`



const existingOverlay = `overlay: 1.0.0
info:
  title: Test API
  version: 1.0.0
actions:
  - target: $["paths"]["/test"]["get"]["responses"]
    update:
      '400':
        description: Bad Request (Existing overlay)
`

describe("overlay utils", () => {
  let CalculateOverlay: (...args: any[]) => Promise<string>;
  let FormatYAML: (...args: any[]) => Promise<string>;

  beforeAll(async () => {
    CalculateOverlay = await getWasmFunction("CalculateOverlay");
    FormatYAML = await getWasmFunction("FormatYAML");
  }, 30000);
  
  describe("calculate overlay", () => {
    it("should calculate the overlay correctly with no overlay", async () => {
      const result = await CalculateOverlay({ original, modified, overlay: "" });
      expect(result).toBeDefined();
      expect(result).toMatchInlineSnapshot(`
        "overlay: ""
        x-speakeasy-jsonpath: rfc9535
        info:
            title: ""
            version: ""
        actions:
            - target: $["paths"]["/test"]["get"]["responses"]
              update:
                '400':
                    description: Bad Request
        "
      `)
    })

    it("should calculate the overlay correctly with an existing overlay", async () => {
      const result = await CalculateOverlay({ original, modified, overlay: existingOverlay });
      expect(result).toBeDefined();
      expect(result).toMatchInlineSnapshot(`
        "overlay: 1.0.0
        x-speakeasy-jsonpath: rfc9535
        info:
            title: Test API
            version: 1.0.0
        actions:
            - target: $["paths"]["/test"]["get"]["responses"]
              update:
                '400':
                    description: Bad Request (Existing overlay)
            - target: $["paths"]["/test"]["get"]["responses"]["400"]["description"]
              update: Bad Request
        "
      `)
    })

    it("should generate consistent YAML formatting", async () => {
      // This test validates that overlay calculation produces consistent, standard YAML formatting
      // instead of trying to preserve arbitrary original formatting

      const originalSpec = `openapi: "3.1.0"
info:
    title: "Text Generation API"
    description: "An API for building text generation applications"
    version: "1.0.0"`

      const modifiedSpec = `openapi: "3.1.0"
info:
    title: "Text Generation API and some other stuff"
    description: "An API for building text generation applications"
    version: "1.0.0"`

      // Existing overlay with specific formatting (quotes, indentation)
      const existingOverlay = `overlay: 1.0.0
x-speakeasy-jsonpath: rfc9535
info:
  title: "Speakeasy Modifications"
  version: "1.0.0"

actions:
  - target: $.openapi
    update: "3.1.0"

  - target: $.components.schemas.TokenGenerationConfig.properties.responseModalities
    description: There seems to be a rogue responseModalities field
    remove: true`

      const result = await CalculateOverlay({
        original: originalSpec,
        modified: modifiedSpec,
        overlay: existingOverlay
      });

      expect(result).toBeDefined();

      // Should have standard YAML formatting
      expect(result).toContain('overlay: 1.0.0'); // Standard formatting (no quotes)
      expect(result).toContain('x-speakeasy-jsonpath: rfc9535');
      expect(result).toContain('title: Speakeasy Modifications'); // Standard formatting (no quotes)

      // Should contain existing actions
      expect(result).toContain('target: $.openapi');
      expect(result).toContain('update: "3.1.0"'); // Quotes preserved where needed for strings
      expect(result).toContain('target: $.components.schemas.TokenGenerationConfig.properties.responseModalities');
      expect(result).toContain('remove: true');

      // Should add new action for the title change
      expect(result).toContain('target: $["info"]["title"]');
      expect(result).toContain('update: "Text Generation API and some other stuff"');

      // Should be valid YAML
      expect(() => parseYAML(result)).not.toThrow();
    })
  })

  describe("format yaml", () => {
    it("should normalize YAML formatting consistently", async () => {
      const messyYAML = `overlay: 1.0.0
x-speakeasy-jsonpath:    rfc9535
info:
  title:   "Test Overlay"
  version:  "1.0.0"

actions:
  -  target: $["info"]["title"]
     update: "New Title"
  -   target: $["info"]["description"]
      update:   "New Description"`

      const result = await FormatYAML({ yamlContent: messyYAML });
      expect(result).toBeDefined();

      // Should have consistent formatting
      expect(result).toContain('x-speakeasy-jsonpath: rfc9535');
      expect(result).toContain('title: Test Overlay');
      expect(result).toContain('version: 1.0.0');
      expect(result).toContain('- target: $["info"]["title"]');
      expect(result).toContain('  update: "New Title"');

      // Should be valid YAML
      const formatted2 = await FormatYAML({ yamlContent: result });
      expect(formatted2).toBe(result); // Should be idempotent
    });

    it("should handle empty and whitespace-only YAML", async () => {
      expect(await FormatYAML({ yamlContent: "" })).toBe("");
      expect(await FormatYAML({ yamlContent: "   " })).toBe("   ");
      expect(await FormatYAML({ yamlContent: "\n\n" })).toBe("\n\n");
    });

    it("should preserve content while normalizing formatting", async () => {
      const originalOverlay = `overlay: 1.0.0
info:
  title: Test API
  version: 1.0.0
actions:
  - target: $["paths"]["/test"]["get"]["responses"]
    update:
      '400':
        description: Bad Request (Existing overlay)`;

      const formatted = await FormatYAML({ yamlContent: originalOverlay });
      expect(formatted).toBeDefined();
      expect(formatted).toContain('overlay: 1.0.0');
      expect(formatted).toContain('Bad Request (Existing overlay)');
    });
  })

  describe("apply overlay", () => {
    let ApplyOverlay: (...args: any[]) => Promise<string>;

    beforeAll(async () => {
      ApplyOverlay = await getWasmFunction("ApplyOverlay");
    }, 30000);

    it("should apply the overlay correctly with no overlay", async () => {
      const result = await ApplyOverlay({ original, overlay:existingOverlay });
      expect(result).toBeDefined();
      expect(result).toMatchInlineSnapshot(`"{"type":"success","result":"openapi: 3.1.0\\ninfo:\\n    title: Test API\\n    version: 1.0.0\\npaths:\\n    /test:\\n        get:\\n            operationId: getTest\\n            responses:\\n                '200':\\n                    description: OK\\n                '400':\\n                    description: Bad Request (Existing overlay)\\n"}"`)
    })
  })
})
