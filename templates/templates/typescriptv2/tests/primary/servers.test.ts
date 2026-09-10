import { expect, test } from "vitest";

import { SDK, ServerSomething } from "../index.js";
import {
  SelectServerWithIDServerBroken,
  SelectServerWithIDServerList,
} from "../sdk/models/operations/selectserverwithid.js";
import { ServerWithTemplatesServerList } from "../sdk/models/operations/serverwithtemplates.js";
import { pathToFunc } from "../lib/url.js";
import { recordTest, HTTPBIN_URL } from "./common_helpers.js";
import {
  ConnectionError,
  RequestTimeoutError,
} from "../sdk/models/errors/httpclienterrors.js";

test("Test Select Global Server Valid", async () => {
  recordTest("servers-select-global-server-valid");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });
  const res = await sdk.servers.selectGlobalServer();
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Select Global Server Broken", async () => {
  recordTest("servers-select-global-server-broken");

  const sdk = new SDK({ serverIdx: 1 });

  try {
    await sdk.servers.selectGlobalServer({ timeoutMs: 1000 });
    expect.unreachable("expected error to be thrown");
  } catch (err: unknown) {
    switch (true) {
      case err instanceof ConnectionError:
      case err instanceof RequestTimeoutError:
        break;
      default:
        expect.fail(
          "expected error to be ConnectionError or RequestTimeoutError",
        );
    }
  }
});

test("Test Select Server With ID Default", async () => {
  recordTest("servers-select-server-with-id-default");

  const sdk = new SDK();
  const res = await sdk.servers.selectServerWithID({ serverURL: HTTPBIN_URL });
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Select Server With ID Valid", async () => {
  recordTest("servers-select-server-with-id-valid");

  const sdk = new SDK({ serverURL: "http://broken" }); // broken server overridden by operation
  const res = await sdk.servers.selectServerWithID({
    serverURL: HTTPBIN_URL,
  });
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Select Server With ID Broken", async () => {
  recordTest("servers-select-server-with-id-broken");

  const sdk = new SDK();

  try {
    await sdk.servers.selectServerWithID({
      serverURL: SelectServerWithIDServerList[SelectServerWithIDServerBroken],
      timeoutMs: 1000,
    });
    expect.unreachable("Expected error to be thrown");
  } catch (err: unknown) {
    switch (true) {
      case err instanceof ConnectionError:
      case err instanceof RequestTimeoutError:
        break;
      default:
        expect.fail(
          "expected error to be ConnectionError or RequestTimeoutError",
        );
    }
  }
});

test("Test Server with Templates Global", async () => {
  recordTest("servers-server-with-templates-global");

  const sdk = new SDK({
    serverIdx: 2,
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.serverWithTemplatesGlobal();
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Server with Templates Global Defaults", async () => {
  recordTest("servers-server-with-templates-global-defaults");

  const sdk = new SDK({
    serverIdx: 2,
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.serverWithTemplatesGlobal();
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Server with Templates Global Enum", async () => {
  recordTest("servers-server-with-templates-global-enum");

  const sdk = new SDK({
    serverURL: `${HTTPBIN_URL}/anything/${ServerSomething.SomethingElseAgain}`,
  });

  const res = await sdk.servers.serverWithTemplatesGlobal();
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Server with Templates", async () => {
  recordTest("servers-server-with-templates");

  const sdk = new SDK();

  const serverURL = pathToFunc(ServerWithTemplatesServerList[0])({
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.serverWithTemplates({ serverURL });

  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Server with Templates Defaults", async () => {
  recordTest("servers-server-with-templates-defaults");

  const sdk = new SDK();

  const serverURL = pathToFunc(ServerWithTemplatesServerList[0])({
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.serverWithTemplates({ serverURL });

  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Server by ID with Templates", async () => {
  recordTest("servers-server-by-id-with-templates");

  const sdk = new SDK();

  const serverURL = pathToFunc("http://{hostname}:{port}")({
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.serversByIDWithTemplates({ serverURL });

  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Servers Global Server With Templated Protocol", async () => {
  recordTest("servers-global-server-with-templated-protocol");

  const sdk = new SDK({
    serverIdx: 4,
    protocol: "http",
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  const res = await sdk.servers.selectGlobalServer();

  expect(res.httpMeta.response.status).toBe(200);
  expect(res.httpMeta.response.url).toEqual(
    `${HTTPBIN_URL}/anything/selectGlobalServer`,
  );
});

test("Test Servers Global Server With Invalid Templated Protocol", async () => {
  recordTest("servers-global-server-with-invalid-templated-protocol");

  const sdk = new SDK({
    serverIdx: 4,
    protocol: "invalid",
    hostname: new URL(HTTPBIN_URL).hostname,
    port: new URL(HTTPBIN_URL).port,
  });

  try {
    await sdk.servers.selectGlobalServer();
    expect.unreachable("Expected error to be thrown");
  } catch (err) {
    expect(err).toBeInstanceOf(ConnectionError);
    expect(err).toHaveProperty(["cause", "cause", "message"], "unknown scheme");
  }
});

test("Test Servers Override Global Server URL", async () => {
  recordTest("servers-override-global-server-url");

  const sdk = new SDK();

  const res = await sdk.servers.serversOverrideGlobalServerURL({
    serverURL: `${HTTPBIN_URL}/anything`,
  });

  expect(res.httpMeta.request.url).toEqual(`${HTTPBIN_URL}/anything/ping`);
  expect(res.httpMeta.response.status).toBe(200);
});

test("Test Servers Override Operation Server URL", async () => {
  recordTest("servers-override-operation-server-url");

  const sdk = new SDK();

  const res = await sdk.servers.serversOverrideOperationServerURL({
    serverURL: `${HTTPBIN_URL}/anything`,
  });

  expect(res.httpMeta.request.url).toEqual(`${HTTPBIN_URL}/anything/ping`);
  expect(res.httpMeta.response.status).toBe(200);
});
