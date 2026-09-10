function generateSDKInvoker(
  symbolManager: Record<string, boolean>,
  resourceOperationType: TerraformResourceOperationType,
  reqIdentifier: string,
  resIdentifier: string,
  operation: TerraformOperation,
  entity: TerraformEntity,
): string {
  let res = "";
  function generateCheck(
    reqIdentifier: string,
    resIdentifier: string,
    resourceOperationType: TerraformResourceOperationType,
    op: TerraformOperation,
  ) {
    const sdkAccessor = `r.client${
      op.APIOperation.OwningSDK.Group !== ""
        ? "." + sanitizeFieldName(op.APIOperation.OwningSDK.Type.Name)
        : ""
    }`;
    const sdkMethodName = sanitizeMethodName(op.APIOperation);

    let subRes = "";
    let args = `ctx`;
    if (reqIdentifier) {
      if (
        operation.APIOperation.Request.Field?.Type?.Type.toString() ===
          "array" ||
        operation.APIOperation.Request.Field?.Type?.Type.toString() === "set"
      ) {
        addImportScope(operation.APIOperation.Request.Field.Type);
        args += `, ${templateType(operation.APIOperation.Request.Field.Type)}{${
          operation.APIOperation.Request.Field.Type.ContainsNull ? "" : "*"
        }${reqIdentifier}}`;
      } else {
        args += `, ${
          isRequestOptional(operation.APIOperation.Request) ? "" : "*"
        }${reqIdentifier}`;
      }
    }

    if (operation.IncludeOperationSecurity) {
      const operationSecurityVariableName = getPluralizedVarSymbolName(
        symbolManager,
        sanitizeMethodName(op.APIOperation),
        "Security",
      );

      subRes += templateOperationSecurity(
        operation.APIOperation.Security,
        operationSecurityVariableName,
        symbolManager,
      );
      subRes += `\n`;
      args += `, ${operationSecurityVariableName}`;
    }

    let optionsLength = 0;

    if (operation.ServerAttributeName) {
      optionsLength++;
    }

    if (op.Options?.Polling) {
      optionsLength++;
    }

    const optsVariableName = getPluralizedVarSymbolName(
      symbolManager,
      sanitizeMethodName(op.APIOperation),
      "Options",
    );

    if (optionsLength > 0) {
      addGenImport("models/operations", true);

      subRes += `\n`;
      subRes += `${optsVariableName} := make([]operations.Option, 0, 1)\n`;
      args += `, ${optsVariableName}...`;
    }

    if (operation.ServerAttributeName) {
      const serverURLVariableName = getPluralizedVarSymbolName(
        symbolManager,
        operation.ServerAttributeName,
      );
      const serverURLDataModelFieldName = sanitizeFieldName(
        operation.ServerAttributeName,
      );

      subRes += `\n`;
      subRes += `if !data.${serverURLDataModelFieldName}.IsNull() && !data.${serverURLDataModelFieldName}.IsUnknown() {\n`;
      subRes += `  ${serverURLVariableName} := data.${serverURLDataModelFieldName}.ValueString()\n`;
      subRes += `\n`;
      subRes += `  if ${serverURLVariableName} != "" {\n`;
      subRes += `    ${optsVariableName} = append(${optsVariableName}, operations.WithServerURL(${serverURLVariableName}))\n`;
      subRes += `  }\n`;
      subRes += `}\n`;
      subRes += `\n`;
    }

    if (op.Options?.Polling) {
      const sdkPollingMethodName = `${sdkMethodName}${op.Options.Polling.Name}`;
      const sdkPollingOption = `${sdkAccessor}.${sdkPollingMethodName}()`;

      addGenImport("models/operations", true);

      // Closing parenthesis added after polling override options
      subRes += `${optsVariableName} = append(${optsVariableName}, operations.WithPolling(\n${sdkPollingOption},\n`;

      if (op.Options.Polling.DelaySeconds !== null) {
        addGenImport("polling", true);
        subRes += `polling.WithDelaySecondsOverride(${op.Options.Polling.DelaySeconds}),\n`;
      }

      if (op.Options.Polling.IntervalSeconds !== null) {
        addGenImport("polling", true);
        subRes += `polling.WithIntervalSecondsOverride(${op.Options.Polling.IntervalSeconds}),\n`;
      }

      if (op.Options.Polling.LimitCount !== null) {
        addGenImport("polling", true);
        subRes += `polling.WithLimitCountOverride(${op.Options.Polling.LimitCount}),\n`;
      }

      subRes += `))\n`;
    }

    subRes += `${templateIndent(
      1,
    )}${resIdentifier}, err := ${sdkAccessor}.${sdkMethodName}(${args})\n`;
    subRes += `${templateIndent(1)}if err != nil {\n`;
    subRes += `${templateIndent(
      2,
    )}resp.Diagnostics.AddError("failure to invoke API", err.Error())\n`;
    subRes += `${templateIndent(
      2,
    )}if ${resIdentifier} != nil && ${resIdentifier}.RawResponse != nil {\n`;
    subRes += `${templateIndent(
      3,
    )}resp.Diagnostics.AddError("unexpected http request/response", debugResponse(${resIdentifier}.RawResponse))\n`;
    subRes += `${templateIndent(2)}}\n`;
    subRes += `${templateIndent(2)}return\n`;
    subRes += `${templateIndent(1)}}\n`;
    subRes += `${templateIndent(1)}if ${resIdentifier} == nil {\n`;
    subRes += `${templateIndent(
      2,
    )}resp.Diagnostics.AddError("unexpected response from API", fmt.Sprintf("%v", ${resIdentifier}))\n`;
    subRes += `${templateIndent(2)}return\n`;
    subRes += `${templateIndent(1)}}\n`;
    const missingCodes = operation.EntityMissingCodes ?? [];
    if (missingCodes.length && resourceOperationType === "read") {
      const conditions = missingCodes
        .map((code) => `${resIdentifier}.StatusCode == ${code}`)
        .join(" || ");
      subRes += `${templateIndent(1)}if ${conditions} {\n`;
      subRes += `${templateIndent(2)}resp.State.RemoveResource(ctx)\n`;
      subRes += `${templateIndent(2)}return\n`;
      subRes += `${templateIndent(1)}}\n`;
    }
    const successCodes = [...operation.SuccessCodes];

    // Add special handling for 409 Conflict if it's in the spec and this is a create operation
    if (operation.Has409ConflictCode && resourceOperationType === "create") {
      subRes += `${templateIndent(1)}if ${resIdentifier}.StatusCode == 409 {\n`;
      subRes += `${templateIndent(2)}resp.Diagnostics.AddError(\n`;
      subRes += `${templateIndent(3)}"Resource Already Exists",\n`;
      subRes += `${templateIndent(
        3,
      )}"When creating this resource, the API indicated that this resource already exists. You can bring the existing resource under management using Terraform import functionality or retry with a unique configuration.",\n`;
      subRes += `${templateIndent(2)})\n`;
      subRes += `${templateIndent(2)}return\n`;
      subRes += `${templateIndent(1)}}\n`;
    }

    // Unlike read operations, do not return early for entity missing codes if
    // there are multiple delete operations. Instead the logic will move to the
    // next delete operation logic (if any) and eventually return no errors,
    // signaling resource destroy success to Terraform.
    if (missingCodes.length && resourceOperationType === "delete") {
      successCodes.push(...missingCodes.map((code) => code.toString()));
    }

    const unexpectedCodeErrorDiagnostic = [
      `resp.Diagnostics.AddError(fmt.Sprintf("unexpected response from API. Got an unexpected response code %v", ${resIdentifier}.StatusCode), debugResponse(${resIdentifier}.RawResponse))`,
      `return`,
    ]
      .map((line) => `${templateIndent(2)}${line}\n`)
      .join("");
    switch (successCodes.length) {
      case 0:
        break;
      case 1:
        const successCode = successCodes[0];
        const checkCode = successCode.includes("X")
          ? `fmt.Sprintf("%v", ${resIdentifier}.StatusCode)[0] != '2'`
          : `${resIdentifier}.StatusCode != ${successCode}`;
        subRes += `${templateIndent(1)}if ${checkCode} {\n`;
        subRes += unexpectedCodeErrorDiagnostic;
        subRes += `${templateIndent(1)}}\n`;
        break;
      default:
        const hasXXCode = successCodes.some((code) => code.includes("X"));
        // Handle both numeric 2XX codes and 2XX wildcard code with single check,
        // since ultimately it just needs a "2" prefix.
        if (hasXXCode) {
          subRes += `${templateIndent(
            1,
          )}if fmt.Sprintf("%v", ${resIdentifier}.StatusCode)[0] != '2' {\n`;
          subRes += unexpectedCodeErrorDiagnostic;
          subRes += `${templateIndent(1)}}\n`;
        } else {
          subRes += `${templateIndent(
            1,
          )}switch ${resIdentifier}.StatusCode {\n`;
          subRes += `${templateIndent(1)}case ${successCodes.join(", ")}:\n`;
          subRes += `${templateIndent(2)}break\n`;
          subRes += `${templateIndent(1)}default:\n`;
          subRes += unexpectedCodeErrorDiagnostic;
          subRes += `${templateIndent(1)}}\n`;
        }
    }
    return subRes;
  }

  res += generateCheck(
    reqIdentifier,
    resIdentifier,
    resourceOperationType,
    operation,
  );

  // ResponseBodyFieldDef is nil when there is no response body relevant to
  // the entity or when the operation skips data model refresh (e.g. delete,
  // final invoke). This single nil check covers both cases.
  if (operation.ResponseBodyFieldDef) {
    const responseBodyFieldDef = operation.ResponseBodyFieldDef;

    let accessor = `${resIdentifier}.${sanitizeFieldName(
      responseBodyFieldDef.Name,
    )}`;

    if (responseBodyFieldDef.Optional || responseBodyFieldDef.Nullable) {
      // TODO: Switch to idiomatic Go conditional (not done yet to separate code
      // churn changes from functional changes)
      // res += `if ${accessor} == nil {\n`;
      res += `if !(${accessor} != nil) {\n`;
      res += `resp.Diagnostics.AddError("unexpected response from API. Got an unexpected response body", debugResponse(${resIdentifier}.RawResponse))\n`;
      res += `return\n`;
      res += `}\n`;
    }

    // arrays are always passed as direct arrays, not pointers-to-arrays
    if (
      !responseBodyFieldDef.Optional &&
      !isGoArray(responseBodyFieldDef.Type)
    ) {
      // If it's a value type (non optional), we should pass it in as a pointer
      accessor = `&` + accessor;
    }

    // For paginated operations, reinitialize list/array fields before the first refresh
    // to prevent appending to stale state during terraform refresh
    // Only nil out fields that exist in the response (not request-only fields)
    if (
      operation.SupportsPagination &&
      entity.SchemaTypeDef?.Fields &&
      responseBodyFieldDef.Type.Fields
    ) {
      for (const field of entity.SchemaTypeDef.Fields) {
        if (isGoArray(field.Type)) {
          // Check if this field exists in the response
          const existsInResponse = responseBodyFieldDef.Type.Fields.some(
            (responseField) => responseField.Name === field.Name,
          );
          if (existsInResponse) {
            const fieldName = sanitizeFieldName(field.Name);
            res += `data.${fieldName} = nil\n`;
          }
        }
      }
    }

    // ResponseBodySDKMethod is nil when the response body type has no
    // compatible data model conversion, in which case we skip the refresh.
    // The method name already accounts for IsArrayWrapper.
    if (operation.ResponseBodySDKMethod) {
      const refreshFromFunc = `data.${operation.ResponseBodySDKMethod.MethodName}`;
      res += `resp.Diagnostics.Append(${refreshFromFunc}(ctx, ${accessor})...)\n\n`;
      res += `if resp.Diagnostics.HasError() {\n`;
      res += `return\n`;
      res += `}\n`;
    }

    // Apply pagination loop when the operation supports it and has a
    // compatible response body conversion method.
    if (operation.SupportsPagination && operation.ResponseBodySDKMethod) {
      res += `for {\n`;
      res += `  var err error\n`;
      res += `\n`;
      res += `  ${resIdentifier}, err = ${resIdentifier}.Next()\n`;
      res += `\n`;
      res += `  if err != nil {\n`;
      res += `    resp.Diagnostics.AddError("failed to retrieve next page of results", err.Error())\n`;
      res += `    if ${resIdentifier} != nil && ${resIdentifier}.RawResponse != nil {\n`;
      res += `      resp.Diagnostics.AddError("unexpected http request/response", debugResponse(${resIdentifier}.RawResponse))\n`;
      res += `    }\n`;
      res += `    return\n`;
      res += `  }\n`;
      res += `\n`;
      res += `  if ${resIdentifier} == nil {\n`;
      res += `    break\n`;
      res += `  }\n`;
      res += `\n`;

      const refreshFromFunc = `data.${operation.ResponseBodySDKMethod.MethodName}`;
      res += `  resp.Diagnostics.Append(${refreshFromFunc}(ctx, ${accessor})...)\n`;
      res += `\n`;
      res += `  if resp.Diagnostics.HasError() {\n`;
      res += `    return\n`;
      res += `  }\n`;
      res += `}\n`;
    }
  }

  return res;
}

