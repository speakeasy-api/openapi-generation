// @ts-ignore
function getModelsPackageId(): string {
  return `${getPackageId()}.${MODELS_PROJECT_NAME}`;
}

registerTemplateFunc("getModelsPackageId", getModelsPackageId);

// @ts-ignore
function getModelsTopLevelFolder(): string {
  return `${getSDKTopLevelFolder()}.${MODELS_PROJECT_NAME}`;
}

registerTemplateFunc("getModelsTopLevelFolder", getModelsTopLevelFolder);

function getModelsFilePath(): string {
  const winPath = getModelsTopLevelFolder().replaceAll("/", "\\");
  return `${winPath}\\${getModelsPackageId()}.csproj`;
}

registerTemplateFunc("getModelsFilePath", getModelsFilePath);

// @ts-ignore
function getModelsProjectReference(): string {
  const modelsFolder = getModelsTopLevelFolder().split("/").slice(-1)[0];
  return `..\\${modelsFolder}\\${getModelsPackageId()}.csproj`;
}

registerTemplateFunc("getModelsProjectReference", getModelsProjectReference);

function stripModelsPrefix(location: string): string {
  if (location === MODELS_PROJECT_NAME) {
    return "";
  }

  const prefix = `${MODELS_PROJECT_NAME}/`;
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
    getModelsPackageId(),
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
  return getAccessNamespace(
    getJsonContextLocation(outputLocation),
    true,
    false,
    getPackageId(),
  );
}
