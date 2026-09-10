require("includes/includes.ts");
require("includes/terraform.ts");

// @ts-ignore
function runJobs(jobs: Job[]) {
  for (const job of jobs) {
    runJob(job);
  }
}

function runJob(job: Job) {
  if (!handleCommonJobs(job)) {
    switch (job.ID) {
      default:
        throw new Error(`Unknown job ID: ${job.ID}`);
    }
  }
}

/**
 * Collects type file definitions from entity TypeDefs and returns template
 * file jobs for generating them under internal/provider/types/. Walks entity
 * AssociatedTypes and Fields via templateTypeDefDataModelFieldType() to
 * discover complex types that need their own Go source files.
 */
function getTypeFileJobs(entities: TerraformEntity[]): TemplateFileJob[] {
  function getSymbol(fieldName: string, typedef: TypeDef): string {
    if (typedef.Extensions?.All?.["Symbol"]) {
      return typedef.Extensions.All["Symbol"];
    }

    throw new Error(
      `Symbol not pre-assigned for type: ${typedef.Name || fieldName}`,
    );
  }

  // Collect type files from entity walks, deduplicating by filename (symbol name).
  const additionalFiles: Record<string, GeneratedDataModelTypeFile> = {};
  for (const entity of entities) {
    entity.SchemaTypeDef.AssociatedTypes.forEach((associatedTypeDef) => {
      const unionName = sanitizeTFUnionTypeName(
        entity.SchemaTypeDef,
        associatedTypeDef,
      );
      const innerType = templateTypeDefDataModelFieldType(
        associatedTypeDef,
        unionName,
        true,
        1,
        "types",
        `${entity.Name}.${unionName}`,
        getSymbol,
      );

      innerType.additionalFiles.forEach(
        (file) => (additionalFiles[file.filename] = file),
      );
    });

    entity.SchemaTypeDef.Fields.forEach((fieldDef) => {
      const innerType = templateTypeDefDataModelFieldType(
        fieldDef.Type,
        fieldDef.Name,
        fieldDef.Optional || fieldDef.Nullable,
        1,
        "types",
        `${entity.Name}.${fieldDef.Name}`,
        getSymbol,
      );

      innerType.additionalFiles.forEach(
        (file) => (additionalFiles[file.filename] = file),
      );
    });
  }

  // Convert collected files to template file jobs.
  const jobs: TemplateFileJob[] = [];
  for (const additionalFile of Object.values(additionalFiles)) {
    if (!additionalFile.filename) {
      continue;
    }
    jobs.push(
      createTemplateFileJob(
        `provider/model.go.stmpl`,
        `internal/provider/types/${sanitizeTFStateName(
          additionalFile.filename,
        )}.go`,
        {
          PackageName: "types",
          Content: additionalFile.content,
          Imports: additionalFile.imports,
        },
      ),
    );
  }
  return jobs;
}

/**
 * Generates the embedded Go SDK used by the Terraform provider.
 */
function generateEmbeddedGoSDK() {
  const goDefaultConfig = getTargetDefaultTemplateConfig("go");
  goDefaultConfig.PackageName = getSDKPackage();
  goDefaultConfig.DefaultErrorName = context.Global.Config.DefaultErrorName;
  goDefaultConfig.UnionStrategy = context.Global.Config.UnionStrategy;
  goDefaultConfig.RespectRequiredFields =
    context.Global.Config.RespectRequiredFields;
  goDefaultConfig.InferUnionDiscriminators =
    context.Global.Config.InferUnionDiscriminators;
  goDefaultConfig.RespectTitlesForPrimitiveUnionMembers =
    context.Global.Config.RespectTitlesForPrimitiveUnionMembers;
  goDefaultConfig.SDKVersion = context.Global.Config.SDKVersion;
  goDefaultConfig.IncludeEmptyObjects =
    context.Global.Config.IncludeEmptyObjects;
  goDefaultConfig.WrapperName = "speakeasy-sdk/terraform";
  goDefaultConfig.SDKHooksConfigAccess = false;
  goDefaultConfig.NullableOptionalWrapper = false;

  interoptTemplateTarget(
    "go",
    "internal/sdk",
    context.Global.AST,
    goDefaultConfig,
    {
      documentation: false,
      module: false,
      tests: false,
    },
  );
}

