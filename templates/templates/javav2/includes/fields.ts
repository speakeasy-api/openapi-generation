type FieldMetadata = {
  fieldType: string;
  constructor: AccessorMetadata;
  getter: AccessorMetadata;
  setters: AccessorMetadata[];
  javaParam: JavaParam;
};

type AccessorMetadata = {
  paramType?: string;
  returnType?: string;
  body: string[];
};

// Helper types for better code organization
type FieldTypeInfo = {
  requiredType: string;
  boxedType: string;
  nullableType?: string;
  fieldType: string;
};

type FieldVariant =
  | "nullable-optional-default"
  | "nullable-optional"
  | "nullable-only"
  | "optional-only"
  | "required";

/**
 * Determines the field variant based on nullable, optional, and default properties
 */
function getFieldVariant(param: JavaParam): FieldVariant {
  if (param.Nullable) {
    if (param.Optional) {
      return param.Default ? "nullable-optional-default" : "nullable-optional";
    }
    return "nullable-only";
  }
  if (param.Optional) {
    return "optional-only";
  }
  return "required";
}

function canFieldBeNull(field: FieldDef | JavaParam): boolean {
  return Boolean(field.Nullable || field.Optional || field.Default);
}

/**
 * Extracts common type information to avoid repeated calculations
 */
function getFieldTypeInfo(param: JavaParam): FieldTypeInfo {
  const boxedType = toNonPrimitive(param.Type);
  const isRequired = !canFieldBeNull(param);
  let requiredType = isRequired ? param.Type : boxedType;
  let fieldType = requiredType;
  let nullableType = undefined;
  if (param.Nullable) {
    nullableType = `org.openapitools.jackson.nullable.JsonNullable<${boxedType}>`;
    fieldType = nullableType;
  }

  return {
    boxedType,
    requiredType,
    nullableType,
    fieldType,
  };
}

/**
 * Builds constructor metadata based on field variant and type info
 */
function buildConstructorMetadata(
  param: JavaParam,
  typeInfo: FieldTypeInfo,
  variant: FieldVariant,
): AccessorMetadata {
  const { requiredType, boxedType, nullableType } = typeInfo;

  switch (variant) {
    case "nullable-optional-default":
      return {
        paramType: `${templateNullableAnnotation(nullableType!)}`,
        body: [
          `${javaImportOptional()}.ofNullable(${param.Name})`,
          `.orElse(${javaImportJsonNullable()}.of(Builder.${singletonValueGet(
            param.Name,
          )}))`,
        ],
      };

    case "nullable-optional":
      return {
        paramType: `${templateNullableAnnotation(nullableType!)}`,
        body: [
          `${javaImportOptional()}.ofNullable(${param.Name})`,
          `.orElse(${javaImportJsonNullable()}.undefined())`,
        ],
      };

    case "nullable-only":
      return {
        paramType: `${templateNullableAnnotation(
          param.Default ? nullableType! : boxedType,
        )}`,
        body: param.Default
          ? [
              `${javaImportOptional()}.ofNullable(${param.Name})`,
              `.orElse(${javaImportJsonNullable()}.of(Builder.${singletonValueGet(
                param.Name,
              )}))`,
            ]
          : [`${javaImportJsonNullable()}.of(${param.Name})`],
      };

    case "optional-only":
      return {
        paramType: `${templateNullableAnnotation(boxedType)}`,
        body: param.Default
          ? [
              `${javaImportOptional()}.ofNullable(${param.Name})`,
              `.orElse(Builder.${singletonValueGet(param.Name)})`,
            ]
          : [`${param.Name}`],
      };

    case "required":
      if (isPrimitive(requiredType)) {
        return {
          paramType: requiredType,
          body: [`${param.Name}`],
        };
      }
      return {
        paramType: param.Default
          ? `${templateNullableAnnotation(requiredType)}`
          : `${templateNonNullAnnotation(requiredType)}`,
        body: [
          `${javaImportOptional()}.ofNullable(${param.Name})`,
          param.Default
            ? `.orElse(Builder.${singletonValueGet(param.Name)})`
            : `.orElseThrow(() -> new IllegalArgumentException("${param.Name} cannot be null"))`,
        ],
      };
  }
}

