import path = require("path");
import { promisify } from "util";
import { gunzip } from "zlib";
import { readFile } from "fs/promises";
import "../assets/wasm/wasm_exec.js";
import { load } from "js-yaml";

export const gunzipAsync = promisify(gunzip);

export async function getTestFileContents(fileName: string): Promise<string> {
    const extension = path.extname(fileName);
  const filePath = path.join(__dirname, fileName);
  let schema = await readFile(filePath, 'utf8');

  if (extension === '.yaml' || extension === '.yml') {
    const yaml = load(schema);
    schema = JSON.stringify(yaml);
  }

  return schema;
}


// Helper to promisify Go functions
export function promisifyGoFunction<T>(fn: (args: any[]) => T): (...args: any[]) => Promise<T> {
  return (...args: any[]) => {
    return new Promise((resolve, reject) => {
      try {
        const result = fn(args);
        resolve(result);
      } catch (error) {
        reject(error);
      }
    });
  };
}


/**
 * Get a WASM function from the global object. Optionally promisify it if the function is
 * async.
 */
export async function getWasmFunction(name: string, promisify: boolean = true) {
  const binaryName = getBinaryNameForFunction(name);
  const gzippedBuffer = await readFile(path.join(__dirname, `../assets/wasm/${binaryName}.wasm.gz`));
    
  // Decompress the gzipped buffer
  const wasmBuffer = await gunzipAsync(gzippedBuffer);
  
  // Instantiate the WASM module
  const go = new Go();
  const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);
  
  const instance = (result as unknown as { instance: WebAssembly.Instance }).instance;
  
  go.run(instance);

  // Get the global functions
  const global = globalThis as unknown as { [key: string]: any };
  return promisify ? promisifyGoFunction<string>(global[name]) : global[name];
}

/**
 * Map function names to their corresponding WASM binary
 */
function getBinaryNameForFunction(functionName: string): string {
  switch (functionName) {
    // Lightweight specialized binaries
    case 'CalculateOverlay':
    case 'ApplyOverlay':
      return 'overlay';
    
    case 'SerializeSandboxAST':
      return 'ast';
    
    // Heavy functionality in full-stack binary
    case 'GenerateUsageSnippets':
    case 'SerializeDocsData':
    case 'InitializeLS':
    case 'Healthcheck':
      return 'full-stack';
    
    // Fallback (shouldn't be needed)
    default:
      return 'full-stack';
  }
}

