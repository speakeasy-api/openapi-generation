type ServerReadmeContext = {
  ServerVariableSetter: string;
  ServerByName: string;
  ServerByIndex: string;
  ServerByUrl: string;
};

type UsageGlobalServer = {
  ID?: string;
  Index?: number;
  Variables?: ServerVariable[];
};

function getUsageServerUrl(
  usageContext: UsageContext,
  scope?: UsageExampleScope,
  additionalContext?: TemplateValueContext,
  defaultURL: string = "https://api.example.com",
): string {
  if (scope == undefined) {
    // Fallback to the global server_url scope if no scope is provided
    scope = usageContext.Scopes?.find(
      (scope) => scope.Feature === "server_url" && scope.IsGlobal,
    );
  }

  let serverURL = scope?.Value;

  if (serverURL) {
    const envVar = getEnvVarFromDirective(serverURL);
    if (envVar != undefined) {
      return languageSpecificEnvVarWrapping(envVar, additionalContext);
    }
  }

  if (!serverURL) {
    if (
      scope?.IsGlobal == false &&
      usageContext.Operation?.Servers != undefined
    ) {
      serverURL = usageContext.Operation.Servers.GetDefaultURL(true);
    } else if (context.Global.AST.MainSDK.Servers?.HasAbsoluteURL()) {
      serverURL = context.Global.AST.MainSDK.Servers.GetDefaultURL(true);
    } else {
      serverURL = defaultURL;
    }
  }

  return templateStringValue(
    serverURL,
    {
      Optional: false,
      Nullable: false,
    } as FieldDef,
    additionalContext,
  );
}

function getUsageGlobalServerIndex(selectMostComplex: boolean): number {
  const servers = context.Global.AST.MainSDK.Servers?.Servers;
  if (!servers) {
    return -1;
  }

  let serverIndex = 0;

  const serverIDToShowInSnippets =
    context.Global.Config?.ServerToShowInSnippets;
  if (serverIDToShowInSnippets != undefined) {
    const isArrayIndex = /^\d+$/.test(serverIDToShowInSnippets);
    if (isArrayIndex) {
      serverIndex = Number(serverIDToShowInSnippets);
      if (serverIndex >= servers.length) {
        logger().Error(
          `Error: usageSnippets.serverToShowInSnippets:${serverIndex} is out of bounds. There are only ${servers.length} servers defined.`,
        );
        serverIndex = 0;
      }
    } else {
      serverIndex = servers.findIndex(
        (s: Server) => s.ID === serverIDToShowInSnippets,
      );
      if (serverIndex === -1) {
        logger().Error(
          `Error: usageSnippets.serverToShowInSnippets:"${serverIDToShowInSnippets}" does not match any server IDs.`,
        );
        serverIndex = 0;
      }
    }
  }

  if (!selectMostComplex) {
    return serverIndex;
  }

  // if applicable, select the server with the most variables
  serverIndex = servers.reduceRight(
    (selected: number, s: Server, i: number) => {
      if (s.Variables?.length > servers[selected].Variables?.length) {
        return i;
      }

      return selected;
    },
    serverIndex,
  );

  return serverIndex;
}

function getUsageGlobalServer(selectMostComplex: boolean): UsageGlobalServer {
  const mainSDK = context.Global.AST.MainSDK;
  const servers = mainSDK.Servers?.Servers;

  if (!servers) {
    return {
      Variables: mainSDK.Servers?.GetVariables(),
    };
  }

  const idx = getUsageGlobalServerIndex(selectMostComplex);
  const server = servers[idx];

  if (mainSDK.Servers?.ServerMap) {
    return {
      ID: server.ID,
      Variables: server.Variables,
    };
  }

  return {
    Index: idx,
    Variables: server.Variables,
  };
}