type GetterStyle = "presence-aware" | "raw" | "always-optional";

function isRawGetter(
  style: GetterStyle = context.Global.Config.GetterStyle,
): boolean {
  return style === "raw";
}
registerTemplateFunc("isRawGetter", isRawGetter);

function isAlwaysOptionalGetter(
  style: GetterStyle = context.Global.Config.GetterStyle,
): boolean {
  return style === "always-optional";
}
registerTemplateFunc("isAlwaysOptionalGetter", isAlwaysOptionalGetter);

function removeGetter(path: string): string {
  const ending = ".get()";
  if (path.endsWith(ending)) {
    return path.substring(0, path.lastIndexOf(ending));
  } else {
    return path;
  }
}

/**
 * Builds getter metadata based on field variant and type info
 */
function buildGetterMetadata(
  param: JavaParam,
  typeInfo: FieldTypeInfo,
  variant: FieldVariant,
  style: GetterStyle = context.Global.Config.GetterStyle,
): AccessorMetadata {
  const { requiredType, boxedType, nullableType } = typeInfo;

  switch (variant) {
    case "nullable-optional-default":
    case "nullable-optional":
    case "nullable-only":
      if (isRawGetter(style)) {
        return {
          returnType: templateNullableAnnotation(boxedType),
          body: [`this.${param.Name}.orElse(null)`],
        };
      }
      if (isAlwaysOptionalGetter(style)) {
        return {
          returnType: `${javaImportOptional()}<${boxedType}>`,
          body: [
            `${javaImportOptional()}.ofNullable(this.${
              param.Name
            }.orElse(null))`,
          ],
        };
      }
      return {
        returnType: nullableType!,
        body: [`this.${param.Name}`],
      };

    case "optional-only":
      if (isRawGetter(style)) {
        return {
          returnType: templateNullableAnnotation(boxedType),
          body: [`this.${param.Name}`],
        };
      }
      return {
        returnType: `${javaImportOptional()}<${boxedType}>`,
        body: [`${javaImportOptional()}.ofNullable(this.${param.Name})`],
      };

    case "required":
      if (isAlwaysOptionalGetter(style)) {
        return {
          returnType: `${javaImportOptional()}<${boxedType}>`,
          body: [`${javaImportOptional()}.ofNullable(this.${param.Name})`],
        };
      }
      return {
        returnType: requiredType,
        body: [`this.${param.Name}`],
      };
  }
}

/**
 * Builds setter metadata based on field variant and type info
 */
function buildSettersMetadata(
  param: JavaParam,
  typeInfo: FieldTypeInfo,
  variant: FieldVariant,
): AccessorMetadata[] {
  const { requiredType, boxedType } = typeInfo;

  switch (variant) {
    case "nullable-optional-default":
    case "nullable-optional":
    case "nullable-only":
      return buildNullableSetters(param, boxedType);

    case "optional-only":
      return buildOptionalOnlySetters(param, boxedType);

    case "required":
      return buildRequiredSetters(param, requiredType);
  }
}

function buildNullableSetters(
  param: JavaParam,
  paramType: string,
): AccessorMetadata[] {
  const setters: AccessorMetadata[] = [
    {
      paramType: `${templateNullableAnnotation(paramType)}`,
      body: [
        `this.${param.Name} = ${javaImportJsonNullable()}.of(${param.Name});`,
      ],
    },
  ];

  if (param.BigType) {
    if (parameterIsBigInteger(param)) {
      setters.push({
        paramType: "long",
        body: [
          `this.${param.Name} = ${javaImportJsonNullable()}.of(${javaImport(
            "java.math.BigInteger",
          )}.valueOf(${param.Name}));`,
        ],
      });
    } else if (parameterIsBigDecimal(param)) {
      setters.push({
        paramType: "double",
        body: [
          `this.${param.Name} = ${javaImportJsonNullable()}.of(${javaImport(
            "java.math.BigDecimal",
          )}.valueOf(${param.Name}));`,
        ],
      });
    }
  }

  return setters;
}

