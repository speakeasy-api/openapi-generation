import { describe, expect, test } from "vitest";

import {
  bigint,
  bigintConst,
  bigintNullable,
  bigintOptional,
} from "../types/bigint.js";

describe("bigint()", () => {
  test("parses string to bigint", () => {
    const result = bigint().parse("12345");
    expect(result).toBe(BigInt(12345));
    expect(typeof result).toBe("bigint");
  });

  test("serializes bigint to string", () => {
    const result = bigint().parse(BigInt(67890));
    expect(result).toBe("67890");
    expect(typeof result).toBe("string");
  });

  test("handles null by coercing to BigInt(0)", () => {
    const result = bigint().parse(null);
    expect(result).toBe(BigInt(0));
    expect(typeof result).toBe("bigint");
  });

  test("handles undefined by coercing to BigInt(0)", () => {
    const result = bigint().parse(undefined);
    expect(result).toBe(BigInt(0));
    expect(typeof result).toBe("bigint");
  });

  test("rejects invalid string", () => {
    const result = bigint().safeParse("not-a-number");
    expect(result.success).toBe(false);
  });

  test("full round-trip preserves value", () => {
    const original = BigInt(99999);
    const serialized = bigint().parse(original);
    expect(typeof serialized).toBe("string");
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = bigint().parse(parsed.value);
    expect(deserialized).toBe(original);
    expect(typeof deserialized).toBe("bigint");
  });
});

describe("bigintOptional()", () => {
  test("parses string to bigint", () => {
    const result = bigintOptional().parse("12345");
    expect(result).toBe(BigInt(12345));
    expect(typeof result).toBe("bigint");
  });

  test("serializes bigint to string", () => {
    const result = bigintOptional().parse(BigInt(67890));
    expect(result).toBe("67890");
    expect(typeof result).toBe("string");
  });

  test("handles undefined by preserving undefined", () => {
    const result = bigintOptional().parse(undefined);
    expect(result).toBe(undefined);
  });

  test("handles null by coercing to undefined", () => {
    const result = bigintOptional().parse(null);
    expect(result).toBe(undefined);
  });

  test("rejects invalid string", () => {
    const result = bigintOptional().safeParse("not-a-number");
    expect(result.success).toBe(false);
  });

  test("full round-trip with defined value", () => {
    const original = BigInt(11111);
    const serialized = bigintOptional().parse(original);
    expect(typeof serialized).toBe("string");
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = bigintOptional().parse(parsed.value);
    expect(deserialized).toBe(original);
    expect(typeof deserialized).toBe("bigint");
  });

  test("full round-trip with undefined value", () => {
    const original = undefined;
    const serialized = bigintOptional().parse(original);
    expect(serialized).toBe(undefined);
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = bigintOptional().parse(parsed.value);
    expect(deserialized).toBe(undefined);
  });
});

describe("bigintNullable()", () => {
  test("parses string to bigint", () => {
    const result = bigintNullable().parse("12345");
    expect(result).toBe(BigInt(12345));
    expect(typeof result).toBe("bigint");
  });

  test("serializes bigint to string", () => {
    const result = bigintNullable().parse(BigInt(67890));
    expect(result).toBe("67890");
    expect(typeof result).toBe("string");
  });

  test("handles null by preserving null", () => {
    const result = bigintNullable().parse(null);
    expect(result).toBe(null);
  });

  test("handles undefined by coercing to null", () => {
    const result = bigintNullable().parse(undefined);
    expect(result).toBe(null);
  });

  test("rejects invalid string", () => {
    const result = bigintNullable().safeParse("not-a-number");
    expect(result.success).toBe(false);
  });

  test("full round-trip with defined value", () => {
    const original = BigInt(22222);
    const serialized = bigintNullable().parse(original);
    expect(typeof serialized).toBe("string");
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = bigintNullable().parse(parsed.value);
    expect(deserialized).toBe(original);
    expect(typeof deserialized).toBe("bigint");
  });

  test("full round-trip with null value", () => {
    const original = null;
    const serialized = bigintNullable().parse(original);
    expect(serialized).toBe(null);
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = bigintNullable().parse(parsed.value);
    expect(deserialized).toBe(null);
  });
});

describe("const bigint with default and refine", () => {
  test("refine validates using string comparison for bigint input", () => {
    const constSchema = bigint().refine(
      (v) => (typeof v === "string" ? v : v?.toString()) === "123",
      {
        message: "Value must be 123",
      },
    );

    const result = constSchema.parse(BigInt("123"));
    expect(result).toBe("123");
  });

  test("refine validates using string comparison for string input", () => {
    const constSchema = bigint().refine(
      (v) => (typeof v === "string" ? v : v?.toString()) === "123",
      {
        message: "Value must be 123",
      },
    );

    const result = constSchema.parse("123");
    expect(result).toBe(BigInt(123));
  });

  test("refine rejects non-matching bigint value", () => {
    const constSchema = bigint().refine(
      (v) => (typeof v === "string" ? v : v?.toString()) === "123",
      {
        message: "Value must be 123",
      },
    );

    const result = constSchema.safeParse(BigInt("456"));
    expect(result.success).toBe(false);
  });

  test("refine rejects non-matching string value", () => {
    const constSchema = bigint().refine(
      (v) => (typeof v === "string" ? v : v?.toString()) === "123",
      {
        message: "Value must be 123",
      },
    );

    const result = constSchema.safeParse("456");
    expect(result.success).toBe(false);
  });
});

describe("bigintConst()", () => {
  const constValue = BigInt(9007199254740991) as 9007199254740991n;
  const schema = bigintConst(constValue);

  test("accepts matching bigint value", () => {
    const result = schema.parse(BigInt("9007199254740991"));
    expect(result).toBe(constValue);
  });

  test("accepts matching string value", () => {
    const result = schema.parse("9007199254740991");
    expect(result).toBe(constValue);
  });

  test("accepts matching number value", () => {
    const result = schema.parse(9007199254740991);
    expect(result).toBe(constValue);
  });

  test("rejects non-matching bigint value", () => {
    const result = schema.safeParse(BigInt("123"));
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 9007199254740991",
      );
    }
  });

  test("rejects non-matching string value", () => {
    const result = schema.safeParse("456");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 9007199254740991",
      );
    }
  });

  test("rejects non-matching number value", () => {
    const result = schema.safeParse(789);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 9007199254740991",
      );
    }
  });
});
