import { describe, expect, test } from "vitest";

import {
  Decimal,
  decimal,
  decimalConst,
  decimalNullable,
  decimalOptional,
  decimalStr,
  decimalStrConst,
  decimalStrNullable,
  decimalStrOptional,
} from "../types/decimal.js";

describe("decimal()", () => {
  test("parses number to Decimal", () => {
    const result = decimal().parse(123.456);
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toNumber()).toBe(123.456);
  });

  test("passes through Decimal unchanged", () => {
    const input = new Decimal("9.87654321");
    const result = decimal().parse(input);
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toString()).toBe("9.87654321");
  });

  test("rejects null", () => {
    const result = decimal().safeParse(null);
    expect(result.success).toBe(false);
  });

  test("rejects undefined", () => {
    const result = decimal().safeParse(undefined);
    expect(result.success).toBe(false);
  });

  test("rejects invalid input", () => {
    const result = decimal().safeParse("not-a-number");
    expect(result.success).toBe(false);
  });
});

describe("decimalOptional()", () => {
  test("parses number to Decimal", () => {
    const result = decimalOptional().parse(456.789);
    if (result === undefined) throw new Error("Expected result to be defined");
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toNumber()).toBe(456.789);
  });

  test("passes through Decimal unchanged", () => {
    const input = new Decimal("1.23456789");
    const result = decimalOptional().parse(input);
    if (result === undefined) throw new Error("Expected result to be defined");
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toString()).toBe("1.23456789");
  });

  test("handles undefined by preserving undefined", () => {
    const result = decimalOptional().parse(undefined);
    expect(result).toBe(undefined);
  });

  test("handles null by coercing to undefined", () => {
    const result = decimalOptional().parse(null);
    expect(result).toBe(undefined);
  });

  test("rejects invalid string", () => {
    const result = decimalOptional().safeParse("invalid");
    expect(result.success).toBe(false);
  });

  test("full round-trip with undefined value", () => {
    const original = undefined;
    const serialized = decimalOptional().parse(original);
    expect(serialized).toBe(undefined);
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = decimalOptional().parse(parsed.value);
    expect(deserialized).toBe(undefined);
  });
});

describe("decimalNullable()", () => {
  test("parses number to Decimal", () => {
    const result = decimalNullable().parse(789.012);
    if (result === null) throw new Error("Expected result to be non-null");
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toNumber()).toBe(789.012);
  });

  test("passes through Decimal unchanged", () => {
    const input = new Decimal("0.123456");
    const result = decimalNullable().parse(input);
    if (result === null) throw new Error("Expected result to be non-null");
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toString()).toBe("0.123456");
  });

  test("handles null by preserving null", () => {
    const result = decimalNullable().parse(null);
    expect(result).toBe(null);
  });

  test("handles undefined by coercing to null", () => {
    const result = decimalNullable().parse(undefined);
    expect(result).toBe(null);
  });

  test("rejects invalid string", () => {
    const result = decimalNullable().safeParse("xyz");
    expect(result.success).toBe(false);
  });

  test("full round-trip with null value", () => {
    const original = null;
    const serialized = decimalNullable().parse(original);
    expect(serialized).toBe(null);
    const jsonString = JSON.stringify({ value: serialized });
    const parsed = JSON.parse(jsonString);
    const deserialized = decimalNullable().parse(parsed.value);
    expect(deserialized).toBe(null);
  });
});

describe("decimalStr()", () => {
  test("parses string to Decimal", () => {
    const result = decimalStr().parse("3.141592653589793");
    expect(result).toBeInstanceOf(Decimal);
    expect(result.toString()).toBe("3.141592653589793");
  });

  test("serializes Decimal to string", () => {
    const result = decimalStr().parse(new Decimal("9.87654321"));
    expect(result).toBe("9.87654321");
    expect(typeof result).toBe("string");
  });

  test("rejects number input", () => {
    const result = decimalStr().safeParse(123.456);
    expect(result.success).toBe(false);
  });
});

describe("decimalStrOptional()", () => {
  test("parses string to Decimal", () => {
    const result = decimalStrOptional().parse("2.718281828");
    expect(result).toBeDefined();
    expect(result).toBeInstanceOf(Decimal);
    expect((result as Decimal).toString()).toBe("2.718281828");
  });

  test("handles undefined", () => {
    const result = decimalStrOptional().parse(undefined);
    expect(result).toBe(undefined);
  });
});

describe("decimalStrNullable()", () => {
  test("parses string to Decimal", () => {
    const result = decimalStrNullable().parse("1.414213562");
    expect(result).not.toBeNull();
    expect(result).toBeInstanceOf(Decimal);
    expect((result as Decimal).toString()).toBe("1.414213562");
  });

  test("handles null", () => {
    const result = decimalStrNullable().parse(null);
    expect(result).toBe(null);
  });
});

describe("decimalConst()", () => {
  const constValue = new Decimal("3.141592653589793");
  const schema = decimalConst(constValue);

  test("accepts matching Decimal value", () => {
    const result = schema.parse(new Decimal("3.141592653589793"));
    expect(result).toBe(constValue);
    expect(result).toBeInstanceOf(Decimal);
  });

  test("accepts matching number value", () => {
    const result = schema.parse(3.141592653589793);
    expect(result).toBe(constValue);
    expect(result).toBeInstanceOf(Decimal);
  });

  test("accepts matching string value", () => {
    const result = schema.parse("3.141592653589793");
    expect(result).toBe(constValue);
    expect(result).toBeInstanceOf(Decimal);
  });

  test("rejects non-matching Decimal value", () => {
    const result = schema.safeParse(new Decimal("2.718281828"));
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 3.141592653589793",
      );
    }
  });

  test("rejects non-matching number value", () => {
    const result = schema.safeParse(1.234);
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 3.141592653589793",
      );
    }
  });

  test("rejects non-matching string value", () => {
    const result = schema.safeParse("9.876");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 3.141592653589793",
      );
    }
  });
});

describe("decimalStrConst()", () => {
  const constValue = new Decimal("3.141592653589793238462643383279");
  const schema = decimalStrConst(constValue);

  test("accepts matching Decimal value", () => {
    const result = schema.parse(
      new Decimal("3.141592653589793238462643383279"),
    );
    expect(result).toBe(constValue);
    expect(result).toBeInstanceOf(Decimal);
  });

  test("accepts matching string value", () => {
    const result = schema.parse("3.141592653589793238462643383279");
    expect(result).toBe(constValue);
    expect(result).toBeInstanceOf(Decimal);
  });

  test("rejects non-matching Decimal value", () => {
    const result = schema.safeParse(new Decimal("1.414213562"));
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 3.141592653589793238462643383279",
      );
    }
  });

  test("rejects non-matching string value", () => {
    const result = schema.safeParse("2.718281828");
    expect(result.success).toBe(false);
    if (!result.success) {
      expect(result.error.issues[0]?.message).toBe(
        "Value must be 3.141592653589793238462643383279",
      );
    }
  });
});
