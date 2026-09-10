// No-zod test helpers.
//
// Adapted from tests/primary/primary_helpers.ts. The no-zod variant uses
// `preserveModelFieldNames: true`, so wire field names match the spec
// (camelCase like `dateTime`, `strOpt`). Date / date-time fields are ISO
// strings rather than Date / RFCDate instances, and decimal / bigint
// fields are plain numbers / strings.

import { DeepObject } from "../sdk/models/shared/deepobject.js";
import { DeepObjectCamelCase } from "../sdk/models/shared/deepobjectcamelcase.js";
import { IntEnum, SimpleObject } from "../sdk/models/shared/simpleobject.js";
import {
  IntEnumVal,
  SimpleObjectCamelCase,
} from "../sdk/models/shared/simpleobjectcamelcase.js";
import { DeepObjectWithType } from "../sdk/models/shared/deepobjectwithtype.js";
import {
  SimpleObjectWithType,
  SimpleObjectWithTypeIntEnum,
} from "../sdk/models/shared/simpleobjectwithtype.js";

export const createSimpleObject = (): SimpleObject => {
  return {
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32Enum: 55,
    intEnum: IntEnum.Second,
    num: 1.1,
    float32: 1.1,
    enum: "one",
    any: "any",
    date: "2020-01-01",
    dateTime: "2020-01-01T00:00:00.001Z",
    boolOpt: true,
    strOpt: "testOptional",
  };
};

export const createSimpleObjectCamelCase = (): SimpleObjectCamelCase => {
  return {
    strVal: "test",
    boolVal: true,
    intVal: 1,
    int32Val: 1,
    int32EnumVal: 55,
    intEnumVal: IntEnumVal.Second,
    numVal: 1.1,
    float32Val: 1.1,
    enumVal: "one",
    anyVal: "any",
    dateVal: "2020-01-01",
    dateTimeVal: "2020-01-01T00:00:00.001Z",
    boolOptVal: true,
    strOptVal: "testOptional",
  };
};

export const createDeepObject = (): DeepObject => {
  return {
    any: createSimpleObject(),
    arr: [createSimpleObject(), createSimpleObject()],
    bool: true,
    int: 1,
    map: {
      key: createSimpleObject(),
    },
    num: 1.1,
    obj: createSimpleObject(),
    str: "test",
  };
};

export const createDeepObjectCamelCase = (): DeepObjectCamelCase => {
  return {
    anyVal: createSimpleObjectCamelCase(),
    arrVal: [createSimpleObjectCamelCase(), createSimpleObjectCamelCase()],
    boolVal: true,
    intVal: 1,
    mapVal: {
      key: createSimpleObjectCamelCase(),
    },
    numVal: 1.1,
    objVal: createSimpleObjectCamelCase(),
    strVal: "test",
  };
};

export const createSimpleObjectWithType = (): SimpleObjectWithType & {
  type: "simpleObjectWithType";
} => {
  return {
    type: "simpleObjectWithType" as const,
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    intEnum: SimpleObjectWithTypeIntEnum.Second,
    int32Enum: 55,
    num: 1.1,
    float32: 1.1,
    enum: "one",
    any: "any",
    date: "2020-01-01",
    dateTime: "2020-01-01T00:00:00.000Z",
    boolOpt: true,
    strOpt: "testOptional",
  };
};

export const createDeepObjectWithType = (): DeepObjectWithType & {
  type: "deepObjectWithType";
} => {
  return {
    type: "deepObjectWithType" as const,
    any: createSimpleObject(),
    num: 1.1,
    bool: true,
    int: 1,
    str: "test",
    obj: createSimpleObject(),
    map: { key: createSimpleObject() },
    arr: [createSimpleObject(), createSimpleObject()],
  };
};

export const sortKeys = (obj: any): any => {
  if (Array.isArray(obj)) {
    return obj.map(sortKeys);
  } else if (obj && typeof obj === "object") {
    return Object.keys(obj)
      .sort()
      .reduce((result, key) => {
        if (key == "date" || key == "dateTime") {
          result[key] = obj[key];
          return result;
        }
        result[key] = sortKeys(obj[key]);
        return result;
      }, {} as any);
  } else {
    return obj;
  }
};
