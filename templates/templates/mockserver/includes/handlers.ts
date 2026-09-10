/** Go net/http package HTTP method constants.
 * Reference: https://pkg.go.dev/net/http#pkg-constants
 */
type GoNetHTTPMethod =
  | "http.MethodDelete"
  | "http.MethodGet"
  | "http.MethodHead"
  | "http.MethodOptions"
  | "http.MethodPatch"
  | "http.MethodPost"
  | "http.MethodPut"
  | "http.MethodTrace";

/** Go net/http package HTTP status constants.
 * Reference: https://pkg.go.dev/net/http#pkg-constants
 */
type GoNetHTTPStatus =
  | "http.StatusContinue" // 100 // RFC 9110, 15.2.1
  | "http.StatusSwitchingProtocols" // 101 // RFC 9110, 15.2.2
  | "http.StatusProcessing" // 102 // RFC 2518, 10.1
  | "http.StatusEarlyHints" // 103 // RFC 8297
  | "http.StatusOK" // 200 // RFC 9110, 15.3.1
  | "http.StatusCreated" // 201 // RFC 9110, 15.3.2
  | "http.StatusAccepted" // 202 // RFC 9110, 15.3.3
  | "http.StatusNonAuthoritativeInfo" // 203 // RFC 9110, 15.3.4
  | "http.StatusNoContent" // 204 // RFC 9110, 15.3.5
  | "http.StatusResetContent" // 205 // RFC 9110, 15.3.6
  | "http.StatusPartialContent" // 206 // RFC 9110, 15.3.7
  | "http.StatusMultiStatus" // 207 // RFC 4918, 11.1
  | "http.StatusAlreadyReported" // 208 // RFC 5842, 7.1
  | "http.StatusIMUsed" // 226 // RFC 3229, 10.4.1
  | "http.StatusMultipleChoices" // 300 // RFC 9110, 15.4.1
  | "http.StatusMovedPermanently" // 301 // RFC 9110, 15.4.2
  | "http.StatusFound" // 302 // RFC 9110, 15.4.3
  | "http.StatusSeeOther" // 303 // RFC 9110, 15.4.4
  | "http.StatusNotModified" // 304 // RFC 9110, 15.4.5
  | "http.StatusUseProxy" // 305 // RFC 9110, 15.4.6
  | "http.StatusTemporaryRedirect" // 307 // RFC 9110, 15.4.8
  | "http.StatusPermanentRedirect" // 308 // RFC 9110, 15.4.9
  | "http.StatusBadRequest" // 400 // RFC 9110, 15.5.1
  | "http.StatusUnauthorized" // 401 // RFC 9110, 15.5.2
  | "http.StatusPaymentRequired" // 402 // RFC 9110, 15.5.3
  | "http.StatusForbidden" // 403 // RFC 9110, 15.5.4
  | "http.StatusNotFound" // 404 // RFC 9110, 15.5.5
  | "http.StatusMethodNotAllowed" // 405 // RFC 9110, 15.5.6
  | "http.StatusNotAcceptable" // 406 // RFC 9110, 15.5.7
  | "http.StatusProxyAuthRequired" // 407 // RFC 9110, 15.5.8
  | "http.StatusRequestTimeout" // 408 // RFC 9110, 15.5.9
  | "http.StatusConflict" // 409 // RFC 9110, 15.5.10
  | "http.StatusGone" // 410 // RFC 9110, 15.5.11
  | "http.StatusLengthRequired" // 411 // RFC 9110, 15.5.12
  | "http.StatusPreconditionFailed" // 412 // RFC 9110, 15.5.13
  | "http.StatusRequestEntityTooLarge" // 413 // RFC 9110, 15.5.14
  | "http.StatusRequestURITooLong" // 414 // RFC 9110, 15.5.15
  | "http.StatusUnsupportedMediaType" // 415 // RFC 9110, 15.5.16
  | "http.StatusRequestedRangeNotSatisfiable" // 416 // RFC 9110, 15.5.17
  | "http.StatusExpectationFailed" // 417 // RFC 9110, 15.5.18
  | "http.StatusTeapot" // 418 // RFC 9110, 15.5.19 (Unused)
  | "http.StatusMisdirectedRequest" // 421 // RFC 9110, 15.5.20
  | "http.StatusUnprocessableEntity" // 422 // RFC 9110, 15.5.21
  | "http.StatusLocked" // 423 // RFC 4918, 11.3
  | "http.StatusFailedDependency" // 424 // RFC 4918, 11.4
  | "http.StatusTooEarly" // 425 // RFC 8470, 5.2.
  | "http.StatusUpgradeRequired" // 426 // RFC 9110, 15.5.22
  | "http.StatusPreconditionRequired" // 428 // RFC 6585, 3
  | "http.StatusTooManyRequests" // 429 // RFC 6585, 4
  | "http.StatusRequestHeaderFieldsTooLarge" // 431 // RFC 6585, 5
  | "http.StatusUnavailableForLegalReasons" // 451 // RFC 7725, 3
  | "http.StatusInternalServerError" // 500 // RFC 9110, 15.6.1
  | "http.StatusNotImplemented" // 501 // RFC 9110, 15.6.2
  | "http.StatusBadGateway" // 502 // RFC 9110, 15.6.3
  | "http.StatusServiceUnavailable" // 503 // RFC 9110, 15.6.4
  | "http.StatusGatewayTimeout" // 504 // RFC 9110, 15.6.5
  | "http.StatusHTTPVersionNotSupported" // 505 // RFC 9110, 15.6.6
  | "http.StatusVariantAlsoNegotiates" // 506 // RFC 2295, 8.1
  | "http.StatusInsufficientStorage" // 507 // RFC 4918, 11.5
  | "http.StatusLoopDetected" // 508 // RFC 5842, 7.2
  | "http.StatusNotExtended" // 510 // RFC 2774, 7
  | "http.StatusNetworkAuthenticationRequired"; // 511 // RFC 6585, 6