// @ts-ignore
function postJobs() {
  generateEmbeddedGoSDK();
}

// @ts-ignore
function getJobs(): Job[] {
  const jobs: Job[] = [];

  context.Global.Config.SDKName = "SDK";

  // Template/copy the auxiliary files
  jobs.push(getTemplateAuxiliaryFilesJob());
  jobs.push(...getGeneratedLicenseJobs());

  // Generate UseConfigValue plan modifiers for all types from a single template
  const planModifierTypes = [
    {
      PackageName: "stringplanmodifier",
      TypeName: "String",
      NullConstructor: "types.StringNull()",
    },
    {
      PackageName: "boolplanmodifier",
      TypeName: "Bool",
      NullConstructor: "types.BoolNull()",
    },
    {
      PackageName: "int32planmodifier",
      TypeName: "Int32",
      NullConstructor: "types.Int32Null()",
    },
    {
      PackageName: "int64planmodifier",
      TypeName: "Int64",
      NullConstructor: "types.Int64Null()",
    },
    {
      PackageName: "float32planmodifier",
      TypeName: "Float32",
      NullConstructor: "types.Float32Null()",
    },
    {
      PackageName: "float64planmodifier",
      TypeName: "Float64",
      NullConstructor: "types.Float64Null()",
    },
    {
      PackageName: "numberplanmodifier",
      TypeName: "Number",
      NullConstructor: "types.NumberNull()",
    },
    {
      PackageName: "listplanmodifier",
      TypeName: "List",
      NullConstructor: "types.ListNull(req.PlanValue.ElementType(ctx))",
    },
    {
      PackageName: "setplanmodifier",
      TypeName: "Set",
      NullConstructor: "types.SetNull(req.PlanValue.ElementType(ctx))",
    },
    {
      PackageName: "mapplanmodifier",
      TypeName: "Map",
      NullConstructor: "types.MapNull(req.PlanValue.ElementType(ctx))",
    },
    {
      PackageName: "objectplanmodifier",
      TypeName: "Object",
      NullConstructor: "types.ObjectNull(req.PlanValue.AttributeTypes(ctx))",
    },
  ];
  for (const pmType of planModifierTypes) {
    jobs.push(
      createTemplateFileJob(
        "planmodifier/use_config_value.go.stmpl",
        `internal/planmodifiers/${pmType.PackageName}/use_config_value.go`,
        pmType,
      ),
    );
  }

  // Template test-group-specific files (e.g., tests/review/ for internal testing)
  // These files are only included when generating with a specific test group flag (-t)
  // Note: We use "." as outDir because empty string is falsy in JS, which would
  // skip path replacement and output files to tests/{group}/... instead of root
  const testGroup = context.Global.AST.MainSDK.TestGroup;
  if (testGroup) {
    const groupTestPath = `tests/${testGroup}`;
    if (directoryExists(groupTestPath)) {
      jobs.push(getTemplateDirectoryJob(groupTestPath, "."));
    }
  }

  const Actions = context.Global.AST.TerraformProvider?.Actions ?? [];
  const Resources =
    context.Global.AST.TerraformProvider?.ManagedResources ?? [];
  const DataSources = context.Global.AST.TerraformProvider?.DataResources ?? [];
  const EphemeralResources =
    context.Global.AST.TerraformProvider?.EphemeralResources ?? [];

  for (const entity of (Resources as TerraformEntity[]).concat(
    DataSources,
    EphemeralResources,
    Actions,
  )) {
    logger().Info(`Processing resource ${entity.Name}`);
  }
  logger().Info(
    `Collected ${Resources.length} managed resources: [${Resources.map(
      (r) => `${sanitizeTFStateName(r.Name)}`,
    ).join(", ")}]`,
  );
  logger().Info(
    `Collected ${DataSources.length} data resources: [${DataSources.map(
      (r) => `${sanitizeTFStateName(r.Name)}`,
    ).join(", ")}]`,
  );
  logger().Info(
    `Collected ${
      EphemeralResources.length
    } ephemeral resources: [${EphemeralResources.map(
      (r) => `${sanitizeTFStateName(r.Name)}`,
    ).join(", ")}]`,
  );
  logger().Info(
    `Collected ${Actions.length} action resources: [${Actions.map(
      (r) => `${sanitizeTFStateName(r.Name)}`,
    ).join(", ")}]`,
  );

  // Type file jobs must precede resource template jobs because resource
  // templates reference types from internal/provider/types/.
  jobs.push(
    ...getTypeFileJobs(
      (Actions as TerraformEntity[]).concat(
        Resources,
        DataSources,
        EphemeralResources,
      ),
    ),
  );

  const providerContext: ProviderContext = {
    ...context.Global,
    Actions,
    Resources,
    DataSources,
    EphemeralResources,
  };

  registerDocumentation(providerContext);

  checkEnvVariables(providerContext);

  jobs.push(
    createTemplateFileJob(
      "provider/provider.go.stmpl",
      "internal/provider/provider.go",
      providerContext,
    ),
  );

  jobs.push(
    createTemplateFileJob(
      "examples/provider.tf.stmpl",
      "examples/provider/provider.tf",
      providerContext,
    ),
  );

  const providerResourceContext: ProviderResourceContext = {
    HasConfigureData: hasProviderConfigureData(providerContext),
    SDKType: `*sdk.${sanitizeClassName(providerContext.AST.MainSDK.Type.Name)}`,
  };

  for (const resource of Resources) {
    const resourceContext: ManagedResourceContext = {
      Resource: resource,
      ProviderContext: providerResourceContext,
    };
    const resourceName = sanitizeTFStateName(resource.Name.toLowerCase());

    jobs.push(
      createTemplateFileJob(
        "provider/resource.go.stmpl",
        `internal/provider/${resourceName}_resource.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "provider/resource_sdk.go.stmpl",
        `internal/provider/${resourceName}_resource_sdk.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "examples/resource.tf.stmpl",
        `examples/resources/${resource.TerraformTypeName}/resource.tf`,
        resourceContext,
      ),
    );
  }

  for (const dataSource of DataSources) {
    const resourceContext: DataResourceContext = {
      DataSource: dataSource,
      ProviderContext: providerResourceContext,
    };
    const resourceName = sanitizeTFStateName(dataSource.Name.toLowerCase());

    jobs.push(
      createTemplateFileJob(
        "provider/data_source.go.stmpl",
        `internal/provider/${resourceName}_data_source.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "provider/data_source_sdk.go.stmpl",
        `internal/provider/${resourceName}_data_source_sdk.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "examples/data-source.tf.stmpl",
        `examples/data-sources/${dataSource.TerraformTypeName}/data-source.tf`,
        resourceContext,
      ),
    );
  }

  for (const action of Actions) {
    const resourceContext: ActionContext = {
      Action: action,
      ProviderContext: providerResourceContext,
    };
    const resourceName = sanitizeTFStateName(action.Name.toLowerCase());

    jobs.push(
      createTemplateFileJob(
        "provider/action.go.stmpl",
        `internal/provider/${resourceName}_action.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "provider/action_sdk.go.stmpl",
        `internal/provider/${resourceName}_action_sdk.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "examples/action.tf.stmpl",
        `examples/actions/${action.TerraformTypeName}/action.tf`,
        resourceContext,
      ),
    );
  }

  for (const ephemeralResource of EphemeralResources) {
    const resourceContext: EphemeralResourceContext = {
      EphemeralResource: ephemeralResource,
      ProviderContext: providerResourceContext,
    };
    const resourceName = sanitizeTFStateName(
      ephemeralResource.Name.toLowerCase(),
    );

    jobs.push(
      createTemplateFileJob(
        "provider/ephemeral_resource.go.stmpl",
        `internal/provider/${resourceName}_ephemeral_resource.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "provider/ephemeral_resource_sdk.go.stmpl",
        `internal/provider/${resourceName}_ephemeral_resource_sdk.go`,
        resourceContext,
      ),
    );
    jobs.push(
      createTemplateFileJob(
        "examples/ephemeral-resource.tf.stmpl",
        `examples/ephemeral-resources/${ephemeralResource.TerraformTypeName}/ephemeral-resource.tf`,
        resourceContext,
      ),
    );
  }

  jobs.push(...getReadmeJobs());

  jobs.push(...getGitFilesJobs("go"));

  const contributingFileJob = getTemplateContributingFileJob();
  if (contributingFileJob) {
    jobs.push(contributingFileJob);
  }

  return jobs;
}
