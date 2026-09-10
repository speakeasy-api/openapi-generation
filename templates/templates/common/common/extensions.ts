// This file describes various x-speakeasy OpenAPI Specification extensions and
// their configuration.

/**
 * Describes x-speakeasy-entity-operation extension configuration from the AST
 * in Operations.Extensions.EntityOperation. The extension is typically
 * configured along with the x-speakeasy-entity extension to build a single
 * model to describe an API entity. That model is used to describe a Terraform
 * data, ephemeral, or managed resource currently, but may represent other
 * target entities in the future.
 */
type ExtEntityOperationConfig = {
  TerraformActions: ExtEntityOperationV1Config[];
  TerraformDataResources: ExtEntityOperationV1Config[];
  TerraformEphemeralResources: ExtEntityOperationV1Config[];
  TerraformManagedResources: ExtEntityOperationV1Config[];
};

/**
 * Describes an individual entity operation configuration from the
 * x-speakeasy-entity-operation extension.
 */
type ExtEntityOperationV1Config = {
  /**
   * Name of the entity for the operation.
   */
  Entity: string;

  /**
   * Type of the entity operation, such as "create", "read", "update", or
   * "delete".
   */
  OperationTypes: string[];

  /**
   * Optional order of the entity operation. These must be unique across all
   * entity operations for the same entity. If not specified, the order is
   * assumed to be first.
   */
  Order?: number;

  /**
   * SDK options for the entity operation.
   */
  Options?: ExtEntityOperationV1Options;

  /**
   * Returns a string representation of the entity operation configuration.
   */
  String(): string;
};

/**
 * Describes SDK options for entity operations.
 */
type ExtEntityOperationV1Options = {
  /**
   * Polling configuration for the entity operation.
   */
  Polling?: ExtEntityOperationV1Polling;

  /**
   * Patch configuration for update operations.
   */
  Patch?: ExtEntityOperationV1Patch;
};

/**
 * Describes polling configuration for entity operations.
 */
type ExtEntityOperationV1Polling = {
  /**
   * Overrides the number of seconds before the first request.
   */
  DelaySeconds?: number;

  /**
   * Overrides the number of seconds between requests.
   */
  IntervalSeconds?: number;

  /**
   * Overrides the number of requests to limit polling.
   */
  LimitCount?: number;

  /**
   * Name of the polling option to use.
   */
  Name: string;
};

/**
 * Describes patch configuration for update operations.
 */
type ExtEntityOperationV1Patch = {
  /**
   * Style of patch semantics to use for updates.
   * Valid values: "only-send-changed-attributes"
   */
  Style: string;
};
