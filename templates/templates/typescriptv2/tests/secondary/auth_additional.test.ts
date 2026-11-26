import { Buffer } from "node:buffer";
import { inspect } from "node:util";
import { expect, test } from "vitest";

import { BasicAuthUser } from "../sdk/models/operations/basicauth.js";
import { ErrorType1 } from "../sdk/models/errors/errortype1.js";
import { ErrorUnionPostRequestBody } from "../sdk/models/operations/errorunionpost.js";
import { SDK } from "../index.js";
import {
  HTTPBIN_URL,
  recordTest,
  RequestLogEntry,
  newRequestRecorderClient,
} from "./common_helpers.js";

test("Test No Auth", async () => {
  recordTest("auth-hoisted-no-auth-retained");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  await sdk.auth.noAuth();
});

test("Test Basic Auth", async () => {
  recordTest("auth-hoisted-basic-auth");

  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    security: {
      username: "testUser",
      password: "testPass",
    },
  });

  const res: BasicAuthUser = await sdk.auth.basicAuth("testPass", "testUser");

  expect(res?.authenticated).toBe(true);
});

test("Test Multiple Mixed Options Auth", async () => {
  recordTest("auth-hoisted-operation-auth-retained");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  await sdk.authNew.multipleMixedOptionsAuth(
    {
      basicAuth: {
        username: "testUser",
        password: "testPass",
      },
    },
    {
      basicAuth: {
        username: "testUser",
        password: "testPass",
      },
    },
  );
});

test("Test Authenticated Request Operation Level OAuth2", async () => {
  recordTest("auth-operation-level-oauth2");

  const log: RequestLogEntry[] = [];
  const httpClient = newRequestRecorderClient(log);
  const sdk = new SDK({ serverURL: HTTPBIN_URL, httpClient });

  // A token should be requested with 'read', 'write' and 'erase' scopes.
  await sdk.hooks.authenticatedRequest({
    clientID: "speakeasy-sdks",
    clientSecret: "supersecret-" + randSeq(10),
    audience: "",
  });

  // Requires 'read' and 'write' scopes. The same token should be reused
  // since [read, write, erase] is a superset of [read, write].
  await sdk.hooks.authenticatedRequestUnflattened({
    clientCredentials: {
      clientID: "speakeasy-sdks",
      clientSecret: "supersecret-" + randSeq(10),
      audience: "",
    },
  });

  // Check that only a single token request was made (token should be reused)
  const tokenRequests = log.filter(
    (entry) =>
      entry.url.includes("/clientcredentials/token") &&
      entry.body.includes("scope=read+write+erase"),
  );
  expect(tokenRequests).toHaveLength(1);
});

test("Test Typed Error Inspection Omits Credentials", async () => {
  recordTest("auth-typed-error-inspection-omits-credentials");

  const username = "testUser";
  const password = "secret-" + randSeq(10);
  const sdk = new SDK({
    serverURL: HTTPBIN_URL,
    security: { username, password },
  });

  const req: ErrorUnionPostRequestBody = { error: "Error1" };
  await sdk.errors.errorUnionPost(req).then(
    () => {
      expect.unreachable("sdk call is expected to fail");
    },
    (err) => {
      expect(err).toBeInstanceOf(ErrorType1);
      expect(err.rawResponse.status).toBe(500);

      const dump = inspect(err, { depth: Infinity });
      expect(dump).not.toContain(password);
      expect(dump).not.toContain(
        Buffer.from(`${username}:${password}`).toString("base64"),
      );
    },
  );
});

function randSeq(length: number) {
  const chars = "abcdefghijklmnopqrstuvwxyz";
  let result = "";
  for (let i = 0; i < length; i++) {
    result += chars.charAt(Math.floor(Math.random() * chars.length));
  }

  return result;
}
