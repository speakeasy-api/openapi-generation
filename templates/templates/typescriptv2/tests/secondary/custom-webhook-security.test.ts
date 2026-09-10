import { expect, test } from "vitest";

import { SDK } from "../index.js";
import { recordTest, HTTPBIN_URL } from "./common_helpers.js";

const consumerURL = `${HTTPBIN_URL}/example-consumer-endpoint/`;
const secret = "secret";

test("Test Webhooks Consume Custom Security", async () => {
  recordTest("webhooks-consume-custom-security");

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
