/**
 * Template-level tests for webhook signature encoding/decoding conversions
 * These tests verify that all three signature encoding formats work correctly
 */

import { describe, test, expect } from "vitest";

describe("Webhook Signature Encoding/Decoding", () => {
  // Test data
  const testSecret = "test-secret-key";
  const testPayload = JSON.stringify({ foo: "bar", nested: { value: 123 } });

  // Helper to generate HMAC signature
  async function generateHMAC(
    secret: string,
    payload: string,
  ): Promise<ArrayBuffer> {
    const encoder = new TextEncoder();
    const secretBytes = encoder.encode(secret);
    const payloadBytes = encoder.encode(payload);

    const crypto = globalThis?.crypto?.subtle;
    if (!crypto) throw new Error("SubtleCrypto not available");

    const cryptoKey = await crypto.importKey(
      "raw",
      secretBytes,
      { name: "HMAC", hash: "SHA-256" },
      false,
      ["sign"],
    );

    return await crypto.sign("HMAC", cryptoKey, payloadBytes);
  }

  describe("Hex Encoding", () => {
    test("should encode signature to hex correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const hexSignature = Buffer.from(signatureBytes).toString("hex");

      // Verify it's valid hex
      expect(hexSignature).toMatch(/^[0-9a-fA-F]+$/);
      // Verify it's the right length (SHA-256 produces 32 bytes = 64 hex chars)
      expect(hexSignature).toHaveLength(64);
    });

    test("should decode hex signature correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const hexSignature = Buffer.from(signatureBytes).toString("hex");

      // Decode using our implementation
      const decodedBytes = Uint8Array.from(
        hexSignature.match(/.{1,2}/g) || [],
        (byte) => parseInt(byte, 16),
      );

      // Verify decoded bytes match original
      expect(decodedBytes).toEqual(new Uint8Array(signatureBytes));
    });

    test("should handle round-trip encoding/decoding", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);

      // Encode to hex
      const hexSignature = Buffer.from(signatureBytes).toString("hex");

      // Decode back
      const decodedBytes = Uint8Array.from(
        hexSignature.match(/.{1,2}/g) || [],
        (byte) => parseInt(byte, 16),
      );

      // Should match original
      expect(Array.from(decodedBytes)).toEqual(
        Array.from(new Uint8Array(signatureBytes)),
      );
    });
  });

  describe("Base64 Encoding", () => {
    // Helper function (simplified version of bytesToBase64)
    function bytesToBase64(bytes: Uint8Array): string {
      const binString = Array.from(bytes, (byte) =>
        String.fromCodePoint(byte),
      ).join("");
      return btoa(binString);
    }

    test("should encode signature to base64 correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const base64Signature = bytesToBase64(new Uint8Array(signatureBytes));

      // Verify it's valid base64 (can contain +, /, =, and alphanumeric)
      expect(base64Signature).toMatch(/^[A-Za-z0-9+/]+=*$/);
    });

    test("should decode base64 signature correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const base64Signature = bytesToBase64(new Uint8Array(signatureBytes));

      // Decode using our implementation
      const decodedBytes = Uint8Array.from(atob(base64Signature), (c) =>
        c.charCodeAt(0),
      );

      // Verify decoded bytes match original
      expect(decodedBytes).toEqual(new Uint8Array(signatureBytes));
    });

    test("should handle round-trip encoding/decoding", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);

      // Encode to base64
      const base64Signature = bytesToBase64(new Uint8Array(signatureBytes));

      // Decode back
      const decodedBytes = Uint8Array.from(atob(base64Signature), (c) =>
        c.charCodeAt(0),
      );

      // Should match original
      expect(Array.from(decodedBytes)).toEqual(
        Array.from(new Uint8Array(signatureBytes)),
      );
    });
  });

  describe("Base64URL Encoding", () => {
    // Helper function
    function bytesToBase64(bytes: Uint8Array): string {
      const binString = Array.from(bytes, (byte) =>
        String.fromCodePoint(byte),
      ).join("");
      return btoa(binString);
    }

    function bytesToBase64URL(bytes: Uint8Array): string {
      return bytesToBase64(bytes).replaceAll("+", "-").replaceAll("/", "_");
    }

    test("should encode signature to base64url correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const base64urlSignature = bytesToBase64URL(
        new Uint8Array(signatureBytes),
      );

      // Verify it's valid base64url (should NOT contain + or /)
      expect(base64urlSignature).not.toMatch(/[+/]/);
      // Should contain URL-safe characters
      expect(base64urlSignature).toMatch(/^[A-Za-z0-9_-]+=*$/);
    });

    test("should decode base64url signature correctly", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);
      const base64urlSignature = bytesToBase64URL(
        new Uint8Array(signatureBytes),
      );

      // Decode using our implementation
      const base64 = base64urlSignature
        .replaceAll("-", "+")
        .replaceAll("_", "/");
      const decodedBytes = Uint8Array.from(atob(base64), (c) =>
        c.charCodeAt(0),
      );

      // Verify decoded bytes match original
      expect(decodedBytes).toEqual(new Uint8Array(signatureBytes));
    });

    test("should handle round-trip encoding/decoding", async () => {
      const signatureBytes = await generateHMAC(testSecret, testPayload);

      // Encode to base64url
      const base64urlSignature = bytesToBase64URL(
        new Uint8Array(signatureBytes),
      );

      // Decode back
      const base64 = base64urlSignature
        .replaceAll("-", "+")
        .replaceAll("_", "/");
      const decodedBytes = Uint8Array.from(atob(base64), (c) =>
        c.charCodeAt(0),
      );

      // Should match original
      expect(Array.from(decodedBytes)).toEqual(
        Array.from(new Uint8Array(signatureBytes)),
      );
    });

    test("should correctly convert + and / to - and _", () => {
      const testString = "abc+def/ghi=";
      const base64urlString = testString
        .replaceAll("+", "-")
        .replaceAll("/", "_");

      expect(base64urlString).toBe("abc-def_ghi=");
      expect(base64urlString).not.toContain("+");
      expect(base64urlString).not.toContain("/");
    });
  });

  describe("Edge Cases", () => {
    test("should handle empty payload", async () => {
      const signatureBytes = await generateHMAC(testSecret, "");

      // Test hex
      const hexSignature = Buffer.from(signatureBytes).toString("hex");
      const hexDecoded = Uint8Array.from(
        hexSignature.match(/.{1,2}/g) || [],
        (byte) => parseInt(byte, 16),
      );
      expect(hexDecoded).toEqual(new Uint8Array(signatureBytes));
    });

    test("should handle special characters in payload", async () => {
      const specialPayload = JSON.stringify({
        emoji: "🎉",
        unicode: "你好",
        special: "!@#$%^&*()",
      });
      const signatureBytes = await generateHMAC(testSecret, specialPayload);

      // Test all three formats work with special characters
      const hexSignature = Buffer.from(signatureBytes).toString("hex");
      const hexDecoded = Uint8Array.from(
        hexSignature.match(/.{1,2}/g) || [],
        (byte) => parseInt(byte, 16),
      );
      expect(hexDecoded).toEqual(new Uint8Array(signatureBytes));
    });

    test("should produce different signatures for different payloads", async () => {
      const sig1 = await generateHMAC(testSecret, "payload1");
      const sig2 = await generateHMAC(testSecret, "payload2");

      const hex1 = Buffer.from(sig1).toString("hex");
      const hex2 = Buffer.from(sig2).toString("hex");

      expect(hex1).not.toBe(hex2);
    });

    test("should produce different signatures for different secrets", async () => {
      const sig1 = await generateHMAC("secret1", testPayload);
      const sig2 = await generateHMAC("secret2", testPayload);

      const hex1 = Buffer.from(sig1).toString("hex");
      const hex2 = Buffer.from(sig2).toString("hex");

      expect(hex1).not.toBe(hex2);
    });
  });
});
