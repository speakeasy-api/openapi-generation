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
    ...commonFields,
    packageName: {
      Name: "packageName",
      Required: true,
      DefaultValue: "openapi",
      Description:
        "The name of the Postman collection. This show as the name when imported into Postman. This is also used as the file name in `{example}_postman_collection.json` if no file name is provided.",
      ValidationRegex: /^[\w\d\-~]([\w\d.\-_\/~]*[\w\d\-~])?$/.source,
      ValidationMessage:
        "Letters, numbers, or /.-_~ only. Cannot start or end with slash or dot.",
    },
    fileName: {
      Name: "fileName",
      Required: false,
      DefaultValue: null,
      Description:
        "The collection file name. If not file name is provided the packageName is used in the `{example}_postman_collection.json` if no file name is provided.",
      ValidationRegex: /^[\w\d\-~]([\w\d.\-_\/~]*[\w\d\-~])?$/.source,
      ValidationMessage:
        "Letters, numbers, or /._~ only. Cannot start or end with slash or dot.",
    },
    imports: {
      Name: "imports",
      Required: false,
      DefaultValue: {
        option: "openapi",
        paths: {
          shared: "shared",
          operations: "operations",
          errors: "errors",
          callbacks: "callbacks",
          webhooks: "webhooks",
        },
      },
      Description: "Configuration for model import structure",
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
