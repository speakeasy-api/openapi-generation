class ValidationError extends Error {
  constructor(message: string) {
    super(message);
    this.name = "ValidationError";
  }
}

type ErrorType = {
  Type: TypeDef;
  StatusCodes: string[];
  ContentType: string;
  Description: string;
};

// @ts-ignore
function sanitizeErrorType(typeDef: TypeDef): string {
  return sanitizeClass(typeDef, "", false);
}

// @ts-ignore
function collectErrorTypes(
  operation: Operation,
  dissectUnionsOfErrors: boolean,
): ErrorType[] {
  const errors: ErrorType[] = [];
  for (const response of operation.Response?.Responses) {
    if (response.Error) {
      if (response.Content.length > 0) {
        for (const c of response.Content) {
          const error = {
            StatusCodes: response.Code.toString().split(","),
            ContentType: c.ContentType.toString(),
            Description: c.Content.Comments?.Description || "",
          } as Partial<ErrorType>;
          if (dissectUnionsOfErrors && isUnionOfErrors(c.Content.Type)) {
            if (c.Content.Type.Discriminator) {
              for (const m of c.Content.Type.Discriminator.Mapping) {
                errors.push({
                  ...error,
                  Type: m.Type,
                  Description:
                    m.Type.Comments?.Description || error.Description,
                } as ErrorType);
              }
            } else {
              for (const t of c.Content.Type.AssociatedTypes) {
                errors.push({
                  ...error,
                  Type: t,
                  Description: t.Comments?.Description || error.Description,
                } as ErrorType);
              }
            }
          } else {
            errors.push({
              ...error,
              Type: c.Content.Type,
              Description:
                c.Content.Type.Comments?.Description || error.Description,
            } as ErrorType);
          }
        }
      }
    }
  }

  return errors;
}

type ErrorTypeGroup = {
  Type: TypeDef;
  StatusCodes: string[];
  ContentTypes: string[];
  Descriptions: string[];
  Operations: Operation[];
  IsCommonError: boolean;
};

type ErrorGroups = {
  All: ErrorTypeGroup[];
  PrimaryErrors: ErrorTypeGroup[];
  PrimaryErrorCount: number;
  LessCommonErrors: ErrorTypeGroup[];
  LessCommonErrorCount: number;
  HasLessCommonErrors: boolean;
  TotalOperationCount: number;
  HasErrorsApplicableToNotAllOperations: boolean;
};

function groupGlobalErrorsByErrorType(): ErrorGroups {
  const errors: Record<string, ErrorTypeGroup> = {};
  const shouldDissectUnionsOfErrors = !["pythonv2", "go"].includes(
    context.Global.Config.Language,
  );
  for (const op of iterateOperations()) {
    const errorTypes = collectErrorTypes(op, shouldDissectUnionsOfErrors);
    for (const errorType of errorTypes) {
      errors[errorType.Type.Name] ||= {
        Type: errorType.Type,
        StatusCodes: [],
        ContentTypes: [],
        Operations: [],
        Descriptions: [],
        IsCommonError: false,
      };
      errors[errorType.Type.Name].Operations = [
        ...new Set([...errors[errorType.Type.Name].Operations, op]),
      ];
      errors[errorType.Type.Name].Descriptions = [
        ...new Set([
          ...errors[errorType.Type.Name].Descriptions,
          errorType.Description || "",
        ]),
      ].sort();
      errors[errorType.Type.Name].StatusCodes = [
        ...new Set([
          ...errors[errorType.Type.Name].StatusCodes,
          ...errorType.StatusCodes,
        ]),
      ].sort();
      errors[errorType.Type.Name].ContentTypes = [
        ...new Set([
          ...errors[errorType.Type.Name].ContentTypes,
          errorType.ContentType,
        ]),
      ].sort();
    }
  }

  const all = Object.values(errors).sort((a, b) => {
    if (a.Operations.length === b.Operations.length) {
      return a.StatusCodes.sort()
        .join(",")
        .localeCompare(b.StatusCodes.sort().join(","));
    }
    return b.Operations.length - a.Operations.length;
  });

  const totalOperationCount = countNestedOperations();
  const primaryErrors = [];
  const lessCommonErrors = [];

  // Applicable to 80% of operations
  const commonErrorThreshold = totalOperationCount * 0.8;
  all.forEach((group) => {
    if (group.Operations.length >= commonErrorThreshold) {
      group.IsCommonError = true;
      primaryErrors.push(group);
    } else {
      group.IsCommonError = false;
      lessCommonErrors.push(group);
    }
  });

  const hasErrorsApplicableToNotAllOperations = all.some(
    (group) => group.Operations.length < totalOperationCount,
  );

  return {
    All: all,
    PrimaryErrors: primaryErrors,
    PrimaryErrorCount: primaryErrors.length,
    LessCommonErrors: lessCommonErrors,
    LessCommonErrorCount: lessCommonErrors.length,
    HasLessCommonErrors: lessCommonErrors.length > 0,
    TotalOperationCount: totalOperationCount,
    HasErrorsApplicableToNotAllOperations:
      hasErrorsApplicableToNotAllOperations,
  };
}
registerTemplateFunc(
  "groupGlobalErrorsByErrorType",
  groupGlobalErrorsByErrorType,
);

