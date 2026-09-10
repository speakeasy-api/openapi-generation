type GlobalParametersReadmeInfo = {
  NeedsGlobalParameters: boolean;
  HasRequired: boolean;
  ParameterLen: number;
  ExampleParameter: FieldDef | null;
  Usage: UsageContext | null;
};

function prepareGlobalParametersSection(): GlobalParametersReadmeInfo {
  let info: GlobalParametersReadmeInfo = {
    NeedsGlobalParameters: false,
    HasRequired: false,
    ParameterLen: 0,
    ExampleParameter: null,
    Usage: null,
  };
  const scope: UsageExampleScope = {
    OpFilter: "global-parameters",
    Feature: "",
    IsGlobal: true,
  };
  const usages = selectExampleOperations(
    context.Global.AST.MainSDK,
    1,
    [scope],
    false,
  );

  if (context.Global.AST.MainSDK.Globals && usages.length > 0) {
    info.Usage = usages[0];
    info.NeedsGlobalParameters = true;
    info.ParameterLen = context.Global.AST.MainSDK.Globals.Fields.length;
    for (const param of context.Global.AST.MainSDK.Globals.Fields) {
      if (!param.Optional) {
        info.HasRequired = true;
      }
      if (!info.ExampleParameter) {
        info.ExampleParameter = param;
      }

      // Add any missing global parameter scopes
      if (
        !info.Usage.Scopes?.find(
          (s) =>
            s.Value &&
            originalFieldName(s.Value.field) === originalFieldName(param),
        )
      ) {
        // Find an example for the global parameter
        let example = undefined;

        if (info.Usage.Operation.Request?.Params) {
          const params: ParamDef[] = [
            ...info.Usage.Operation.Request.Params.QueryParams,
            ...info.Usage.Operation.Request.Params.PathParams,
            ...info.Usage.Operation.Request.Params.HeaderParams,
          ];
          for (const opParam of params) {
            if (originalFieldName(opParam.Field) !== originalFieldName(param)) {
              continue;
            }

            if (opParam.Examples?.length > 0) {
              example = opParam.Examples[0];
              break;
            }
          }
        }

        info.Usage.Scopes?.push({
          OpFilter: "",
          Feature: "parameter",
          IsGlobal: true,
          Value: {
            field: param,
            example: example,
          },
        });
      }
    }
  }

  return info;
}

function operationUsesEventStreams(operation: Operation): boolean {
  for (const response of operation.Response.Responses) {
    for (const content of response.Content) {
      if (content.SerializationMethod == "eventstream") {
        return true;
      }
    }
  }
  return false;
}
registerTemplateFunc("operationUsesEventStreams", operationUsesEventStreams);

function operationUsesJsonLStreams(operation: Operation): boolean {
  for (const response of operation.Response.Responses) {
    for (const content of response.Content) {
      if (content.SerializationMethod == "jsonl") {
        return true;
      }
    }
  }
  return false;
}
registerTemplateFunc("operationUsesJsonLStreams", operationUsesJsonLStreams);

function isFirstContentTypeEventStream(operation: Operation): boolean {
  for (const response of operation.Response.Responses) {
    // Skip error responses
    if (response.Error) {
      continue;
    }
    if (response.Content.length > 0) {
      return response.Content[0].SerializationMethod == "eventstream";
    }
  }
  return false;
}
registerTemplateFunc(
  "isFirstContentTypeEventStream",
  isFirstContentTypeEventStream,
);

function isFirstContentTypeJsonLStream(operation: Operation): boolean {
  for (const response of operation.Response.Responses) {
    // Skip error responses
    if (response.Error) {
      continue;
    }
    if (response.Content.length > 0) {
      return response.Content[0].SerializationMethod == "jsonl";
    }
  }
  return false;
}
registerTemplateFunc(
  "isFirstContentTypeJsonLStream",
  isFirstContentTypeJsonLStream,
);

function operationHasResponseStream(op: Operation): boolean {
  for (const subResp of op.Response.Responses) {
    for (const content of subResp.Content) {
      if (
        content.SerializationMethod === "raw" &&
        content.Content.Type.Type == "response-stream"
      ) {
        return true;
      }
    }
  }
}
registerTemplateFunc("operationHasResponseStream", operationHasResponseStream);

function operationHasFileUpload(op: Operation): boolean {
  return fileUploadOpPredicate(context.Global.AST.MainSDK, op);
}

registerTemplateFunc("operationHasFileUpload", operationHasFileUpload);

// @ts-ignore
function templateGlobalsTable(): string {
  // NOTE: The `Required` column is removed until required-field detection is reliable
  // as globals are currently marked as optional by default.

  // const headers: string[] = ["Name", "Type", "Required", "Description"];
  const headers: string[] = ["Name", "Type", "Description"];
  if (context.Global.Config.EnvVarPrefix) {
    headers.push("Environment");
  }

  const contents: string[][] = [headers];

  context.Global.AST.MainSDK.Globals.Fields.forEach((field: FieldDef) => {
    const row: string[] = [
      templateReadmeFieldName(field),
      templateGlobalType(field),
      // !field.Optional ? ":heavy_check_mark:" : "",
      templateGlobalDescription(field),
    ];

    if (context.Global.Config.EnvVarPrefix) {
      row.push(templateGlobalEnvVars(field));
    }

    contents.push(row);
  });

  return createMarkdownTable(contents, false);
}

registerTemplateFunc("templateGlobalsTable", templateGlobalsTable);

// @ts-ignore
function templateReadmeGlobalValue(fieldDef: FieldDef): string {
  return templateValue(fieldDef);
}
registerTemplateFunc("templateReadmeGlobalValue", templateReadmeGlobalValue);

// @ts-ignore
function templateReadmeFieldName(fieldDef: FieldDef): string {
  return sanitizeFieldName(fieldDef.Name);
}
registerTemplateFunc("templateReadmeFieldName", templateReadmeFieldName);

// @ts-ignore
function templateGlobalDescription(field: FieldDef): string {
  if (field.Comments?.Summary ?? "" !== "") {
    return sanitizeMarkdownTableContent(field.Comments.Summary);
  }

  if (field.Comments?.Description ?? "" !== "") {
    return sanitizeMarkdownTableContent(field.Comments.Description);
  }

  return `The ${sanitizeFieldName(field.Name)} parameter.`;
}
