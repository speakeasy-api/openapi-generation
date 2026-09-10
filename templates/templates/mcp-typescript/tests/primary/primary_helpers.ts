import { DeepObject } from "../models/deepobject.js";
import { SimpleObject } from "../models/simpleobject.js";

export const createSimpleObject = (): SimpleObject => {
  return {
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32Enum: 55,
    intEnum: 2,
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
