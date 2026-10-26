// @ts-ignore
function getCommonConfigFields(newSDK: boolean): SDKGenConfigFields {
  // Return any config fields that are shared by the majority of SDKs, will get merged with the config fields from a particular SDK
  return {
    inputModelSuffix: {
      Name: "inputModelSuffix",
      Required: false,
      DefaultValue: "input",
      Description:
        "The suffix to add to models with writeOnly fields that are created as input models",
      ValidationRegex: /^[\w\d.\-_]*$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
      ValidationExamples: {
        valid: ["input", "Input", "request_input", ""],
        invalid: ["input v2", "input/v2", "input!"],
      },
    },
    outputModelSuffix: {
      Name: "outputModelSuffix",
      Required: false,
      DefaultValue: "output",
      Description:
        "The suffix to add to models with writeOnly fields that are created as input models",
      ValidationRegex: /^[\w\d.\-_]*$/.source,
      ValidationMessage: "Letters, numbers, or .-_ only",
      ValidationExamples: {
        valid: ["output", "Output", "request_output", ""],
        invalid: ["output v2", "output/v2", "output!"],
      },
    },
    customCasings: {
      Name: "customCasings",
      Required: false,
      Description:
        "Custom casing terms to preserve in generated PascalCase symbols. Use { initialism: true } to render the key in all caps, for example { mcp: { initialism: true } }.",
    },
    inferUnionDiscriminators: {
      Name: "inferUnionDiscriminators",
      Required: false,
      DefaultValue: newSDK,
      Description:
        "Infer union discriminators for oneOfs missing explicit OpenAPI discriminator mapping",
      ValidationRegex: /^(true|false)$/.source,
      ValidationMessage: "true or false only",
    },
  };
}
