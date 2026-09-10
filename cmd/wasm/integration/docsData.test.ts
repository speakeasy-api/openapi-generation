import { describe, it, expect, beforeAll } from "vitest";
import { readFile } from "fs/promises";
import { join } from "path";
import { load } from "js-yaml";
import {  gunzipAsync } from "./testHelpers";

// Do some very basic testing. In time, we'll add more complex specs with more
// in-depth tests.
describe("WASM Docs Data Serialization", () => {
  let SerializeDocsData: (...args: any[]) => Promise<string>;

  beforeAll(async () => {
       // Read the gzipped WASM file as a buffer
       const gzippedBuffer = await readFile(join(__dirname, "../assets/wasm/full-stack.wasm.gz"));

       // Decompress the gzipped buffer
       const wasmBuffer = await gunzipAsync(gzippedBuffer);
   
       // Instantiate the WASM module
       const go = new Go();
       const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
   
       // @ts-expect-error - typing is wrong
       go.run(result.instance);
   
       // Get the global functions
       const global = globalThis as unknown as { [key: string]: any };
       SerializeDocsData = global.SerializeDocsData;
  });

  it("should successfully serialize a simple spec", async () => {
    const schemaPath = join(__dirname, "testdata/petstore.yaml");
    const yaml = load(await readFile(schemaPath, "utf8"));
    const schema = JSON.stringify(yaml);
    const serializedDocsData = await SerializeDocsData(schema);
    const parsedData = JSON.parse(serializedDocsData).map((chunk: string) => JSON.parse(chunk));
    
    // Basic validation of the data structure
    expect(parsedData).toBeDefined();
    expect(parsedData.length).toEqual(8);

    type MinimalChunk = { chunkType: string }
    expect(parsedData.filter((c: MinimalChunk) => c.chunkType === "about").length).toEqual(1);
    expect(parsedData.filter((c: MinimalChunk) => c.chunkType === "schema").length).toEqual(3);
    expect(parsedData.filter((c: MinimalChunk) => c.chunkType === "operation").length).toEqual(3);
    expect(parsedData.filter((c: MinimalChunk) => c.chunkType === "tag").length).toEqual(1);
  });

  it("should handle invalid schema gracefully", async () => {
    const invalidSchema = '{ "invalid": "json", "malformed": ';
    await expect(SerializeDocsData(invalidSchema)).rejects.toThrow();
  });
});
