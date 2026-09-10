function checkEnvVariables(providerContext: ProviderContext) {
  const envVariables = context.Global.Config.EnvironmentVariables;

  if (!envVariables) {
    return "";
  }

  if (!Array.isArray(envVariables)) {
    throw new Error("Environment variables must be an array of objects");
  }

  const invalidKeys = [];
  const providerAttributes = getProviderAttributes(providerContext);

  for (const envVariable of envVariables) {
    if (!envVariable.env || !envVariable.providerAttribute) {
      throw new Error(
        `failed to make environment variable mapping: environmentVariables expected to be an array of { env: "MY_ENV_VARIABLE", providerAttribute: "my_provider_attribute" } objects. Got ${JSON.stringify(
          envVariables,
        )}`,
      );
    }
    if (!providerAttributes[envVariable.providerAttribute]) {
      invalidKeys.push(envVariable.providerAttribute);
    }
  }
  if (invalidKeys.length) {
    throw new Error(
      `failed to make environment variable mapping: the following specified provider attributes were not found: [${invalidKeys.join(
        ", ",
      )}]`,
    );
  }

  return;
}

type ValidEnvVariables = {
  env: string;
  providerAttribute: string;
}[];

function getEnvironmentVariable(key: string): string | undefined {
  const envVariables: ValidEnvVariables =
    context.Global.Config.EnvironmentVariables;

  if (!envVariables) {
    return undefined;
  }

  const envVariable = envVariables.find(
    (v: any) => v.providerAttribute === key,
  );

  return envVariable ? envVariable.env : undefined;
}