function generateSDKCall(
  entity: TerraformEntity,
  operations: TerraformOperation[],
  resourceOperationType: TerraformResourceOperationType,
) {
  if (!operations?.length) {
    switch (resourceOperationType) {
      case "create":
      case "invoke":
      case "open":
        throw new Error(
          `Configuration Error, missing ${entity.Name}#${resourceOperationType} operation`,
        );
      case "read":
        return "// Not Implemented; we rely entirely on CREATE API request response";
      case "update":
        return "// Not Implemented; all attributes marked as RequiresReplace";
      case "delete":
        return "// Not Implemented; entity does not have a configured DELETE operation";
      default:
        throw new Error(`Invalid ${resourceOperationType} operation`);
    }
  }
  const symbolManager = makeSymbolMananger();
  symbolManager["ctx"] = true;
  if (entity.IncludeSDKMethodOptions && resourceOperationType == "update") {
    symbolManager["state"] = true;
    symbolManager["stateModel"] = true;
    symbolManager["options"] = true;
  }
  let resCode = "";

  function invokeOp(op: TerraformOperation) {
    const requestVar = getPluralizedVarSymbolName(symbolManager, "request");
    const resVar = getPluralizedVarSymbolName(symbolManager, "res");

    // RequestBodySDKMethod is nil when the operation has no request body.
    // Every operation with a request body is guaranteed to have a method,
    // so a nil check here is sufficient.
    if (!op.RequestBodySDKMethod) {
      const invokerCode = generateSDKInvoker(
        symbolManager,
        resourceOperationType,
        "",
        resVar,
        op,
        entity,
      );
      resCode += invokerCode;
      return;
    }

    const dataToSDKMethodName = op.RequestBodySDKMethod.MethodName;
    const diagsVariable = getPluralizedVarSymbolName(
      symbolManager,
      requestVar,
      "Diags",
    );

    let extraArgument = "";
    if (entity.IncludeSDKMethodOptions) {
      // Pass opts for create/update operations:
      //  - WriteOnly attributes
      //  - (Update only) Patch semantics
      //  - (Update only) Prior state parameters
      if (
        resourceOperationType == "create" ||
        resourceOperationType == "update"
      ) {
        extraArgument = `, opts`;
      } else {
        extraArgument = `, nil`;
      }
    }

    resCode +=
      `${requestVar}, ${diagsVariable} := data.${dataToSDKMethodName}(ctx${extraArgument})\n` +
      `resp.Diagnostics.Append(${diagsVariable}...)\n\n` +
      `if resp.Diagnostics.HasError() {\n` +
      `return\n` +
      `}\n`;
    resCode += generateSDKInvoker(
      symbolManager,
      resourceOperationType,
      requestVar,
      resVar,
      op,
      entity,
    );

    if (
      resourceOperationType === "create" ||
      resourceOperationType === "update"
    ) {
      resCode += `\nresp.Diagnostics.Append(refreshPlan(ctx, plan, &data)...)\n\n`;
      resCode += `if resp.Diagnostics.HasError() {\n`;
      resCode += `return\n`;
      resCode += `}\n`;
    }
  }

  for (const op of operations) {
    invokeOp(op);
  }

  if (
    (resourceOperationType === "create" &&
      "CreateNeedsReadAfter" in entity.Operations &&
      entity.Operations.CreateNeedsReadAfter) ||
    (resourceOperationType === "update" &&
      "UpdateNeedsReadAfter" in entity.Operations &&
      entity.Operations.UpdateNeedsReadAfter)
  ) {
    for (const op of (entity.Operations as TerraformManagedResourceOperations)
      .Read) {
      invokeOp(op);
    }
  }

  return resCode;
}

registerTemplateFunc("generateSDKCall", generateSDKCall);
