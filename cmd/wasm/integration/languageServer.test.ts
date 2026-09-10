import { describe, it, expect, beforeAll } from "vitest";
import { getWasmFunction } from "./testHelpers";
import { JsonStream } from "./jsonStream";

interface LSPMessage {
  jsonrpc: string;
  method?: string;
  id?: number;
  params?: any;
  result?: any;
  error?: any;
}

interface PublishDiagnosticsParams {
  uri: string;
  diagnostics: Diagnostic[];
}

interface Diagnostic {
  range: Range;
  severity?: number;
  code?: string | number;
  source?: string;
  message: string;
}

interface Range {
  start: Position;
  end: Position;
}

interface Position {
  line: number;
  character: number;
}

/**
 * Recreates the interaction between the language server and Monaco
 * which enables us to test the language server within the vitest test runner.
 * At the moment, we just send a textDocument/didOpen message to the language server
 * and wait for the diagnostics to be published via textDocument/publishDiagnostics.
 */
class LSPTestHarness {
  private messageBuffers: string[] = [];
  private jsonStream: JsonStream;
  private resolveStdinReady: ((value: number) => void) | null = null;
  private messageHandlers: ((message: LSPMessage) => void)[] = [];
  private messageQueue: LSPMessage[] = [];

  constructor() {
    this.jsonStream = new JsonStream();
  }

  // Reader implementation
  asyncReadReady = async (): Promise<number> => {
    console.log('%c%s', 'color: magenta', 'asyncReadReady called, buffer length:', this.messageBuffers.length);
    if (this.messageBuffers.length === 0) {
      return new Promise<number>((resolve) => {
        console.log('%c%s', 'color: magenta', 'Waiting for messages...');
        this.resolveStdinReady = resolve;
      });
    }
    return this.messageBuffers.length;
  };

  syncRead = (): string[] => {
    console.log('%c%s', 'color: magenta', 'syncRead called, buffer length:', this.messageBuffers.length);
    const messages = [...this.messageBuffers];
    this.messageBuffers = [];
    return messages;
  };

  // Writer implementation
  notify = (message: string) => {
    console.log('%c%s', 'color: blue', 'Adding message to buffer:', message);
    // Add the message to the buffer for the language server to read
    this.messageBuffers.push(message);
    
    // If there's a pending read, resolve it
    if (this.resolveStdinReady) {
      console.log('%c%s', 'color: magenta', 'Resolving pending read with length:', this.messageBuffers.length);
      this.resolveStdinReady(this.messageBuffers.length);
      this.resolveStdinReady = null;
    }

    // Process any responses from the language server
    for (let i = 0; i < message.length; i++) {
      const jsonOrNull = this.jsonStream.insert(message.charCodeAt(i));
      if (jsonOrNull) {
        try {
          const parsedMessage: LSPMessage = JSON.parse(jsonOrNull);
          // Only process messages that aren't our own
          if (!this.isEcho(parsedMessage)) {
            console.log('%c%s', 'color: green', 'Received from server:', jsonOrNull);
            this.messageHandlers.forEach(handler => handler(parsedMessage));
          } else {
            console.log('%c%s', 'color: yellow', 'Ignoring echo message:', jsonOrNull);
          }
        } catch (e) {
          console.error('Failed to parse message:', e);
        }
      }
    }
  };

  private isEcho(message: LSPMessage): boolean {
    // Check if this message is an echo of one we sent
    return this.messageQueue.some(queued => {
      if (message.method && queued.method) {
        return message.method === queued.method;
      }
      if (message.id && queued.id) {
        return message.id === queued.id;
      }
      return false;
    });
  }

  onAbort = () => {
    console.error("LSP connection aborted");
  };

  // Helper methods
  private formatMessage(message: LSPMessage): string {
    const content = JSON.stringify(message);
    const header = `Content-Length: ${content.length}\r\n\r\n`;
    return header + content;
  }

  sendMessage(message: LSPMessage) {
    const formattedMessage = this.formatMessage(message);
    console.log('%c%s', 'color: blue', 'Sending to server:', formattedMessage);
    this.messageQueue.push(message);
    this.notify(formattedMessage);
  }

  onMessage(handler: (message: LSPMessage) => void) {
    this.messageHandlers.push(handler);
  }

  sendDidOpen(uri: string, content: string, languageId: string = "yaml") {
    this.sendMessage({
      jsonrpc: "2.0",
      method: "textDocument/didOpen",
      params: {
        textDocument: {
          uri,
          languageId,
          version: 1,
          text: content
        }
      }
    });
  }
}

