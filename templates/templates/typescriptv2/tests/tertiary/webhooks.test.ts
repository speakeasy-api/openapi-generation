import { expect, test } from "vitest";
import { WebhookSecurity } from "../hooks/webhook-security.js";
import { HTTPBIN_URL } from "./common_helpers.js";

const consumerURL = `${HTTPBIN_URL}/example-consumer-endpoint/`;
const secret = "secret";

test("Test Authenticator", async () => {
  const security = new WebhookSecurity();
  const signedRequest = await security.sign({
    secret,
    request: new Request(consumerURL, {
      body: JSON.stringify({ foo: "foo" }),
      method: "POST",
    }),
  });

  // should be base64url
  // should have non-hex characters
  expect(signedRequest.headers.get("X-Signature")).toMatch(/[f-zF-Z]/);
  // should not match the + or / of base64
  expect(signedRequest.headers.get("X-Signature")).not.toMatch(/\+|\//);

  expect(signedRequest.headers.get("X-Signature")).toBeDefined();

  const verified = await security.verify({
    secret,
    request: signedRequest,
  });

  expect(verified).toBe(true);
});