/** Describes a HTTP handler for the mock server. */
type MockServerHandler = {
  /** File name.  */
  FileName: string;

  /** Go function name. */
  FunctionName: string;

  /** Method. */
  Method: string;

  /** Path. */
  Path: string;

  /** Request and response contexts within the handler, such as individual tests. */
  HandlerContexts: HandlerContext[];
};

/** Describes a request and response context within a HTTP handler, such as an individual test. */
type HandlerContext = {
  /** Function name. */
  FunctionName: string;

  /** Test name. */
  TestName: string;

  /** Request handling context. */
  Request: HandlerRequestContext;

  /** Response handling context. */
  Response: HandlerResponseContext;

  /** Operation */
  Operation: Operation;

  /** Operation count */
  OperationCount: number;

  /** AST usage context. */
  UsageContext: UsageContext;
};

/** Describes a request context within a HTTP handler, such as parameters. */
type HandlerRequestContext = {
  /** Mapping of generic request header names to contexts. */
  Headers?: Record<string, RequestHeaderContext>;

  /** Accept header values. */
  AcceptValues?: string;

  /** Request content type. */
  ContentType?: string;

  /** Whether the request body is required. */
  IsRequestBodyRequired?: boolean;

  /** Mapping of request header parameter names to templated values. */
  HeaderParameters?: Record<string, string>;

  /** Mapping of request security header names to contexts. */
  HeaderSecurity?: Record<string, RequestHeaderSecurityContext>;

  /** Mapping of request path parameter names to templated values. */
  PathParameters?: Record<string, string>;

  /** Mapping of request query parameter names to templated values. */
  QueryParameters?: Record<string, string>;
};

/** Describes a response context within a HTTP handler, such as an individual test. */
type HandlerResponseContext = {
  /** Usage context of the associated operation. */
  UsageContext: UsageContext;

  /** Response body assertion. */
  ResponseBodyAssertions?: ResponseBodyAssertion[];

  /** AST response body content. */
  BodyContent?: ResponseBodyContent;

  /** Response body value. */
  Body?: Example;

  /** Response body serialization method, such as json or multipart. */
  BodySerialization?: ResponseBodyContentSerializationMethod;

  /** Response body value field definition. */
  BodyField?: FieldDef;

  /** Response HTTP headers. */
  Headers?: Record<string, string>;

  /** Response HTTP status code. */
  Status: GoNetHTTPStatus;

  /** AST individual response. */
  SubResponse?: SubResponse;
};

