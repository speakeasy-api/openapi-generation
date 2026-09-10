const systemReservedClassNames = ["Application", "Enum", "Console"];

type UnityModel = {
  Name: string;
  Type: TypeDef;
  OutputLocation: string;
};

// @ts-ignore
function getModelJobs(types: BucketedTypes, path: string): TemplateFileJob[] {
  const jobs: TemplateFileJob[] = [];
  for (const [outputLocation, models] of sequencedMapEntries(types)) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const typeDef of types) {
        let modelName = typeDef.Name;
        const ctx: UnityModel = {
          Name: modelName,
          Type: typeDef,
          OutputLocation: outputLocation,
        };

        jobs.push(
          createTemplateFileJob(
            `modelfile.cs.stmpl`,
            `${path}/${getModelsLocation(outputLocation)}/${sanitizeFileName(
              modelName,
            )}.cs`,
            ctx,
          ),
        );
      }
    }
  }

  return jobs;
}

// @ts-ignore
function doesTypeConflict(type: TypeDef): boolean {
  var className = sanitizeClassName(type.Name, false);

  if (systemReservedClassNames.includes(className)) {
    return true;
  }

  if (doesTypeConflictWithSDK(className, context.Global.AST.MainSDK)) {
    return true;
  }

  const typeBuckets: Record<string, TypeDef[]> = {};

  for (const [outputLocation, models] of sequencedMapEntries(
    context.Global.AST.BucketedTypes,
  )) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const t of types) {
        if (!typeBuckets[outputLocation]) {
          typeBuckets[outputLocation] = [];
        }

        typeBuckets[outputLocation].push(t);
      }
    }
  }

  for (const bucketName in typeBuckets) {
    if (bucketName == type.OutputLocation) {
      continue;
    }

    const bucket = typeBuckets[bucketName];

    for (const i in bucket) {
      const t = bucket[i];
      if (sanitizeClassName(t.Name, false) == className) {
        return true;
      }
    }
  }

  return false;
}

function doesTypeConflictWithSDK(className: string, sdk: SDK): boolean {
  if (className == sanitizeSDKName(sdk.Type.Name)) {
    return true;
  }

  for (const subSDK of sdk.SubSDKs) {
    if (doesTypeConflictWithSDK(className, subSDK)) {
      return true;
    }
  }

  return false;
}
