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

function getUsageModelsFilePath(sourceDir: string): string {
  return `${sourceDir.replaceAll("/", "\\")}\\${getModelsFilePath()}`;
}

registerTemplateFunc("getUsageModelsFilePath", getUsageModelsFilePath);

// @ts-ignore
function getModelsProjectReference(): string {
  const modelsFolder = getModelsTopLevelFolder().split("/").slice(-1)[0];
  return `..\\${modelsFolder}\\${getModelsPackageId()}.csproj`;
}

registerTemplateFunc("getModelsProjectReference", getModelsProjectReference);

// Under netstandard2, Tests bind the netstandard2.0 asset of both SDK and Models,
// so both references carry the same SetTargetFramework. That asset's
// IAsyncEnumerable<T> comes from Microsoft.Bcl.AsyncInterfaces, which the Tests
// restore does not pull in transitively: it flows the SDK's netX dependencies.
// @ts-ignore
function templateTestProjectReferences(): string {
  const projectFilePaths = [getProjectFilePath(), getModelsFilePath()];
  if (!isNetstandard2()) {
    return projectFilePaths
      .map((path) => `<ProjectReference Include="..\\${path}" />`)
      .join("\n    ");
  }
  const references = projectFilePaths.map(
    (path) =>
      `<ProjectReference Include="..\\${path}">
      <SetTargetFramework>TargetFramework=netstandard2.0</SetTargetFramework>
    </ProjectReference>`,
  );
  const asyncInterfaces = getNetstandard2Dependency(
    "Microsoft.Bcl.AsyncInterfaces",
  );
  references.push(
    `<PackageReference Include="${asyncInterfaces.name}" Version="${asyncInterfaces.version}" />`,
  );
  return references.join("\n    ");
}

unregisterTemplateFunc("templateTestProjectReferences");
registerTemplateFunc(
  "templateTestProjectReferences",
  templateTestProjectReferences,
);

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
  return [
    getPackageId(),
    ...getJsonContextLocation(outputLocation).split("/"),
  ].join(".");
}

registerTemplateFunc("getJsonContextNamespace", getJsonContextNamespace);
