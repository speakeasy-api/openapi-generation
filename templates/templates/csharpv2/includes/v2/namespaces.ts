// @ts-ignore
function getModelsNamespaceRoot(): string {
  return `${getPackageId()}.${MODELS_FOLDER_NAME}`;
}

// @ts-ignore
function getModelsTopLevelFolder(): string {
  return `${getSDKTopLevelFolder()}/${MODELS_FOLDER_NAME}`;
}

registerTemplateFunc("getModelsTopLevelFolder", getModelsTopLevelFolder);

// Under netstandard2, Tests bind the SDK's netstandard2.0 asset. That asset's
// IAsyncEnumerable<T> comes from Microsoft.Bcl.AsyncInterfaces, which the Tests
// restore does not pull in transitively: it flows the SDK's netX dependencies.
// @ts-ignore
function templateTestProjectReferences(): string {
  const projectPath = `..\\${getProjectFilePath()}`;
  if (!isNetstandard2()) {
    return `<ProjectReference Include="${projectPath}" />`;
  }
  const asyncInterfaces = getNetstandard2Dependency(
    "Microsoft.Bcl.AsyncInterfaces",
  );
  return [
    `<ProjectReference Include="${projectPath}">
      <SetTargetFramework>TargetFramework=netstandard2.0</SetTargetFramework>
    </ProjectReference>`,
    `<PackageReference Include="${asyncInterfaces.name}" Version="${asyncInterfaces.version}" />`,
  ].join("\n    ");
}

unregisterTemplateFunc("templateTestProjectReferences");
registerTemplateFunc(
  "templateTestProjectReferences",
  templateTestProjectReferences,
);

function stripModelsPrefix(location: string): string {
  if (location === MODELS_FOLDER_NAME) {
    return "";
  }

  const prefix = `${MODELS_FOLDER_NAME}/`;
  if (location.startsWith(prefix)) {
    return location.slice(prefix.length);
  }

  return location;
}

// @ts-ignore
function getModelsLocation(outputLocation: string): string {
  return stripModelsPrefix(sanitizeOutputLocation(outputLocation));
}

unregisterTemplateFunc("getModelsLocation");
registerTemplateFunc("getModelsLocation", getModelsLocation);

// @ts-ignore
function getErrorsLocation() {
  return stripModelsPrefix(getScopePath("errors"));
}

unregisterTemplateFunc("getErrorsLocation");
registerTemplateFunc("getErrorsLocation", getErrorsLocation);

// @ts-ignore
function getGeneratorsTopLevelFolder(): string {
  return `${getSDKTopLevelFolder()}.${GENERATORS_PROJECT_NAME}`;
}

registerTemplateFunc(
  "getGeneratorsTopLevelFolder",
  getGeneratorsTopLevelFolder,
);

// @ts-ignore
function getGeneratorsPackageId(): string {
  return `${getPackageId()}.${GENERATORS_PROJECT_NAME}`;
}

registerTemplateFunc("getGeneratorsPackageId", getGeneratorsPackageId);

function getGeneratorsFilePath(): string {
  const winPath = getGeneratorsTopLevelFolder().replaceAll("/", "\\");
  return `${winPath}\\${getGeneratorsPackageId()}.csproj`;
}

registerTemplateFunc("getGeneratorsFilePath", getGeneratorsFilePath);

// @ts-ignore
function getGeneratorsProjectReference(): string {
  const folder = getGeneratorsTopLevelFolder().split("/").slice(-1)[0];
  return `..\\${folder}\\${getGeneratorsPackageId()}.csproj`;
}

registerTemplateFunc(
  "getGeneratorsProjectReference",
  getGeneratorsProjectReference,
);

// @ts-ignore
function getGeneratorsTargetFramework(): string {
  return GENERATORS_TARGET_FRAMEWORK;
}

registerTemplateFunc(
  "getGeneratorsTargetFramework",
  getGeneratorsTargetFramework,
);

// @ts-ignore
function getModelNamespace(
  outputLocation: string,
  full: boolean = true,
): string {
  return getAccessNamespace(
    getModelsLocation(outputLocation),
    full,
    false,
    getModelsNamespaceRoot(),
  );
}

unregisterTemplateFunc("getModelNamespace");
registerTemplateFunc("getModelNamespace", getModelNamespace);

// @ts-ignore
function getScopeNamespace(scope: string, full: boolean = false): string {
  switch (scope) {
    case "utils":
      return getAccessNamespace(getScopePath(scope), full, true);
    case "hooks":
      return getAccessNamespace(getScopePath(scope), full);
    default:
      return getModelNamespace(getScopePath(scope), full);
  }
}

unregisterTemplateFunc("getScopeNamespace");
registerTemplateFunc("getScopeNamespace", getScopeNamespace);

function getJsonContextLocation(outputLocation: string): string {
  outputLocation = sanitizeOutputLocation(outputLocation);
  if (outputLocation) {
    return `${JSONCONTEXT_LOCATION}/${outputLocation}`;
  }

  return `${JSONCONTEXT_LOCATION}`;
}

// @ts-ignore
function getJsonContextNamespace(outputLocation: string = ""): string {
  return [
    getPackageId(),
    ...getJsonContextLocation(outputLocation).split("/"),
  ].join(".");
}

registerTemplateFunc("getJsonContextNamespace", getJsonContextNamespace);