// @ts-ignore
function getUsageServerVariableValue(v: ServerVariable): string {
  if (v.Type.Enum) {
    const options = v.Type.Enum.Values;
    return options[options.length - 1];
  }

  // Use the actual default value from the server definition if available
  if (v.Default && v.Default !== "") {
    return v.Default;
  }

  // Only fall back to fake values if no default is provided
  if (isNumeric(v.Default)) {
    return fakeInteger(v.Name.toLowerCase()).toString();
  }

  return fakeString(v.Name.toLowerCase());
}

function templateServersTable(sdk: SDK): string {
  const servers = sdk.Servers;
  if (!servers) {
    return "";
  }

  const hasVariables = servers.GetVariables().length > 0;
  const hasDescription = servers.Servers.some(
    (s: Server) => s.Comments?.Description != "",
  );

  const headers: string[] = [
    servers.ServerMap ? "Name" : "#",
    "Server",
    ...(hasVariables ? ["Variables"] : []),
    ...(hasDescription ? ["Description"] : []),
  ];

  const contents: string[][] = [headers];
  servers.Servers.forEach((server: Server, i: number) => {
    const serverID: string = servers.ServerMap
      ? `\`${server.ID}\``
      : i.toString();

    const row: string[] = [serverID, `\`${server.URL}\``];

    if (hasVariables) {
      row.push(
        server.Variables?.map((v) => `\`${v.Name}\``).join("<br/>") ?? "",
      );
    }

    if (hasDescription) {
      row.push(server.Comments?.Description ?? "");
    }

    contents.push(row);
  });

  return createMarkdownTable(contents, false);
}

registerTemplateFunc("templateServersTable", templateServersTable);

function templateServerVariablesTable(sdk: SDK): string {
  if (!sdk.Servers) {
    return "";
  }

  const variables = sdk.Servers.GetVariables();
  const hasEnum = variables.some(
    (v: ServerVariable) => v.Type.Type.toString() == "enum",
  );
  const hasDescription = variables.some(
    (v: ServerVariable) => v.Type.Comments?.Description != "",
  );
  const setterName = caser().ToPascal(
    getServerReadmeContext().ServerVariableSetter,
  );
  const headers: string[] = [
    "Variable",
    setterName,
    ...(hasEnum ? ["Supported Values"] : []),
    "Default",
    ...(hasDescription ? ["Description"] : []),
  ];

  const contents: string[][] = [headers];
  variables.forEach((v: ServerVariable) => {
    contents.push([
      `\`${v.Name}\``,
      `\`${templateServerVariableSetter(v)}\``,
      ...(hasEnum ? [templateServerVariableSupportedValues(v)] : []),
      `\`"${v.Default}"\``,
      ...(hasDescription ? [v.Type.Comments?.Description ?? ""] : []),
    ]);
  });

  return createMarkdownTable(contents, false);
}

registerTemplateFunc(
  "templateServerVariablesTable",
  templateServerVariablesTable,
);

function shouldShowServerSelection(): boolean {
  const configVal = context.Global.Config?.ServerToShowInSnippets ?? "";
  return configVal !== "";
}

function addServerSelectionScopeIfInConfig(local: UsageContext): string {
  if (!shouldShowServerSelection()) {
    return "";
  }

  if (!local.Scopes) {
    local.Scopes = [];
  }

  const existing = local.Scopes.find(
    (s) => s.Feature === "server_selection" && s.IsGlobal,
  );
  if (existing) {
    return "";
  }

  local.Scopes.push({
    OpFilter: "",
    Feature: "server_selection",
    IsGlobal: true,
    Value: false,
  } as UsageExampleScope);
  return "";
}

registerTemplateFunc(
  "addServerSelectionScopeIfInConfig",
  addServerSelectionScopeIfInConfig,
);

function templateServerVariableSupportedValues(
  variable: ServerVariable,
): string {
  if (!variable.Type.Enum) {
    return templateGlobalType({ Type: variable.Type } as FieldDef);
  }

  return variable.Type.Enum.Values.map((val) => `- \`"${val}"\``).join("<br/>");
}