function buildOptionalOnlySetters(
  param: JavaParam,
  paramType: string,
): AccessorMetadata[] {
  const setters: AccessorMetadata[] = [
    {
      paramType: `${templateNullableAnnotation(paramType)}`,
      body: [`this.${param.Name} = ${param.Name};`],
    },
  ];

  if (param.BigType) {
    if (parameterIsBigInteger(param)) {
      setters.push({
        paramType: "long",
        body: [
          `this.${param.Name} = ${javaImport("java.math.BigInteger")}.valueOf(${
            param.Name
          });`,
        ],
      });
    } else if (parameterIsBigDecimal(param)) {
      setters.push({
        paramType: "double",
        body: [
          `this.${param.Name} = ${javaImport("java.math.BigDecimal")}.valueOf(${
            param.Name
          });`,
        ],
      });
    }
  }
  return setters;
}

function buildRequiredSetters(
  param: JavaParam,
  paramType: string,
): AccessorMetadata[] {
  if (isPrimitive(paramType)) {
    return [
      {
        paramType: paramType,
        body: [`this.${param.Name} = ${param.Name};`],
      },
    ];
  }

  const setters: AccessorMetadata[] = [
    {
      paramType: `${templateNonNullAnnotation(paramType)}`,
      body: [
        `this.${param.Name} = ${javaImportUtils()}.checkNotNull(${
          param.Name
        }, "${param.Name}");`,
      ],
    },
  ];

  if (param.BigType) {
    if (parameterIsBigInteger(param)) {
      setters.push({
        paramType: "long",
        body: [
          `this.${param.Name} = ${javaImport("java.math.BigInteger")}.valueOf(${
            param.Name
          });`,
        ],
      });
    } else if (parameterIsBigDecimal(param)) {
      setters.push({
        paramType: "double",
        body: [
          `this.${param.Name} = ${javaImport("java.math.BigDecimal")}.valueOf(${
            param.Name
          });`,
        ],
      });
    }
  }

  return setters;
}

function getSettersMetadata(param: JavaParam): AccessorMetadata[] {
  const typeInfo = getFieldTypeInfo(param);
  const variant = getFieldVariant(param);
  return buildSettersMetadata(param, typeInfo, variant);
}
registerTemplateFunc("getSettersMetadata", getSettersMetadata);

function getGetterMetadata(
  param: JavaParam,
  styleOverride?: GetterStyle,
): AccessorMetadata {
  const typeInfo = getFieldTypeInfo(param);
  const variant = getFieldVariant(param);
  return buildGetterMetadata(param, typeInfo, variant, styleOverride);
}
registerTemplateFunc("getGetterMetadata", getGetterMetadata);

/**
 * The effective getter style for a field:
 * "presence-aware" for generator-synthesized response envelope fields
 * (contentType/statusCode/rawResponse), whose getters implement the
 * utils.Response/AsyncResponse interface's fixed String/int/HttpResponse
 * signatures and so must not be rewritten; the configured getterStyle
 * otherwise.
 */
function fieldGetterStyle(field: FieldDef): GetterStyle {
  return field.IsResponseMetadata
    ? "presence-aware"
    : context.Global.Config.GetterStyle;
}
registerTemplateFunc("fieldGetterStyle", fieldGetterStyle);

/**
 * Returns the full return expression for a non-null-friendly model getter,
 * honoring getterStyle. Field storage is unchanged by getterStyle: required
 * fields store T, optional XOR nullable fields store Optional<T>, and
 * optional+nullable fields store JsonNullable<T>.
 */
function templateStyledGetterReturn(
  field: FieldDef,
  isAsync: boolean,
  style: GetterStyle = context.Global.Config.GetterStyle,
): string {
  const name = sanitizeFieldName(field.Name);
  const optional = field.Optional && !field.IsAdditionalProperties;
  const nullable = field.Nullable && !field.IsAdditionalProperties;

  if (isRawGetter(style) && (optional || nullable)) {
    return `${name}.orElse(null)`;
  }
  if (isAlwaysOptionalGetter(style)) {
    if (!optional && !nullable) {
      return `${javaImportOptional()}.ofNullable(${name})`;
    }
    if (optional && nullable) {
      return `${javaImportOptional()}.ofNullable(${name}.orElse(null))`;
    }
  }
  return `${castToRemoveWildcard(field, isAsync)}${name}`;
}
registerTemplateFunc("templateStyledGetterReturn", templateStyledGetterReturn);