/** Describes a generic request header context within a HTTP handler. */
type RequestHeaderContext = {
  /** Enabled if the header should only be checked for existence. */
  Exists?: boolean;

  /** Templated value(s) to check, if value(s) are fully known. */
  Values?: string;
};

/** Returns all HTTP handlers for the mock server based on operations. */
function collectMockServerHandlers(ast: AST): MockServerHandler[] {
  const testGroups = ast.Tests?.TestGroups ?? [];

  if (testGroups.length == 0) {
    return [];
  }

  let handlers: Record<string, MockServerHandler> = {};

  for (const testGroup of testGroups) {
    for (const test of testGroup.Tests) {
      if (!test.Workflow) {
        continue;
      }
      handlers = collectWorkflowHandlers(test.Workflow, handlers);
    }
  }

  // Post-process: add gorilla mux wildcard regex to path params whose test
  // values contain slashes (e.g., folder="product/test" → {folder:.+}).
  applyPathParamWildcards(handlers);

  const handlersToReturn: MockServerHandler[] = [];

  Object.keys(handlers)
    .sort()
    .forEach((key) => {
      handlersToReturn.push(handlers[key]);
    });

  return handlersToReturn;
}

// Apply gorilla mux wildcard regex to path parameters whose spec examples
// contain slashes (e.g., folder path "product/test"). Called after all
// handlers are collected so we can inspect all handler contexts.
function applyPathParamWildcards(
  handlers: Record<string, MockServerHandler>,
): void {
  for (const handler of Object.values(handlers)) {
    // Collect path param original names that have slash-containing example values
    const slashParams = new Set<string>();
    for (const ctx of handler.HandlerContexts) {
      const pathParams = ctx.Operation?.Request?.Params?.PathParams;
      if (!pathParams) continue;
      for (const param of pathParams) {
        for (const example of param.Examples || []) {
          if (example.Reference) continue;
          const json = example.ToJSON?.();
          if (typeof json === "string" && json.includes("/")) {
            slashParams.add(originalFieldName(param.Field));
          }
        }
      }
    }
    if (slashParams.size === 0) continue;
    // Replace {param} → {param:.+} only for params with slash-containing examples
    handler.Path = handler.Path.replace(/\{([^}:]+)\}/g, (match, name) => {
      if (slashParams.has(name)) {
        return `{${name}:.+}`;
      }
      return match;
    });
  }
}

function collectWorkflowHandlers(
  workflow: ArazzoWorkflow,
  handlers: Record<string, MockServerHandler>,
): Record<string, MockServerHandler> {
  let operationCount = 0;

  for (const step of workflow.Steps ?? []) {
    switch (step.Type) {
      case "operation":
        const invocation = step.Invocation;
        const operation = invocation?.Operation ?? step.Operation;
        const usageContext = step.UsageContext;
        const stepIdx = invocation?.StepIdx ?? step.StepIdx;

        if (!operation || !usageContext) {
          continue;
        }

        const method = operation.Method;
        const path = operation.Path.split("#")[0];

        const pathName = `path_${method}_${path}`;

        const fileName = `${sanitizeFileName(pathName)}.go`;
        const handlerFunctionName = sanitizePathHandlerName(pathName);

        const functionName = sanitizeTestHandlerMethodName(
          operation,
          workflow.Name,
          stepIdx,
        );
        const handlerContext = collectHandlerContext(
          functionName,
          workflow,
          operation,
          usageContext,
          operationCount++,
        );

        if (pathName in handlers) {
          const handler = handlers[pathName];

          handler.HandlerContexts.push(handlerContext);

          handlers[pathName] = handler;
        } else {
          handlers[pathName] = {
            FileName: fileName,
            Method: method,
            Path: path,
            FunctionName: handlerFunctionName,
            HandlerContexts: [handlerContext],
          };
        }
        break;
      case "workflow":
        const nestedWorkflow = resolveWorkflowStepTarget(step);
        if (nestedWorkflow) {
          handlers = collectWorkflowHandlers(nestedWorkflow, handlers);
        }
        break;
    }
  }

  return handlers;
}

