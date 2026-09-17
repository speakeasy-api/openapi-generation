function templateAuxiliaryFile(data: any, sourceDir: string, filePath: string) {
  let outFile = filePath.replace(`${sourceDir}/`, "");
  const templateName = `auxPath:${outFile}`;

  outFile = templateStringInput(templateName, outFile, data);
  if (outFile.trim().length == 0) {
    return;
  }

  if (filePath.endsWith(".stmpl")) {
    outFile = outFile.replace(".stmpl", "");
    templateFile(filePath, outFile, context.Global);
  } else {
    copy(filePath, outFile);
  }
}

function templateAuxiliarySourceDir(data: any, sourceDir: string) {
  getDirectoryFiles(sourceDir).map((f) => {
    if (f.includes("/Utils/Sse/") && !isFeatureUsed("serverEvents")) {
      return;
    }
    templateAuxiliaryFile(data, sourceDir, f);
  });
}

// @ts-ignore
function templateAuxiliaryFiles(data: any, _unused: string) {
  templateAuxiliarySourceDir(data, "auxiliary/v2");

  // AOT-safe Utils cherry picked from csharp v1 templates
  const v1UtilsFiles: [string, string][] = [
    ["{{getDefaultHttpClient}}", getDefaultHttpClient()],
  ];

  [
    "Constants",
    "Retries/RetryConfig",
    "Retries/BackoffStrategy",
    "Retries/Retries",
  ].map((f) => v1UtilsFiles.push([f, f]));

  if (isFeatureUsed("serverEvents")) {
    ["Sse/SseEvent", "Sse/SseParseException", "Sse/SseStreamParser"].map((f) =>
      v1UtilsFiles.push([f, f]),
    );
  }

  const v1Utils = `auxiliary/{{ getSDKTopLevelFolder }}/Utils`;
  const sdkFolder = getSDKTopLevelFolder();
  for (const [src, out] of v1UtilsFiles) {
    templateFile(
      `${v1Utils}/${src}.cs.stmpl`,
      `${sdkFolder}/Utils/${out}.cs`,
      context.Global,
    );
  }

  templateFile(
    `${v1Utils}/OptionalNullable.cs.stmpl`,
    `${getModelsTopLevelFolder()}/OptionalNullable.cs`,
    context.Global,
  );
}