function templateErrorGroup(
  group: ErrorTypeGroup,
  totalOperationCount: number = 0,
): string {
  const parts = [
    templateErrorGroupDescription(group),
    templateErrorGroupStatusCodes(group),
    templateErrorGroupContentTypes(group),
    templateErrorGroupOperations(
      group,
      totalOperationCount,
      !group.IsCommonError,
    ),
  ];

  return `${templateErrorGroupMarkdownLink(group)}: ${
    parts.filter(Boolean).join(" ") || "Generic error."
  }`;
}
registerTemplateFunc("templateErrorGroup", templateErrorGroup);

// @ts-ignore
function templateErrorGroupMarkdownLink(_group: ErrorTypeGroup): string {
  // This function must return a markdown formatted link to the file that contains the error class
  // definition for the given error group, e.g. [`ErrorClassName`](./path/to/error_file.py)
  throw new Error(
    `templateErrorGroupMarkdownLink not implemented for ${context.Global.Config.Language}`,
  );
}

function templateErrorGroupDescription(group: ErrorTypeGroup): string {
  if (group.Descriptions.length === 1) {
    return trailingDot(singleLine(group.Descriptions[0]));
  }
  return "";
}

function templateErrorGroupStatusCodes(group: ErrorTypeGroup): string {
  if (group.StatusCodes.length === 1 && group.StatusCodes[0] !== "default") {
    return trailingDot(`Status code \`${group.StatusCodes[0]}\``);
  }
  return "";
}

function templateErrorGroupContentTypes(group: ErrorTypeGroup): string {
  const filteredContentTypes = group.ContentTypes.filter(
    (ct) => !ct.includes("json"),
  );
  if (filteredContentTypes.length === 1) {
    return trailingDot(`Content type: \`${filteredContentTypes[0]}\``);
  }
  return "";
}

function templateErrorGroupOperations(
  group: ErrorTypeGroup,
  totalOperationCount: number = 0,
  showOperationCount: boolean = true,
): string {
  totalOperationCount ||= countNestedOperations();
  if (group.Operations.length === totalOperationCount) {
    return "";
  }

  if (showOperationCount) {
    return `Applicable to ${group.Operations.length} of ${totalOperationCount} methods.*`;
  }

  const notApplicableToAllOperations =
    totalOperationCount > group.Operations.length;
  if (notApplicableToAllOperations) {
    return `*`;
  }

  return "";
}

// @ts-ignore
function templateErrorsTable(
  operation: Operation,
  header: string = "",
  uniformCellWidth: boolean = false,
): string {
  const contents = [["Error Type", "Status Code", "Content Type"]];

  const errors = collectErrorTypes(
    operation,
    !["pythonv2", "go"].includes(context.Global.Config.Language),
  );

  errors.forEach((e) => {
    contents.push([
      sanitizeErrorType(e.Type),
      e.StatusCodes.join(", "),
      e.ContentType,
    ]);
  });

  const unhandledRanges = ["4XX", "5XX"].filter(
    (range) => !errors.flatMap((error) => error.StatusCodes).includes(range),
  );

  if (unhandledRanges.length > 0) {
    contents.push([
      templateDefaultError(),
      unhandledRanges.join(", "),
      "\\*/\\*",
    ]);
  }

  const table = createMarkdownTable(contents, uniformCellWidth);
  return header ? `${header}\n\n${table}` : table;
}

registerTemplateFunc("templateErrorsTable", templateErrorsTable);

// @ts-ignore
function getDefaultErrorClassName(): string {
  return sanitizeClassName(context.Global.Config.DefaultErrorName);
}
registerTemplateFunc("getDefaultErrorClassName", getDefaultErrorClassName);

// @ts-ignore
function getDefaultErrorFileName(): string {
  return sanitizeFileName(context.Global.Config.DefaultErrorName);
}
registerTemplateFunc("getDefaultErrorFileName", getDefaultErrorFileName);

// @ts-ignore
function getBaseErrorClassName(): string {
  return sanitizeClassName(context.Global.Config.BaseErrorName);
}
registerTemplateFunc("getBaseErrorClassName", getBaseErrorClassName);

// @ts-ignore
function getBaseErrorFileName(): string {
  return sanitizeFileName(context.Global.Config.BaseErrorName);
}
registerTemplateFunc("getBaseErrorFileName", getBaseErrorFileName);

function pickBestErrorToHandle(op: Operation): TypeDef | undefined {
  for (const res of op.Response.Responses) {
    if (!res.Error) continue;

    for (const c of res.Content) {
      if (isUnionOfErrors(c.Content.Type)) {
        if (c.Content.Type.Discriminator) {
          for (const m of c.Content.Type.Discriminator.Mapping) {
            return m.Type;
          }
        }
        for (const t of c.Content.Type.AssociatedTypes) {
          return t;
        }
      } else {
        return c.Content.Type;
      }
    }
  }
}
registerTemplateFunc("pickBestErrorToHandle", pickBestErrorToHandle);

// @ts-ignore
function getNonErrorStatusCodes(response: ResponseDef): string[] {
  const errorCodes = response.GetErrorStatusCodes();
  const nonErrorCodes = response.Responses.flatMap((resp) => resp.Code).filter(
    (code) => code !== "default" && !errorCodes.includes(code),
  );
  return sanitizeStatusCodes(nonErrorCodes);
}