/** Returns the HTTP handler context for the usage context. */
function collectHandlerContext(
  functionName: string,
  workflow: ArazzoWorkflow,
  operation: Operation,
  usageContext: UsageContext,
  operationCount: number,
): HandlerContext {
  const request = collectHandlerRequestContext(usageContext);
  const response = collectHandlerResponseContext(operation, usageContext);

  return {
    FunctionName: functionName,
    TestName: workflow.Name,
    Request: request,
    Response: response,
    Operation: operation,
    OperationCount: operationCount,
    UsageContext: usageContext,
  };
}

/** Returns the HTTP handler request context for a given usage context. */
function collectHandlerRequestContext(
  usageContext: UsageContext,
): HandlerRequestContext {
  const request = usageContext.Operation.Request;

  let headers = collectStaticRequestHeaders(usageContext);
  let contentType = undefined;
  let isRequestBodyRequired = false;
  let headerParameters: Record<string, string> = {};
  let pathParameters: Record<string, string> = {};
  let queryParameters: Record<string, string> = {};
  let acceptHeaderValues = "";

  if (request?.RequestBody) {
    const requestAnnotation = request.RequestBody.Annotations?.Get(
      "request",
    ) as RequestAnnotation;

    contentType = requestAnnotation.MediaType;

    isRequestBodyRequired = request.IsRequestBodyRequired;
  }

  const acceptTypes = usageContext.Operation.GetAcceptTypes();

  if (acceptTypes.length > 0) {
    acceptHeaderValues = templateStringSlice(acceptTypes);
  }

  if (request?.Params) {
    for (const param of request.Params.HeaderParams) {
      const paramExampleValue = templateHeaderParameterExampleValues(
        usageContext,
        param,
      );

      if (!paramExampleValue) {
        continue;
      }

      headerParameters[param.Field.Name] = paramExampleValue;
    }

    for (const param of request.Params.PathParams) {
      const paramExampleValue = templatePathParameterExampleValue(
        usageContext,
        param,
      );

      if (!paramExampleValue) {
        continue;
      }

      pathParameters[param.Field.Name] = paramExampleValue;
    }

    for (const param of request.Params.QueryParams) {
      const paramExampleValues = templateQueryParameterExampleValues(
        usageContext,
        param,
      );

      if (!paramExampleValues) {
        continue;
      }

      queryParameters = { ...queryParameters, ...paramExampleValues };
    }
  }

  return {
    Headers: headers,
    AcceptValues: acceptHeaderValues,
    ContentType: contentType,
    IsRequestBodyRequired: isRequestBodyRequired,
    HeaderParameters: headerParameters,
    HeaderSecurity: collectRequestHeaderSecurity(usageContext),
    PathParameters: pathParameters,
    QueryParameters: queryParameters,
  };
}

/** Returns the HTTP handler response context for a given usage context. */
function collectHandlerResponseContext(
  operation: Operation,
  usageContext: UsageContext,
): HandlerResponseContext {
  const statusCodeAssertion = usageContext.Assertions?.find(
    (a) => a.TargetType.toString() === "statusCode",
  );

  let subResponse: SubResponse | undefined;

  if (statusCodeAssertion) {
    subResponse = getUsageContextSubResponse(
      usageContext,
      statusCodeAssertion.Value as string,
    );
  }

  if (!subResponse) {
    const statusCode = statusCodeAssertion?.Value as string;

    const bodyContent = getResponseBodyContentFromStatusCode(
      usageContext,
      statusCode,
    );

    const headers: Record<string, string> = {};

    if (bodyContent?.ContentType) {
      headers["Content-Type"] = bodyContent.ContentType;
    }

    return {
      UsageContext: usageContext,
      Status: statusCode
        ? getSubResponseCodeGoNetHTTPStatus(operation, statusCode)
        : "http.StatusNotImplemented",
      Headers: headers,
    };
  }

  const resAssertions = (usageContext.Assertions ?? []).filter(
    (a) => a.TargetType.toString() === "responseBody",
  );

  let bodyContent: ResponseBodyContent | undefined = undefined;
  let body: Example = undefined;
  let bodySerialization: ResponseBodyContentSerializationMethod;
  let bodyField: FieldDef;
  let headers: Record<string, string> = {};

  let responseBodyAssertions: ResponseBodyAssertion[] = [];

  if (resAssertions.length > 0) {
    for (const resAssertion of resAssertions) {
      const responseBodyAssertion = resAssertion.Value as ResponseBodyAssertion;
      responseBodyAssertions.push(responseBodyAssertion);

      if (!bodyContent) {
        bodyContent = responseBodyAssertion.Content;
      }
    }
  }

  if (!bodyContent) {
    bodyContent = getResponseBodyContent(
      usageContext,
      subResponse,
      resAssertions.length > 0 ? resAssertions[0] : undefined,
    );
  }

  if (bodyContent) {
    if (bodyContent.ContentType) {
      headers["Content-Type"] = bodyContent.ContentType;
    }

    body = getResponseContentValue(usageContext, bodyContent);
    bodySerialization = bodyContent.SerializationMethod;
    bodyField = bodyContent.Content;
  }

  return {
    UsageContext: usageContext,
    ResponseBodyAssertions: responseBodyAssertions,
    BodyContent: bodyContent,
    Body: body,
    BodySerialization: bodySerialization,
    BodyField: bodyField,
    Headers: headers,
    Status: getSubResponseCodesGoNetHTTPStatus(
      operation,
      subResponse.Code,
      statusCodeAssertion,
    ),
    SubResponse: subResponse,
  };
}

