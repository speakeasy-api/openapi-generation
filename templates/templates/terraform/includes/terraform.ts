/**
 * Union type of all Terraform resource Go AST types. Used where functions need
 * to accept any resource type generically.
 */
type TerraformEntity =
  | TerraformAction
  | TerraformDataResource
  | TerraformEphemeralResource
  | TerraformManagedResource;

/** Terraform resource types, such as action, data, ephemeral, or managed. */
type TerraformResourceType = "action" | "data" | "ephemeral" | "managed";

/**
 * Action entity operation types.
 */
type ActionOperationType = "invoke";

/**
 * Data resource entity operation types.
 */
type DataResourceOperationType = "read";

/**
 * Ephemeral resource entity operation types.
 */
type EphemeralResourceOperationType = "close" | "open";

/**
 * Managed resource entity operation types.
 */
type ManagedResourceOperationType = "create" | "delete" | "read" | "update";

/** Terraform entity/resource operation types that translate into resource
 *  method logic. */
type TerraformResourceOperationType =
  | ActionOperationType
  | DataResourceOperationType
  | EphemeralResourceOperationType
  | ManagedResourceOperationType;

// NOTE: Ensure any changes here are reflected in the target tsconfig.json.

require("./frameworkTypes.ts");
require("./frameworkSchemaPlanModifiers.ts");
require("./frameworkSchemaTypes.ts");
require("./frameworkSchemaDefaults.ts");
require("./frameworkSchemaValidators.ts");
require("./frameworkProviderSchemaTypes.ts");
require("./frameworkActionSchemaTypes.ts");
require("./frameworkDataResourceSchemaTypes.ts");
require("./frameworkEphemeralResourceSchemaTypes.ts");
require("./frameworkManagedResourceSchemaTypes.ts");
require("./tfSanitization.ts");
require("./serverUrl.ts");
require("./tfSecurity.ts");
require("./provider.ts");
require("./resources.ts");
require("./generateValueType.ts");
require("./tfUtils.ts");
require("./generateSchemaTypeUtils.ts");
require("./generateSchemaType.ts");
require("./generateSchemaVersion.ts");
require("./generateTFToSDK.ts");
require("./generateSDKToTF.ts");
require("./generateSDKMethods.ts");
require("./generateInvoker.ts");
require("./generateImportState.ts");
require("./generateEnvVariables.ts");
require("./generateProvider.ts");
require("./generateResource.ts");
require("./generateTFExamples.ts");
require("./documentation.ts");
