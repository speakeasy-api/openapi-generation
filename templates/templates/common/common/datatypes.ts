enum DataTypeValue {
  String = "string",
  Integer = "integer",
  Int32 = "int32",
  Number = "number",
  Float32 = "float32",
  Decimal = "decimal",
  Boolean = "boolean",
  Date = "date",
  DateTime = "date-time",
  Map = "map",
  Array = "array",
  EventStream = "event-stream",
  JsonL = "jsonl",
  Any = "any",
  Bytes = "bytes",
  Class = "class",
  Enum = "enum",
  Request = "request",
  Union = "union",
  Response = "response",
  BigInt = "bigint",
  Error = "error",
}

function isBasicType(typeDef: TypeDef) {
  switch (typeDef.Type.toString()) {
    case DataTypeValue.Class:
    case DataTypeValue.Union:
    case DataTypeValue.Array:
    case DataTypeValue.Map:
    case DataTypeValue.EventStream:
    case DataTypeValue.JsonL:
    case DataTypeValue.Any:
    case DataTypeValue.Error:
      return false;
    case DataTypeValue.String:
    case DataTypeValue.DateTime:
    case DataTypeValue.Date:
    case DataTypeValue.Enum:
    case DataTypeValue.Integer:
    case DataTypeValue.Int32:
    case DataTypeValue.BigInt:
    case DataTypeValue.Number:
    case DataTypeValue.Float32:
    case DataTypeValue.Decimal:
    case DataTypeValue.Boolean:
    case DataTypeValue.Bytes:
    case DataTypeValue.Response:
      return true;
    default:
      throw new Error(`invalid type ${typeDef.Type.toString()}`);
  }
}

function isNumericType(typeDef: TypeDef) {
  switch (typeDef.Type.toString()) {
    case DataTypeValue.Integer:
    case DataTypeValue.Int32:
    case DataTypeValue.BigInt:
    case DataTypeValue.Number:
    case DataTypeValue.Float32:
    case DataTypeValue.Decimal:
      return true;
    case DataTypeValue.Class:
    case DataTypeValue.Union:
    case DataTypeValue.Array:
    case DataTypeValue.Map:
    case DataTypeValue.EventStream:
    case DataTypeValue.JsonL:
    case DataTypeValue.Any:
    case DataTypeValue.String:
    case DataTypeValue.DateTime:
    case DataTypeValue.Date:
    case DataTypeValue.Enum:
    case DataTypeValue.Boolean:
    case DataTypeValue.Bytes:
    case DataTypeValue.Response:
    case DataTypeValue.Error:
      return false;
    default:
      throw new Error(`invalid type ${typeDef.Type.toString()}`);
  }
}
