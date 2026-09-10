type EnvVar = {
  name: string;
  defaultValue: string;
};

const ENV_VAR_DIRECTIVE = "x-env:";

function getEnvVarFromDirective(input: string): EnvVar | undefined {
  if (typeof input !== "string" || !input.startsWith(ENV_VAR_DIRECTIVE)) {
    return undefined;
  }

  const value = input.substring(ENV_VAR_DIRECTIVE.length).trim();

  const parts = value.split(";");

  if (parts.length == 1) {
    return { name: parts[0].trim(), defaultValue: "" };
  }

  return { name: parts[0].trim(), defaultValue: parts[1].trim() };
}