/** Returns the collection of static request header contexts, such as Accept and
 * User-Agent headers, for the HTTP handler. */
function collectStaticRequestHeaders(
  usageContext: UsageContext,
): Record<string, RequestHeaderContext> {
  let result: Record<string, RequestHeaderContext> = {};

  if (usageContext.Operation.UsesUserAgentHeader) {
    result["X-Speakeasy-User-Agent"] = {
      Exists: true,
    };
  } else {
    result["User-Agent"] = {
      Exists: true,
    };
  }

  return result;
}

function getSubResponseCodesGoNetHTTPStatus(
  op: Operation,
  codes: string[],
  statusCodeAssertion: Assertion | undefined,
): GoNetHTTPStatus {
  if (!codes || codes.length == 0) {
    // TODO: not sure this is the right thing to do here
    return "http.StatusNotImplemented";
  }

  if (statusCodeAssertion) {
    const statusCode = statusCodeAssertion?.Value as string;

    if (codes.includes(statusCode)) {
      return getSubResponseCodeGoNetHTTPStatus(op, statusCode);
    }
  }

  // TODO: not sure this is the right thing to do here
  return getSubResponseCodeGoNetHTTPStatus(op, codes[0]);
}

/** Returns the Go net/http status constant for a SubResponse code. */
function getSubResponseCodeGoNetHTTPStatus(
  op: Operation,
  code: string,
): GoNetHTTPStatus {
  switch (code.toUpperCase()) {
    case "100":
      return "http.StatusContinue";
    case "101":
      return "http.StatusSwitchingProtocols";
    case "102":
      return "http.StatusProcessing";
    case "103":
      return "http.StatusEarlyHints";
    case "200":
      return "http.StatusOK";
    case "201":
      return "http.StatusCreated";
    case "202":
      return "http.StatusAccepted";
    case "203":
      return "http.StatusNonAuthoritativeInfo";
    case "204":
      return "http.StatusNoContent";
    case "205":
      return "http.StatusResetContent";
    case "206":
      return "http.StatusPartialContent";
    case "207":
      return "http.StatusMultiStatus";
    case "208":
      return "http.StatusAlreadyReported";
    case "226":
      return "http.StatusIMUsed";
    case "300":
      return "http.StatusMultipleChoices";
    case "301":
      return "http.StatusMovedPermanently";
    case "302":
      return "http.StatusFound";
    case "303":
      return "http.StatusSeeOther";
    case "304":
      return "http.StatusNotModified";
    case "305":
      return "http.StatusUseProxy";
    case "307":
      return "http.StatusTemporaryRedirect";
    case "308":
      return "http.StatusPermanentRedirect";
    case "400":
      return "http.StatusBadRequest";
    case "401":
      return "http.StatusUnauthorized";
    case "402":
      return "http.StatusPaymentRequired";
    case "403":
      return "http.StatusForbidden";
    case "404":
      return "http.StatusNotFound";
    case "405":
      return "http.StatusMethodNotAllowed";
    case "406":
      return "http.StatusNotAcceptable";
    case "407":
      return "http.StatusProxyAuthRequired";
    case "408":
      return "http.StatusRequestTimeout";
    case "409":
      return "http.StatusConflict";
    case "410":
      return "http.StatusGone";
    case "411":
      return "http.StatusLengthRequired";
    case "412":
      return "http.StatusPreconditionFailed";
    case "413":
      return "http.StatusRequestEntityTooLarge";
    case "414":
      return "http.StatusRequestURITooLong";
    case "415":
      return "http.StatusUnsupportedMediaType";
    case "416":
      return "http.StatusRequestedRangeNotSatisfiable";
    case "417":
      return "http.StatusExpectationFailed";
    case "418":
      return "http.StatusTeapot";
    case "421":
      return "http.StatusMisdirectedRequest";
    case "422":
      return "http.StatusUnprocessableEntity";
    case "423":
      return "http.StatusLocked";
    case "424":
      return "http.StatusFailedDependency";
    case "425":
      return "http.StatusTooEarly";
    case "426":
      return "http.StatusUpgradeRequired";
    case "428":
      return "http.StatusPreconditionRequired";
    case "429":
      return "http.StatusTooManyRequests";
    case "431":
      return "http.StatusRequestHeaderFieldsTooLarge";
    case "451":
      return "http.StatusUnavailableForLegalReasons";
    case "500":
      return "http.StatusInternalServerError";
    case "501":
      return "http.StatusNotImplemented";
    case "502":
      return "http.StatusBadGateway";
    case "503":
      return "http.StatusServiceUnavailable";
    case "504":
      return "http.StatusGatewayTimeout";
    case "505":
      return "http.StatusHTTPVersionNotSupported";
    case "506":
      return "http.StatusVariantAlsoNegotiates";
    case "507":
      return "http.StatusInsufficientStorage";
    case "508":
      return "http.StatusLoopDetected";
    case "510":
      return "http.StatusNotExtended";
    case "511":
      return "http.StatusNetworkAuthenticationRequired";
    case "2XX":
      return "http.StatusOK";
    case "3XX":
      return "http.StatusMultipleChoices";
    case "4XX":
      return "http.StatusBadRequest";
    case "5XX":
      return "http.StatusInternalServerError";
    case "DEFAULT":
      const usedCodes = new Set<string>();
      for (const subRes of op.Response?.Responses ?? []) {
        for (const code of subRes.Code) {
          usedCodes.add(code);
        }
      }

      for (const statusCode of getValidStatusCodes()) {
        if (!usedCodes.has(statusCode)) {
          return getSubResponseCodeGoNetHTTPStatus(op, statusCode);
        }
      }

      throw new Error(`unhandled code: DEFAULT, no valid status codes found`);
    default:
      throw new Error(`unhandled code: ${code}`);
  }
}

