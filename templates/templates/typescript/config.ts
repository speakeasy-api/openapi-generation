// @ts-ignore
function upgradeConfig(oldVersion, newVersion, cfg, defaults) {
  return cfg;
}

/** Returns the configuration fields available for customers. This data is
 * fetched early in the generation process so values can be passed back to
 * getGeneratorConfig. */
// @ts-ignore
function getConfigFields(
  commonFields: SDKGenConfigFields,
  newSDK: boolean,
): SDKGenConfigFields {
  return {
    templateVersion: {
      Name: "templateVersion",
      Required: false,
      DefaultValue: "v2",
      Description: "The template version to use",
      ValidationRegex: /v2/.source,
      ValidationMessage: "Template version must be v2",
    },
  };
}

/** Represents the initial target implementation configuration presented to
 * the generator from the target. Customer configuration keys and values from
 * getConfigFields are fetched prior so those values are available. */
// @ts-ignore
function getGeneratorConfig(
  _genConfig: GeneratorInitialConfiguration,
): TargetInitialConfiguration {
  const result: TargetInitialConfiguration = {};

  return result;
}