/**
 * Whether the non-null-friendly getter body still performs the
 * wildcard-removal cast under the configured getterStyle. Gates the
 * SuppressWarnings("unchecked") annotation on the getter.
 */
function styledGetterUsesCast(
  field: FieldDef,
  style: GetterStyle = context.Global.Config.GetterStyle,
): boolean {
  const optional = field.Optional && !field.IsAdditionalProperties;
  const nullable = field.Nullable && !field.IsAdditionalProperties;

  if (isRawGetter(style) && (optional || nullable)) {
    return false;
  }
  if (isAlwaysOptionalGetter(style) && optional === nullable) {
    return false;
  }
  return true;
}
registerTemplateFunc("styledGetterUsesCast", styledGetterUsesCast);

function getConstructorMetadata(param: JavaParam): AccessorMetadata {
  const typeInfo = getFieldTypeInfo(param);
  const variant = getFieldVariant(param);
  return buildConstructorMetadata(param, typeInfo, variant);
}
registerTemplateFunc("getConstructorMetadata", getConstructorMetadata);

type BuilderMetadata = {
  fieldType: string;
  setters: AccessorMetadata[];
};

function getConstructorParamType(fieldParam: JavaParam): string {
  const { requiredType, boxedType, nullableType } =
    getFieldTypeInfo(fieldParam);
  const fieldVariant = getFieldVariant(fieldParam);
  switch (fieldVariant) {
    case "nullable-optional-default":
    case "nullable-optional":
      return `${templateNullableAnnotation(nullableType!)}`;
    case "nullable-only":
      return `${templateNullableAnnotation(
        fieldParam.Default ? nullableType! : boxedType,
      )}`;
    case "optional-only":
      return `${templateNullableAnnotation(boxedType)}`;
    case "required":
      if (isPrimitive(requiredType)) {
        return requiredType;
      }
      if (fieldParam.Default) {
        return `${templateNullableAnnotation(requiredType)}`;
      }
      return `${templateNonNullAnnotation(requiredType)}`;
  }
}

function getBuilderMetadata(fieldParam: JavaParam): BuilderMetadata {
  const { requiredType, boxedType, nullableType } =
    getFieldTypeInfo(fieldParam);
  const setters = getSettersMetadata(fieldParam);
  const fieldVariant = getFieldVariant(fieldParam);

  switch (fieldVariant) {
    case "nullable-optional-default":
    case "nullable-optional":
      return {
        fieldType: nullableType!,
        setters,
      };
    case "nullable-only":
      if (fieldParam.Default) {
        return {
          fieldType: nullableType!,
          setters,
        };
      }
      return {
        fieldType: boxedType,
        setters: [
          {
            paramType: `${templateNullableAnnotation(boxedType)}`,
            body: [`this.${fieldParam.Name} = ${fieldParam.Name};`],
          },
        ],
      };
    case "optional-only":
      return {
        fieldType: boxedType,
        setters,
      };
    case "required":
      return {
        fieldType: requiredType,
        setters,
      };
  }
}

registerTemplateFunc("getBuilderMetadata", getBuilderMetadata);

// JVM method-signature local-variable slot ceiling is 255 (JVMS §4.11). `this`
// consumes 1 slot; long/double consume 2 each; reference types consume 1.
// Generated all-args ctors use reference-typed params, so a field count above
// ~254 trips `javac: error: too many parameters`. 250 leaves margin for future
// primitive params and the synthetic `this`.
function exceedsJvmParamLimit(fields: FieldDef[]): boolean {
  const JAVAV2_CTOR_PARAM_LIMIT = 250;
  const ctorParams = nonConstFields(removeAdditionalPropertiesField(fields));

  return ctorParams.length > JAVAV2_CTOR_PARAM_LIMIT;
}

registerTemplateFunc("exceedsJvmParamLimit", exceedsJvmParamLimit);
