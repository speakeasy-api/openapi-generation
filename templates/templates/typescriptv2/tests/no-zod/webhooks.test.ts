import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { WebhookAuthenticationError } from "../sdk/types/webhooks.js";
import { WebhookSecurity } from "../hooks/webhook-security.js";
import { recordTest, HTTPBIN_URL } from "./common_helpers.js";

const consumerURL = `${HTTPBIN_URL}/example-consumer-endpoint/`;
const secret = "secret";

test("Test Webhooks Consume", async () => {
  recordTest("webhooks-consume");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const request = new Request(consumerURL, {
    method: "POST",
    body: `{"data":{"bool":true,"date":"2020-01-01","dateTime":"2020-01-01T00:00:00.001Z","enum":"one","float32":1.1,"int":1,"int32":1,"int32Enum":55,"intEnum":2,"num":1.1,"str":"test","any":"any","boolOpt":true,"strOpt":"testOptional"},"type":"webhook.created"}`,
    headers: { "X-Signature": "AYkwql5cQI4Gb07HK4o8uLo1qNthUzSxDD0Acq6Iznk=" },
  });

  const res = await sdk.validateWebhook({
    request: request,
    secret,
  });

  expect(res).toBeDefined();
  expect(res.type).toBe("webhook.created");
});

test("Test Webhooks Consume Bad Signature", async () => {
  recordTest("webhooks-consume-bad-signature");

  const sdk = new SDK({ serverURL: HTTPBIN_URL });

  const req = new Request(consumerURL, {
    method: "POST",
    body: JSON.stringify({ foo: "foo" }),
    headers: { "X-Signature": "<BAD>" },
  });

  try {
    await sdk.validateWebhook({
      request: req,
      secret,
    });
  } catch (e: any) {
    expect(e).toBeInstanceOf(WebhookAuthenticationError);
  }
});

// No "bad data" test in no-zod: the webhook handler does no schema
// validation, so a syntactically-valid JSON body that doesn't match the
// declared shape is silently accepted (cast through). Caller is expected
// to perform its own shape checks if needed.

test("Test Authenticator", async () => {
  const security = new WebhookSecurity();
  const signedRequest = await security.sign({
    secret,
    request: new Request(consumerURL, {
      body: JSON.stringify({ foo: "foo" }),
      method: "POST",
    }),
  });

  // should be hex
  expect(signedRequest.headers.get("X-Signature")).toMatch(/[0-9a-fA-F]+/);

  expect(signedRequest.headers.get("X-Signature")).toBeDefined();

  const verified = await security.verify({
    secret,
    request: signedRequest,
  });

  expect(verified).toBe(true);
});
