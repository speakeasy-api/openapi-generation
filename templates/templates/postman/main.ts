require("includes/includes.ts");

// @ts-ignore
function runJobs(jobs: Job[]) {
  throw new Error("Not implemented");
}

// @ts-ignore
function getJobs(): Job[] {
  const collection = {
    auth: createAuth(context.Global.AST.MainSDK.Security),
    info: {
      name:
        context.Global.Config.PackageName ||
        context.Global.AST.MainSDK.Type.Name,
      description: context.Global.AST.MainSDK.Comments?.Description || "",
      license:
        context.Global.Config.GeneratedLicense === "agpl"
          ? { name: "AGPL-3.0-only" }
          : undefined,
      schema:
        "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
    },
    item: createItems(context),
    variable: createVariables(context),
  };

  let fileName =
    context.Global.Config.FileName ||
    sanitizeName(
      context.Global.Config.PackageName || context.Global.AST.MainSDK.Type.Name,
    ) + "_postman_collection.json";

  // Start templating from the main SDK file
  templateFile("collection.json.stmpl", fileName, {
    Collection: JSON.stringify(collection, null, 4),
  });

  templateGitignore();

  if (context.Global.Config.GeneratedLicense === "agpl") {
    templateFile("generated-license/LICENSE", "LICENSE", {});
    templateFile("generated-license/NOTICE.stmpl", "NOTICE", {});
  }

  return [];
}
