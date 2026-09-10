// @ts-ignore
function templateAuxiliaryFiles(data: any, sourcePath: string = "auxiliary") {
  let files = getDirectoryFiles(sourcePath);

  for (const file of files) {
    // Only include SSE utils if server events feature is used
    if (file.includes("/Utils/Sse/") && !isFeatureUsed("serverEvents")) {
      continue;
    }

    // Only include the open-union interface when the spec has open unions
    if (file.includes("/Utils/IOpenUnion.cs") && !hasAnyOpenUnion()) {
      continue;
    }

    let outFile = file.replace(`${sourcePath}/`, "");

    outFile = templateStringInput("auxPath:" + outFile, outFile, data);
    if (outFile.trim().length == 0) {
      continue;
    }

    if (file.endsWith(".stmpl")) {
      outFile = outFile.replace(".stmpl", "");

      templateFile(file, outFile, context.Global);
    } else {
      copy(file, outFile);
    }
  }
}
