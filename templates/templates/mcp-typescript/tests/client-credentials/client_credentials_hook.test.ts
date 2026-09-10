import { expect, test } from "vitest";

import { SDKCore } from "../core.js";
import { recordTest, randSeq, API_TEST_SERVICE_URL } from "./common_helpers.js";
import { HTTPClient } from "../lib/http.js";
import { hooksAuthenticatedRequest } from "../funcs/hooksAuthenticatedRequest.js";
import { hooksAuthenticatedRequestGlobalServer } from "../funcs/hooksAuthenticatedRequestGlobalServer.js";
import { hooksAuthenticatedRequestNoScopes } from "../funcs/hooksAuthenticatedRequestNoScopes.js";
import { ClientCredentialsOAuth2Scope } from "../hooks/oauth2scopes.js";

test("test client credentials hook successfully authenticates", async () => {
  recordTest("hooks-client-credentials-success");

  const sdk = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    security: {
      clientID: "speakeasy-sdks",
      clientSecret: "supersecret-" + randSeq(10),
    },
  });

  let res = await hooksAuthenticatedRequest(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);

  res = await hooksAuthenticatedRequest(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);
});

test("test client credentials hook successfully authenticates with global server", async () => {
  recordTest("hooks-client-credentials-success-global-server");

  const sdk = new SDKCore({
    security: {
      clientID: "speakeasy-sdks",
      clientSecret: "supersecret-" + randSeq(10),
    },
  });

  let res = await hooksAuthenticatedRequestGlobalServer(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);

  res = await hooksAuthenticatedRequestGlobalServer(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);
});

test("test client credentials hook successfully authenticates with alt token url", async () => {
  recordTest("hooks-client-credentials-success-alt-token-url");

  const requestLog: any[] = [];
  const client = new HTTPClient().addHook(
    "beforeRequest",
    async (req: Request) => {
      const body = req.body ? await req.clone().text() : "";
      requestLog.push({
        method: req.method,
        url: req.url,
        body: body,
      });
      return req;
    },
  );

  const tokenURL = "/clientcredentials/alt/token";
  const scopes: ClientCredentialsOAuth2Scope[] = ["alt:one", "alt:two"];

  const sdk = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    httpClient: client,
    security: {
      clientID: "speakeasy-sdks",
      clientSecret: "supersecret-" + randSeq(10),
      tokenURL: tokenURL,
      scopes: scopes,
    },
  });

  // 1. Initial token request (should succeed since alt tokenURL expects "alt:one" + "alt:two" scopes)
  let res = await hooksAuthenticatedRequest(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);

  // 2. Since the token is already expired, a new one should be requested
  res = await hooksAuthenticatedRequest(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);

  // Expecting 2 token requests to have been made with the test client
  const tokenRequests = requestLog.filter(
    (entry) =>
      entry.url.includes(tokenURL) &&
      entry.body.includes("scope=alt%3Aone+alt%3Atwo"),
  );
  expect(tokenRequests).toHaveLength(2);
});

test("test client credentials hook no scopes", async () => {
  recordTest("hooks-client-credentials-no-scopes");

  const clientID = "speakeasy-sdks";
  const clientSecret = "supersecret-" + randSeq(10);

  const sdk = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    security: {
      clientID: clientID,
      clientSecret: clientSecret,
    },
  });

  // expected to fail since the token endpoint requires 'read' and 'write' scopes
  // but the authenticatedRequestNoScopes operation does not specify any.
  try {
    const res = await hooksAuthenticatedRequestNoScopes(sdk, {});
    if (!res.ok) throw res.error;
    expect.fail("Expected an error to be thrown");
  } catch (error) {
    expect(error).toBeDefined();
    expect(String(error)).toContain(
      "Error: Received unexpected status code 400 while fetching token: empty_scopes",
    );
  }

  // same check but this time we override the default scopes with an empty list
  const sdk2 = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    security: {
      clientID: clientID,
      clientSecret: clientSecret,
      scopes: [], // overrides global scopes
    },
  });

  try {
    const res2 = await hooksAuthenticatedRequest(sdk2, {});
    if (!res2.ok) throw res2.error;
  } catch (error2) {
    expect(String(error2)).toContain(
      "Error: Received unexpected status code 400 while fetching token: empty_scopes",
    );
  }

  // now use a different tokenUrl that will allow no scopes to be requested
  const tokenURL = "/clientcredentials/token?expires_in=90&skip_scopes=true";
  const requestLog: any[] = [];
  const client3 = new HTTPClient().addHook(
    "beforeRequest",
    async (req: Request) => {
      const body = req.body ? await req.clone().text() : "";
      requestLog.push({
        method: req.method,
        url: req.url,
        body: body,
      });
      return req;
    },
  );

  const sdk3 = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    httpClient: client3,
    security: {
      clientID: clientID,
      clientSecret: clientSecret,
      tokenURL: tokenURL,
    },
  });

  const res3 = await hooksAuthenticatedRequestNoScopes(sdk3, {});
  if (!res3.ok) throw res3.error;
  expect(res3.value.StatusCode).toBe(200);

  // since the token is not expired, it should be reused on subsequent call
  const res4 = await hooksAuthenticatedRequestNoScopes(sdk3, {});
  if (!res4.ok) throw res4.error;
  expect(res4.value.StatusCode).toBe(200);

  const tokenRequests = requestLog.filter((entry) =>
    entry.url.includes(tokenURL),
  );
  expect(tokenRequests).toHaveLength(1);
  expect(tokenRequests[0].body).not.toContain("scope=");
});

test("test client credentials hook lowercase bearer token", async () => {
  recordTest("hooks-client-credentials-lowercase-bearer");

  const sdk = new SDKCore({
    serverURL: API_TEST_SERVICE_URL,
    security: {
      clientID: "speakeasy-sdks",
      clientSecret: "supersecret-" + randSeq(10),
      tokenURL: "/clientcredentials/token?token_type=bearer",
    },
  });

  const res = await hooksAuthenticatedRequest(sdk, {});
  if (!res.ok) throw res.error;
  expect(res.value.StatusCode).toBe(200);
});
