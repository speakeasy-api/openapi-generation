// @ts-ignore
function getModelJobs(types: BucketedTypes, path: string): Job[] {
  const jobs: Job[] = [];

  for (const [outputLocation, models] of sequencedMapEntries(types)) {
    for (const [, types] of sequencedMapEntries(models)) {
      for (const typeDef of types) {
        let modelName = typeDef.Name;

        const fileDir = `${path}.${MODELS_PROJECT_NAME}/${getModelsLocation(
          outputLocation,
        )}`;

        const ctx: CSharpModel = {
          Name: modelName,
          Type: typeDef,
          OutputLocation: outputLocation,
        };

        jobs.push(
          createTemplateFileJob(
            `modelfile.cs.stmpl`,
            `${fileDir}/${sanitizeFileName(modelName)}.cs`,
            ctx,
          ),
        );
      }
    }
  }

  return jobs;
}

function getCsharpV2Jobs(): Job[] {
  return [...getJsonContextJobs(), ...getJsonOptionsJobs()];
}
