// @ts-ignore
function templateClientCredentialsTokenDeserializeV2(): string {
  return "JsonSerializer.Deserialize(responseBody, ClientCredentialsHookJsonContext.Default.TokenResponse)";
}

unregisterTemplateFunc("templateClientCredentialsTokenDeserialize");
registerTemplateFunc(
  "templateClientCredentialsTokenDeserialize",
  templateClientCredentialsTokenDeserializeV2,
);

// @ts-ignore
function templateClientCredentialsHookContextV2(): string {
  const converters = getJsonContextConverters()
    .map((c) => `        typeof(${c})`)
    .join(",\n    ");
  return [
    "[JsonSerializable(typeof(TokenResponse))]",
    "[JsonSourceGenerationOptions(",
    "    DefaultIgnoreCondition = JsonIgnoreCondition.WhenWritingNull,",
    "    GenerationMode = JsonSourceGenerationMode.Default,",
    "    Converters = new[]",
    "    {",
    converters,
    "    }",
    ")]",
    "internal partial class ClientCredentialsHookJsonContext : JsonSerializerContext { }",
  ].join("\n    ");
}

unregisterTemplateFunc("templateClientCredentialsHookContext");
registerTemplateFunc(
  "templateClientCredentialsHookContext",
  templateClientCredentialsHookContextV2,
);