describe("Language Server Integration", () => {
  let InitializeLS: (...args: any[]) => Promise<string>;
  let harness: LSPTestHarness;

  beforeAll(async () => {
    InitializeLS = await getWasmFunction("InitializeLS", false);
    harness = new LSPTestHarness();
  });

  it("should handle OpenAPI document", async () => {
    // Initialize the language server
    const result = InitializeLS({
      asyncReadReady: harness.asyncReadReady,
      syncRead: harness.syncRead,
      notify: harness.notify,
      onAbort: harness.onAbort,
    });

    console.log('%c%s', 'color: magenta', 'Language server initialized:', result);

    // Set up message handler with specific diagnostics handling
    let diagnosticsPromise: Promise<PublishDiagnosticsParams>;
    let diagnosticsResolve: (params: PublishDiagnosticsParams) => void;
    
    diagnosticsPromise = new Promise<PublishDiagnosticsParams>((resolve) => {
      diagnosticsResolve = resolve;
    });

    harness.onMessage((message) => {
      console.log('%c%s', 'color: cyan', 'Received message:', JSON.stringify(message, null, 2));
      if (message.method === 'textDocument/publishDiagnostics') {
        console.log('%c%s', 'color: yellow', 'Received diagnostics:', JSON.stringify(message.params, null, 2));
        diagnosticsResolve(message.params);
      }
    });

    // Send initialize request
    const initializeMessage = {
      jsonrpc: "2.0",
      id: 1,
      method: "initialize",
      params: {
        processId: null,
        rootUri: null,
        capabilities: {
          textDocument: {
            publishDiagnostics: {
              relatedInformation: true
            },
            synchronization: {
              didSave: true,
              willSave: true,
              willSaveWaitUntil: true
            }
          },
          workspace: {
            workspaceFolders: true
          }
        },
        trace: "verbose"
      }
    };

    console.log('%c%s', 'color: magenta', 'Sending initialize:', JSON.stringify(initializeMessage, null, 2));
    harness.sendMessage(initializeMessage);

    // Wait for initialize response
    await new Promise<void>((resolve) => {
      const timeout = setTimeout(() => {
        console.log('%c%s', 'color: red', 'Timeout waiting for initialize response');
        resolve();
      }, 2000);

      harness.onMessage((message) => {
        if (message.id === 1 && !message.method) {
          console.log('%c%s', 'color: green', 'Received initialize response:', JSON.stringify(message, null, 2));
          clearTimeout(timeout);
          resolve();
        }
      });
    });

    // Send initialized notification
    const initializedMessage = {
      jsonrpc: "2.0",
      method: "initialized",
      params: {}
    };

    console.log('%c%s', 'color: magenta', 'Sending initialized:', JSON.stringify(initializedMessage, null, 2));
    harness.sendMessage(initializedMessage);

    // Wait a bit for the server to process initialized
    await new Promise(resolve => setTimeout(resolve, 1000));

    // Sample OpenAPI spec in YAML format
    const openApiSpec = `openapi: "3.0.0"
info:
  title: "Sample API"
  version: "1.0.0"
paths:
  /users:
    get:
      summary: "Get users"
      operationId: "getUsers"
      responses:
        "200":
          description: "OK"`;

    // Send didOpen notification with proper LSP message format
    const didOpenMessage = {
      jsonrpc: "2.0",
      method: "textDocument/didOpen",
      params: {
        textDocument: {
          uri: "file:///sample-api.yaml",
          languageId: "yaml",
          version: 1,
          text: openApiSpec
        }
      }
    };

    console.log('%c%s', 'color: magenta', 'Sending didOpen:', JSON.stringify(didOpenMessage, null, 2));
    harness.sendMessage(didOpenMessage);

    // Wait for diagnostics with timeout
    let diagnostics: PublishDiagnosticsParams;
    try {
      diagnostics = await Promise.race([
        diagnosticsPromise,
        new Promise<PublishDiagnosticsParams>((_, reject) => setTimeout(() => reject(new Error('Timeout waiting for diagnostics')), 10000))
      ]);
    } catch (error) {
      console.error('Error waiting for diagnostics:', error);
      throw error;
    }

    expect(diagnostics.diagnostics).toContainEqual({
      code: "generator-validate-servers",
      codeDescription: { href: "https://www.speakeasy.com/docs/prep-openapi/linting#available-rules" },
      message: "no servers found in document, either add servers to the document or set a `baseServerUrl` in the `gen.yaml` config file",
      range: { end: { character: 4294967295, line: 0 }, start: { character: 0, line: 0 } },
      severity: 4,
      source: "speakeasy"
    });
  });
});
