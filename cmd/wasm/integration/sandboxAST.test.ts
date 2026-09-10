import * as path from "path";

import { beforeAll, describe, expect, it } from 'vitest';
import { getTestFileContents, getWasmFunction } from './testHelpers.js';

import { load } from 'js-yaml';
import { readFile } from "fs/promises";

interface TestCase {
  name: string
  schema: string
  extension: string
  assertFn: (ast: any) => void
}

const testCases: TestCase[] = [
  {
    name: 'petstore',
    schema: './testdata/petstore.yaml',
    extension: '.yaml',
    assertFn: (ast: any) => {
      expect(ast.openAPIVersion).toBe('3.0.0');
      expect(ast.docInfo.title).toBe('Petstore API');
      expect(ast.operations.length).toBe(3);
      expect(ast.operations.map(op => op.operationID)).toEqual([
        'getPets',
        'addPet',
        'getPetById',
      ]);
      expect(ast.mcp).toBeDefined();
      expect(ast.mcp.tools.length).toBe(3);
      expect(ast.mcp.tools.map(op => op.name)).toEqual([
        'get_pets',
        'add_pet',
        'get_pet_by_id',
      ]);
      expect(ast.mcp.tools.map(op => op.disabled)).toEqual([
        false,
        false,
        false,
      ]);
    }
  },
  {
    name: 'sample asset API',
    schema: './testdata/sample-asset-api.json',
    extension: '.json',
    assertFn: (ast: any) => {
      expect(ast.openAPIVersion).toBe('3.0.3');
      expect(ast.operations.map(op => op.operationID)).toEqual([
        'listAssets',
        'createAsset',
        'getAsset',
      ]);
    }
  }
]


describe('WASM AST Serialization', () => {
  let SerializeSandboxAST: (...args: any[]) => Promise<string>;

  beforeAll(async () => {
    SerializeSandboxAST = await getWasmFunction("SerializeSandboxAST");
  });

  testCases.forEach(testCase => {
    it(`should successfully serialize AST for ${testCase.name}`, async () => {
      const testData = await getTestFileContents(testCase.schema);
      const serializedAST = await SerializeSandboxAST(testData);
      const parsedAST = JSON.parse(serializedAST);

      // Basic validation of the AST structure
      expect(parsedAST).toBeDefined();
      testCase.assertFn(parsedAST);
    });
  });

  it('should handle invalid schema gracefully', async () => {
    const invalidSchema = '{ "invalid": "json", "malformed": ';
    await expect(SerializeSandboxAST(invalidSchema)).rejects.toThrow();
  });

  it("should return a JSONPath for an operation without an operationId", async () => {
    const schema = `
      openapi: 3.0.0
      info:
        title: Petstore API
        version: 1.0.0
      paths:
        /pets:
          get:
            summary: Get all pets
    `
    const serializedAST = await SerializeSandboxAST(schema);
    const parsedAST = JSON.parse(serializedAST);
    expect(parsedAST.operations[0].jsonPath).toBe("$.paths['/pets'].get");
  });

  it("should return an empty array for operations if there are none", async () => {
    const schema = `
      openapi: 3.0.0
      info:
        title: Petstore API
        version: 1.0.0
      paths:
        /pets:
    `
    const serializedAST = await SerializeSandboxAST(schema);
    const parsedAST = JSON.parse(serializedAST);
    expect(parsedAST.operations.length).toEqual(0);
  });

  describe('displayName', () => {
    it('should return the x-speakeasy-name-override if present', async () => {
      const schema = `
        openapi: 3.0.0
        info:
          title: Petstore API
        paths:
          /pets:
            get:
              operationId: getPets
              summary: Get all pets
              x-speakeasy-name-override: get-all-pets
              
      `
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.operations[0].displayName).toBe('get-all-pets');
    });

    it('should return the operationId if no x-speakeasy-name-override is present', async () => {
      const schema = `
        openapi: 3.0.0
        info:
          title: Petstore API 
        paths:
          /pets:
            get:
              operationId: getPets
              summary: Get all pets
      `
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.operations[0].displayName).toBe('getPets');
    });

    it('should return a humanized jsonpath if no x-speakeasy-name-override or operationId is present', async () => {
      const schema = `
        openapi: 3.0.0
        info:
          title: Petstore API
        paths:
          /pets:
            get:
              summary: Get all pets
      `
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.operations[0].displayName).toBe('/pets/get');
    })
  });
  describe('group tree', () => {
    it('should create a group tree object key', async () => {
      const schema = await readFile(path.join(__dirname, 'testdata/petstore-nested-tags.yaml'), 'utf8');
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
    })
    it('should create a group tree from the petstore schema', async () => {
      const schema = await readFile(path.join(__dirname, 'testdata/petstore-nested-tags.yaml'), 'utf8');
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
      expect(parsedAST.groupTree.children['pets']).toBeDefined();
    })
    it('should create a group tree with two levels of depth when there are two dots in the tag', async () => {
      const schema = await readFile(path.join(__dirname, 'testdata/petstore-nested-tags.yaml'), 'utf8');
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
      expect(parsedAST.groupTree.children['pets']).toBeDefined();
      expect(Object.keys(parsedAST.groupTree.children['pets'].children)).toHaveLength(2);
    })
    it('should create a flat group if there are no overrides via x-speakeasy-group', async () => {
      const schema = await readFile(path.join(__dirname, 'testdata/petstore.yaml'), 'utf8');
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
      expect(parsedAST.groupTree).not.toHaveProperty('children');
      expect(parsedAST.groupTree.operations).toHaveLength(3);
      expect(parsedAST.groupTree.operations[0].operationID).toBe('getPets');
      expect(parsedAST.groupTree.operations[1].operationID).toBe('addPet');
      expect(parsedAST.groupTree.operations[2].operationID).toBe('getPetById');
    })
    it('should omit children if there are none', async () => {
      // create a schema with one top operation and no children
      const schema = `
        openapi: 3.0.0
        info:
          title: Petstore API
        paths:
          /pets:
            get:
              summary: Get all pets
      `
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
      expect(parsedAST.groupTree).not.toHaveProperty('children');
    })
    it('should omit operations if there are none', async () => {
      // create a schema with one top operation and no children
      const schema = `
        openapi: 3.0.0
        info:
          title: Petstore API
      `
      const serializedAST = await SerializeSandboxAST(schema);
      const parsedAST = JSON.parse(serializedAST);
      expect(parsedAST.groupTree).toBeDefined();
      expect(parsedAST.groupTree).not.toHaveProperty('operations');
    })
  })
});