/** Returns the Go net/http method constant for an OAS method name. */
function templateGoNetHTTPMethod(method: string): GoNetHTTPMethod {
  switch (method) {
    case "delete":
      return "http.MethodDelete";
    case "get":
      return "http.MethodGet";
    case "head":
      return "http.MethodHead";
    case "options":
      return "http.MethodOptions";
    case "patch":
      return "http.MethodPatch";
    case "post":
      return "http.MethodPost";
    case "put":
      return "http.MethodPut";
    case "trace":
      return "http.MethodTrace";
    default:
      throw new Error(`unhandled method: ${method}`);
  }
}

registerTemplateFunc("templateGoNetHTTPMethod", templateGoNetHTTPMethod);

/** Templates the response body value using the SDK. */
function templateResponseBodyValue(
  fieldDef: FieldDef,
  example?: any,
  context?: TemplateValueContext,
): string {
  const value = templateValue(fieldDef, example, false, {
    ...(context ?? {}),
    templateDefaultValue: true,
    isTest: true,
  });

  return `${getResponseBodyDeclaration(fieldDef)}${value}`;
}
registerTemplateFunc("templateResponseBodyValue", templateResponseBodyValue);

function getResponseBodyDeclaration(field: FieldDef): string {
  let declaration = `respBody := `;
  const needsType =
    field.Optional ||
    field.Nullable ||
    !["class", "map", "array", "set"].includes(field.Type.Type.toString());

  if (needsType) {
    const fieldType = sanitizeFieldType(field, "usage");

    declaration = `var respBody ${fieldType} = `;
  }

  return declaration;
}
