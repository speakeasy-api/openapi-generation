import { Enum } from "../sdk/models/shared/enum.js";
import {
  SimpleObject,
  Int32Enum,
  IntEnum,
} from "../sdk/models/shared/simpleobject.js";
import { RFCDate } from "../sdk/types/rfcdate.js";

export const createSimpleObject = (): SimpleObject => {
  return {
    str: "test",
    bool: true,
    int: 1,
    int32: 1,
    int32_enum: Int32Enum.FiftyFive,
    int_enum: IntEnum.Second,
    num: 1.1,
    float32: 1.1,
    enum: Enum.One,
    any: "any",
    date: new RFCDate("2020-01-01"),
    date_time: new Date("2020-01-01T00:00:00.001Z"),
    bool_opt: true,
    str_opt: "testOptional",
  };
};
