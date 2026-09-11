import { DeepObject } from "../sdk/models/shared/deepobject.js";
import { DeepObjectCamelCase } from "../sdk/models/shared/deepobjectcamelcase.js";
import { Enum } from "../sdk/models/shared/enum.js";
import {
  SimpleObject,
  SimpleObjectInt32Enum,
  SimpleObjectIntEnum,
} from "../sdk/models/shared/simpleobject.js";
import {
  SimpleObjectCamelCase,
  Int32EnumVal,
  IntEnumVal,
} from "../sdk/models/shared/simpleobjectcamelcase.js";
import {
  SimpleObjectWithType,
  SimpleObjectWithTypeIntEnum,
  SimpleObjectWithTypeInt32Enum,
} from "../sdk/models/shared/simpleobjectwithtype.js";
import { DeepObjectWithType } from "../sdk/models/shared/deepobjectwithtype.js";
export const createSimpleObject = (): SimpleObject => {
  return {
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32Enum: SimpleObjectInt32Enum.FiftyFive,
    intEnum: SimpleObjectIntEnum.Second,
    num: 1.1,
    float32: 1.1,
    enum: Enum.One,
    any: "any",
    date: new Date("2020-01-01"),
    dateTime: new Date("2020-01-01T00:00:00.001Z"),
    boolOpt: true,
    strOpt: "testOptional",
  };
};

export const createSimpleObjectCamelCase = (): SimpleObjectCamelCase => {
  return {
    str_val: "test",
    bool_val: true,
    int_val: 1,
    int32_val: 1,
    int32_enum_val: Int32EnumVal.FiftyFive,
    int_enum_val: IntEnumVal.Second,
    num_val: 1.1,
    float32_val: 1.1,
    enum_val: Enum.One,
    any_val: "any",
    date_val: new Date("2020-01-01"),
    date_time_val: new Date("2020-01-01T00:00:00.001Z"),
    bool_opt_val: true,
    str_opt_val: "testOptional",
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
    any_val: createSimpleObjectCamelCase(),
    arr_val: [createSimpleObjectCamelCase(), createSimpleObjectCamelCase()],
    bool_val: true,
    int_val: 1,
    map_val: {
      key: createSimpleObjectCamelCase(),
    },
    num_val: 1.1,
    obj_val: createSimpleObjectCamelCase(),
    str_val: "test",
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
    int32Enum: SimpleObjectWithTypeInt32Enum.FiftyFive,
    num: 1.1,
    float32: 1.1,
    enum: Enum.One,
    any: "any",
    date: new Date("2020-01-01"),
    dateTime: new Date("2020-01-01"),
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
