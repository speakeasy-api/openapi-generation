// Load config.ts to make getTemplateDependencies available globally
// @ts-ignore
require("../config.ts");
// @ts-ignore
const deps = getTemplateDependencies();

function javaVersion() {
  return context.Global.Config.LanguageVersion;
}
registerTemplateFunc("javaVersion", javaVersion);

function templateAdditionalPlugins(): string {
  if (context.Global.Config.AdditionalPlugins) {
    const lines = context.Global.Config.AdditionalPlugins;
    return lines.map((line) => `\n    ${line.trim()}`).join("");
  } else {
    return "";
  }
}
registerTemplateFunc("templateAdditionalPlugins", templateAdditionalPlugins);

// @ts-ignore
function templateDependencies(): string {
  // format is scope:groupId:artifactId:version
  const defaultDependencies = [
    `api:com.fasterxml.jackson.core:jackson-annotations:${deps["com.fasterxml.jackson.core:jackson-annotations"].version}`,
    `implementation:com.fasterxml.jackson.core:jackson-databind:${deps["com.fasterxml.jackson.core:jackson-databind"].version}`,
    `implementation:com.fasterxml.jackson.datatype:jackson-datatype-jsr310:${deps["com.fasterxml.jackson.datatype:jackson-datatype-jsr310"].version}`,
    `implementation:com.fasterxml.jackson.datatype:jackson-datatype-jdk8:${deps["com.fasterxml.jackson.datatype:jackson-datatype-jdk8"].version}`,
    ...(isFeatureUsed("nullables")
      ? [
          `api('org.openapitools:jackson-databind-nullable:${deps["org.openapitools:jackson-databind-nullable"].version}') {exclude group: 'com.fasterxml.jackson.core', module: 'jackson-databind'}`,
        ]
      : []),
    `implementation:commons-io:commons-io:${deps["commons-io:commons-io"].version}`,
    `implementation:jakarta.annotation:jakarta.annotation-api:${
      Number(javaVersion()) < 11
        ? deps["jakarta.annotation:jakarta.annotation-api@java8"].version
        : deps["jakarta.annotation:jakarta.annotation-api"].version
    }`,
  ];
  if (useOkHttp()) {
    defaultDependencies.push(
      `implementation:com.squareup.okhttp3:okhttp:${deps["com.squareup.okhttp3:okhttp"].version}`,
    );
  }
  if (slf4jLoggingEnabled()) {
    defaultDependencies.push(
      `api:org.slf4j:slf4j-api:${deps["org.slf4j:slf4j-api"].version}`,
      `testImplementation:org.slf4j:slf4j-simple:${deps["org.slf4j:slf4j-simple"].version}`,
    );
  }
  if (usesPagination()) {
    defaultDependencies.push(
      `implementation:com.jayway.jsonpath:json-path:${deps["com.jayway.jsonpath:json-path"].version}`,
    );
  }
  // reactive-streams backs the reactive async surface on the JDK arm only;
  // the okhttp async arm is CompletableFuture-based and pulls no reactive dep.
  if (asyncEnabled() && !useOkHttp()) {
    defaultDependencies.push(
      `api:org.reactivestreams:reactive-streams:${deps["org.reactivestreams:reactive-streams"].version}`,
    );
  }
  if (sdkHasTests(context.Global.AST)) {
    defaultDependencies.push(
      `testImplementation:org.junit.jupiter:junit-jupiter:${deps["org.junit.jupiter:junit-jupiter"].version}`,
      `testImplementation:org.mockito:mockito-core:${deps["org.mockito:mockito-core"].version}`,
      `testRuntimeOnly:org.junit.platform:junit-platform-launcher:${deps["org.junit.platform:junit-platform-launcher"].version}`,
      `antJUnit:org.apache.ant:ant-junit:${deps["org.apache.ant:ant-junit"].version}`,
    );
  }
  // Reactor only ships for internal test-snippet generation (-t flag in
  // cmd/generate). Customer SDKs never set OutputTests, so they don't pay
  // for these illustrative-only deps.
  if (context.Global.AST.MainSDK.OutputTests && asyncEnabled()) {
    defaultDependencies.push(
      `testImplementation:io.projectreactor:reactor-core:${deps["io.projectreactor:reactor-core"].version}`,
      `testImplementation:io.projectreactor:reactor-test:${deps["io.projectreactor:reactor-test"].version}`,
    );
  }
  let additionalDeps: string[] = [];
  if (context.Global.Config.AdditionalDependencies) {
    additionalDeps = context.Global.Config.AdditionalDependencies;
  }
  const groupIdArtifacts = additionalDeps.map((item) =>
    toGroupIdArtifactId(item),
  );

  // ensure that user specified deps override the default ones
  // (which may include downgrading default versions)
  const trimmedDefaults = defaultDependencies.filter(
    (x) => !groupIdArtifacts.includes(toGroupIdArtifactId(x)),
  );
  const actualDependencies = additionalDeps.concat(trimmedDefaults);

  // render deps for use in build.gradle
  const gradleDeps = actualDependencies
    .map((item) => `\n    ${toGradleDependency(item)}`)
    .join("");
  return `dependencies {${gradleDeps}\n}`;
}
registerTemplateFunc("templateDependencies", templateDependencies);

function usesPagination(): boolean {
  const sdk = context.Global.AST.MainSDK;
  return sdkUsesPagination(sdk);
}

function toGradleDependency(item: string): string {
  if (item.includes("(")) {
    // item is a complex definition (perhaps with exclusions)
    return item;
  } else {
    const arr: string[] = item.split(":");
    const scope: string = arr.shift();
    const dep = arr.join(":");
    return `${scope} '${dep}'`;
  }
}

// returns groupId:artifactId
function toGroupIdArtifactId(item: string): string {
  const arr = item.split(":");
  arr.shift(); // remove scope
  arr.pop(); // remove version
  return arr.join(":");
}
