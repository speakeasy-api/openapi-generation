function parseURL(operation: Operation) {
  let protocol = undefined;
  let baseUrl = "{{baseUrl}}";

  if (operation.Servers && operation.Servers.Servers.length > 0) {
    // Parse the operation.Servers.Servers[0].URL to get the protocol and base URL using regex
    const url = operation.Servers.Servers[0].URL;

    const protocolMatch = url.match(/(https?):\/\/.*\/?/);

    // Parse out the base URL of the server, without the protocol, and without any trailing slashes
    const baseUrlMatch = url.match(/https?:\/\/(.*)\/?/);

    if (protocolMatch && protocolMatch.length > 1) {
      protocol = protocolMatch[1];
    }
    if (baseUrlMatch && baseUrlMatch.length > 1) {
      baseUrl = baseUrlMatch[1];
    }
  }

  return { protocol, baseUrl };
}

function parseJSONRequest(field: FieldDef) {
  if (field.Type.Type === "class") {
    const output = new Map<string, any>();

    for (const f of field.Type.Fields) {
      output.set(f.Name, parseJSONRequest(f));
    }
    return { [field.Name]: Object.fromEntries(output) };
  }

  return undefined;
}

function createHost(baseUrl: string) {
  if (
    baseUrl === "{{baseUrl}}" ||
    baseUrl == "" ||
    baseUrl == undefined ||
    baseUrl == null
  ) {
    return ["{{baseUrl}}"];
  } else {
    return baseUrl.split(".");
  }
}

function createItems(context: Context) {
  const items = [];

  const globalParams = (context.Global.AST.MainSDK.Globals?.Fields || []).map(
    (p) => p.Annotations.Get("param"),
  );

  // Iterate over all the tags in the SDK
  for (const SDK of context.Global.AST.MainSDK.SubSDKs) {
    const tagItems = [];

    for (const operation of SDK.Operations) {
      // All Parameters for the request
      const Parameters = [];

      // Path, Query, and Header parameters for the request
      const PathParams = [];
      const QueryParams = [];
      const HeaderParams = [];

      const addUniqueParam = (param: FieldDef) => {
        if (!param.Annotations.Has("param")) {
          return;
        }

        if (
          !Parameters.some(
            (existing: FieldDef) =>
              existing.Name === param.Name &&
              existing.Annotations.Get("param").ParamType ===
                param.Annotations.Get("param").ParamType,
          )
        ) {
          Parameters.push(param);
        }
      };

      const sortParamByType = (parameter: FieldDef) => {
        if (!parameter.Annotations.Has("param")) {
          return;
        }

        const annotation = parameter.Annotations.Get("param");
        let global = false;

        if (globalParams.length > 0) {
          if (globalParams.some((param) => param.IsEqual(annotation))) {
            global = true;
          }
        }

        switch (annotation.ParamType) {
          case "pathParam":
            PathParams.push({
              key: parameter.Name,
              value: global
                ? `{{${parameter.Name}}}`
                : parameter.Default || parseValue(parameter),
              type: parameter.Type.Type,
            });
            break;

          case "queryParam":
            QueryParams.push({
              key: parameter.Name,
              value: global
                ? `{{${parameter.Name}}}`
                : parameter.Default || parseValue(parameter),
              type: parameter.Type.Type,
            });
            break;

          case "header":
            HeaderParams.push({
              key: parameter.Name,
              value: global
                ? `{{${parameter.Name}}}`
                : parameter.Default || parseValue(parameter),
              type: parameter.Type.Type,
            });
            break;

          default:
            throw new Error(`Unknown parameter type: ${annotation.ParamType}`);
        }
      };

      // Response content types
      const responseContentTypes = new Set<string>();

      // Request body
      let body = undefined;

      for (const param of operation.Request?.Params?.PathParams ?? []) {
        addUniqueParam(param.Field);
      }
      for (const param of operation.Request?.Params?.QueryParams ?? []) {
        addUniqueParam(param.Field);
      }
      for (const param of operation.Request?.Params?.HeaderParams ?? []) {
        addUniqueParam(param.Field);
      }

      // All parameters should be included under each parameter type, for now.
      if (Parameters.length > 0) {
        for (const parameter of Parameters) {
          sortParamByType(parameter);
        }
      }

      // Populate the request body and the request bodies content type
      if (operation.Request?.RequestBody) {
        if (operation.Request?.RequestBody.Annotations.Has("request")) {
          const requestType =
            operation.Request.RequestBody.Annotations.Get("request");

          HeaderParams.push({
            key: "Content-Type",
            value: requestType.MediaType,
            type: "string",
          });

          switch (requestType.MediaType) {
            case "application/json":
              body = {
                mode: "raw",
                raw: seralizeFieldDefAsJSON(operation.Request?.RequestBody),
              };
              break;

            case "text/plain":
              body = { mode: "raw", raw: "" };
              break;

            case "application/x-www-form-urlencoded":
              body = { mode: "x-www-form-urlencoded", urlencoded: [] };
              break;

            case "multipart/form-data":
              body = { mode: "formdata", formdata: [] };
              break;

            default:
              body = undefined;
              break;
          }
        }
      }

      if (operation.Response.Responses.length > 0) {
        for (const response of operation.Response?.Responses) {
          for (const content of response.Content) {
            responseContentTypes.add(content.ContentType);
          }
        }

        HeaderParams.push({
          key: "Accept",
          value: Array.from(responseContentTypes).join(","),
          type: "string",
        });
      }

      const { protocol, baseUrl } = parseURL(operation);

      tagItems.push({
        name:
          operation.Comments?.Summary ||
          operation.Method + " " + operation.Path,
        request: {
          auth: createAuth(operation.Security),
          method: operation.Method.toUpperCase(),
          header: HeaderParams.length > 0 ? HeaderParams : undefined,
          body,
          url: {
            raw: `${protocol ? `${protocol}://` : ""}${baseUrl}${
              operation.Path
            }`,
            protocol,
            host: createHost(baseUrl),
            path: createPath(operation.Path),
            query: QueryParams.length > 0 ? QueryParams : undefined,
            variable: PathParams.length > 0 ? PathParams : undefined,
          },
          description: operation.Comments?.Description,
        },
      });
    }

    items.push({
      name: SDK.FieldName,
      item: tagItems,
      description: SDK.Comments?.Description,
    });
  }

  return items;
}
