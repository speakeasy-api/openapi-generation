interface StringMapper {
  (s: string): string;
}

namespace JavaTypeParser {
  // an example of use is to take a type like
  //
  //   java.util.Map<? super java.lang.String, java.util.List<? extends mine.Thing>>[]
  //
  // and remove the qualifications (for example using javaImport function) to get
  //
  //   Map<? super String, List<? extends Thing>>[]
  //
  export function convert(javaType: string, mapper: StringMapper): string {
    return parse(javaType).render(mapper);
  }

  // performs a unit test of JavaTypeParser
  export function test() {
    const assertEquals = (expected, value) => {
      if (expected !== value) {
        throw new Error(`Expected ${expected} but was ${value}`);
      }
    };
    assertEquals("boo", splitByComma("boo")[0]);
    assertEquals(
      "java.lang.StringM",
      convert("java.lang.String", (x) => x + "M"),
    );
    assertEquals(
      "StringM[]",
      convert("String[]", (x) => x + "M"),
    );
    assertEquals(
      "java.util.HashMapM<java.lang.StringM, MapM<ThingM, OtherM[]>>",
      convert(
        "java.util.HashMap<java.lang.String, Map<Thing, Other[]>>",
        (x) => x + "M",
      ),
    );
    assertEquals(
      "java.util.HashMapM<java.lang.StringM, MapM<ThingM, OtherM[]>>",
      convert(
        "  java.util.HashMap  <  java.lang.String  , Map<Thing  ,   Other[]  >  >  ",
        (x) => x + "M",
      ),
    );
  }

  enum Wildcard {
    NONE,
    EXTENDS,
    SUPER,
  }

  class Cls {
    name: string;
    genericTypes: Array<Cls>;
    wildcard: Wildcard;
    isArray: boolean;

    constructor(
      name: string,
      genericTypes: Array<Cls>,
      wildcard: Wildcard,
      isArray: boolean,
    ) {
      this.name = name;
      this.genericTypes = genericTypes;
      this.wildcard = wildcard;
      this.isArray = isArray;
    }

    render(nameMapper: StringMapper): string {
      let prefix: string;
      if (this.wildcard == Wildcard.NONE) {
        prefix = "";
      } else if (this.wildcard == Wildcard.EXTENDS) {
        prefix = "? extends ";
      } else {
        prefix = "? super ";
      }
      let b = "";
      b += prefix;
      b += nameMapper(this.name);
      if (this.genericTypes.length > 0) {
        b += "<";
        b += this.genericTypes.map((cls) => cls.render(nameMapper)).join(", ");
        b += ">";
      }
      if (this.isArray) {
        b += "[]";
      }
      return b;
    }
  }

  function parse(s: string): Cls {
    s = s.trim();
    const isArray = s.endsWith("[]");
    let wildcard: Wildcard;
    if (s.startsWith("? extends ")) {
      wildcard = Wildcard.EXTENDS;
      s = s.substring("? extends ".length);
    } else if (s.startsWith("? super ")) {
      s = s.substring("? super ".length);
      wildcard = Wildcard.SUPER;
    } else {
      wildcard = Wildcard.NONE;
    }
    let name: string;
    const i = s.indexOf("<");
    if (i == -1) {
      if (isArray) {
        name = s.substring(0, s.length - 2).trim();
      } else {
        name = s;
      }
      return new Cls(name, [], wildcard, isArray);
    } else {
      name = s.substring(0, i).trim();
      const j = s.lastIndexOf(">");
      if (j == -1) {
        throw new Error("type has bad syntax: " + s);
      } else {
        const list = parseMany(s.substring(i + 1, j));
        return new Cls(name, list, wildcard, isArray);
      }
    }
  }

  function parseMany(s: string): Array<Cls> {
    s = s.trim();
    const items = splitByComma(s);
    return items.map((x) => parse(x));
  }

  // just splits by the top-level commas (ignores nested ones)
  function splitByComma(s: string): Array<string> {
    let depth = 0;
    const list: Array<string> = [];
    let b = "";
    for (let i = 0; i < s.length; i++) {
      const ch = s[i];
      if (ch == "," && depth == 0) {
        list.push(b.trim());
        b = "";
      } else {
        b += ch;
        if (ch == "<") {
          depth++;
        } else if (ch == ">") {
          depth--;
        }
      }
    }
    list.push(b.trim());
    if (depth > 0) {
      throw new Error("syntax error, angle brackets not closed: " + s);
    }
    return list;
  }
}

JavaTypeParser.test();

interface StringPredicate {
  apply(value: string): boolean;
}

// For a given generated class this class decides whether
// a class reference in the generated class needs to be
// fully qualified (with package name) or not and accumulates
// imports for the class for inclusion in the imports list of
// the generated class.
class Imports {
  // imports will are designed to be included in the
  // class corresponding to this name
  private fullClassName: string;

  // cached simple name
  private simple: string;

  // map of full class name to qualified/simple name
  private resolved: Map<string, string> = new Map<string, string>();

  private staticImports: Set<string> = new Set<string>();

  // contains the packages of star imports
  private starImports: Set<string> = new Set<string>();

  private simpleNames: Set<string> = new Set<string>();

  // we need to know what classes are in the same package as
  // fullClassName because generation expects that a class in the
  // same package does not need to be fully qualified. Can simply set
  // to x => false if not set up yet (and cross your fingers about clashes)
  private simpleNameInPackage: StringPredicate;

  // TODO this can be true always (i.e removed) once we pass an accurate simpleNameInPackage function
  private trimJavaLang: boolean;

  private static readonly WILDCARD_IMPORT_THRESHOLD = 2;

  constructor(
    fullClassName: string,
    simpleNameInPackage: StringPredicate,
    trimJavaLang: boolean,
  ) {
    this.fullClassName = fullClassName;
    this.simple = Imports.simpleName(fullClassName);
    this.simpleNameInPackage = simpleNameInPackage;
    this.trimJavaLang = trimJavaLang;
    this.resolved.set(fullClassName, Imports.simpleName(fullClassName));
    JavaTypeParser.test();
  }

  add(fullClassName: string): string {
    const simple = Imports.simpleName(fullClassName);
    const currentMapping = this.resolved.get(fullClassName);
    if (currentMapping == undefined) {
      if (
        this.simpleNames.has(simple) ||
        this.simpleNameInPackage.apply(simple) ||
        simple == this.simple
      ) {
        // class will need to be fully qualified
        this.resolved.set(fullClassName, fullClassName);
        return fullClassName;
      } else {
        this.resolved.set(fullClassName, simple);
        this.simpleNames.add(simple);
        return simple;
      }
    } else {
      return currentMapping;
    }
  }

  addStar(packageName: string): string {
    this.starImports.add(packageName);
    return "";
  }

  addStatic(packageNameAndMethod: string): string {
    this.staticImports.add(packageNameAndMethod);
    return Imports.simpleName(packageNameAndMethod);
  }

  render(useWildcardImports: boolean): string {
    const stars = [...this.starImports.keys()].map((x) => `\nimport ${x}.*;`);
    const owningPackage = Imports.packageName(this.fullClassName);
    const statics = [...this.staticImports.keys()].map(
      (x) => `\nimport static ${x};`,
    );
    let explicitList = [...this.resolved.entries()]
      // exclude classes that were fully qualified
      .filter(([k, v]) => k !== v && k !== this.fullClassName)
      // exclude java.lang classes that don't need qualification
      // unless there's a name clash
      .filter(([k, _]) => !this.trimJavaLang || !k.startsWith("java.lang."))
      .filter(([k, _]) => !this.starImports.has(Imports.packageName(k)))
      // don't need to include imports for classes in the same package
      // note that same package classes have priority over java.lang classes
      // which in theory also don't need to be qualified if there is no clash
      // with classes in the owning package
      .filter(([k, _]) => Imports.packageName(k) !== owningPackage)
      .map(([k, _]) => k)
      .sort();
    if (useWildcardImports) {
      // only use wildcard imports when >= 3 classes imported from a package

      // first group sorted explicit imports by package
      const temp: string[][] = [];
      {
        let pkg = "";
        let list: string[] = [];
        for (const cls of explicitList) {
          const p = Imports.packageName(cls);
          if (p !== pkg) {
            if (list.length > 0) {
              temp.push(list);
              list = [];
            }
            pkg = p;
          }
          list.push(cls);
        }
        if (list.length > 0) {
          temp.push(list);
        }
      }
      explicitList = temp.flatMap((list) => {
        const pkg = Imports.packageName(list[0]);
        if (pkg == "java.lang") {
          // even if not explicitly used the existence of one class in say models.shared that
          // has the same name as a java.lang class will provoke a compile error. Because we
          // don't yet have knowledge of all classes in a package (used or not) we are defensive
          // and always include explicit java.lang imports
          return list;
        } else if (list.length > Imports.WILDCARD_IMPORT_THRESHOLD) {
          const result: string[] = [Imports.packageName(list[0]) + ".*"];
          const clashWithJavaLang = list.filter((cls) =>
            javaLangClassNames.has(Imports.simpleName(cls)),
          );
          result.push(...clashWithJavaLang);
          return result;
        } else {
          return list;
        }
      });
    }
    const explicit = explicitList.map((k) => `\nimport ${k};`);
    const hasNonStaticImports = explicit.length > 0 || stars.length > 0;
    if (statics.length > 0 && hasNonStaticImports) {
      // ensure there is a blank line between static imports
      // and non-static imports (convention)
      statics.push("\n");
    }
    const result = statics.concat(explicit.concat(stars).sort()).join("");
    if (result.length > 0) {
      // if there are imports then we add a blank line afterwards
      return result + "\n\n";
    }
    return "";
  }

  className(): string {
    return this.fullClassName;
  }

  private static packageName(fullClassName: string): string {
    const i = fullClassName.lastIndexOf(".");
    if (i == -1) {
      return "";
    } else {
      return fullClassName.substring(0, i);
    }
  }

  private static simpleName(fullClassName: string): string {
    const i = fullClassName.lastIndexOf(".");
    if (i == -1) {
      return fullClassName;
    } else {
      return fullClassName.substring(i + 1);
    }
  }
}

const javaImportsStack: Imports[] = [];

function genImportsWildcarded(fullClassName: string): string {
  const useWildcard = !context.Global.Config.ExplicitDocImports;
  return genImports(fullClassName, useWildcard);
}
registerTemplateFunc("genImportsWildcarded", genImportsWildcarded);

// @ts-ignore
function genImports(
  fullClassName: string,
  useWildcardImports: boolean = false,
): string {
  // no calls to javaImport should be made before genImports is run
  // (we need to have a new instance of Imports on the stack first)
  javaImportsStack.push(new Imports(fullClassName, () => false, false));
  return `{{ templateImports ${useWildcardImports}}}`;
}
registerTemplateFunc("genImports", genImports);

function templateImports(useWildcardImports: boolean = false): string {
  const imports = javaImportsStack.pop();
  return imports.render(useWildcardImports);
}
registerTemplateFunc("templateImports", templateImports);

/**
 * Adds an import for a fully qualified java type (with generics) to
 * the imports list for the class being written and returns the
 * resolved (abbreviated or not) class name for use in the class
 * being written.
 *
 * For example `java.util.List<java.lang.String>` might add (if no conflicts)
 * `java.util.List` and `java.lang.String` to the imports list and return
 * `List<String>`.
 *
 * @param javaType - fully qualified java type name with generics.
 *     For example `java.util.List<java.lang.String>`
 * @param isLocal - prefixes the javaType with the base generated package name
 *     before processing
 * @param additionalContext - additional context to be passed to the template
 * @returns the resolved type
 */
function javaImport(
  javaType: string,
  isLocal: boolean = false,
  additionalContext?: TemplateValueContext,
): string {
  switch (true) {
    case javaType.includes("<") || javaType.includes("["):
      if (isLocal) {
        throw new Error(
          "cannot call javaImport with isLocal=true if type is an array or has generics",
        );
      } else {
        return JavaTypeParser.convert(javaType, (s) => javaImport(s));
      }
    case javaImportsStack.length == 0:
      // this path happens in some markdown generation
      // where genImports is never called (which adds a new Imports
      // object to the javaImportsStack).
      // In that case we just don't bother doing anything because
      // out markdown generation does its own qualification
      // stripping/ignoring
      return javaType;
    case additionalContext?.outputLocation == "tests":
      // Special case for input/output classes in test helpers
      return javaType.split(".").pop();
    default:
      const imports = javaImportsStack[javaImportsStack.length - 1];
      const prefix = isLocal ? templatePackageName() + "." : "";
      // route JDK-transport / Java 9+ types through the okhttp / Java 8
      // compat seam (no-op on the default arm) — see compat.ts
      return imports.add(mapCompatJavaType(prefix + javaType));
  }
}

registerTemplateFunc("javaImport", javaImport);

function javaImportAsync(type: string): string {
  // place type under `.async` sub-package
  // eg:
  //    before: org.openapis.openapi.models.operations.Foobar
  //    after:  org.openapis.openapi.models.operations.async.Foobar
  const parts = type.split(".");
  const className = parts.pop();
  const asyncType = [...parts, "async", className].join(".");

  return javaImport(asyncType);
}

registerTemplateFunc("javaImportAsync", javaImportAsync);

function javaImportLocal(partialClassName: string): string {
  return javaImport(partialClassName, true);
}

registerTemplateFunc("javaImportLocal", javaImportLocal);

function javaImportString(): string {
  return javaImport("java.lang.String");
}

registerTemplateFunc("javaImportString", javaImportString);

function javaImportJsonNullable(): string {
  return javaImport("org.openapitools.jackson.nullable.JsonNullable");
}

registerTemplateFunc("javaImportJsonNullable", javaImportJsonNullable);

function javaImportException(): string {
  return javaImport("java.lang.Exception");
}

registerTemplateFunc("javaImportException", javaImportException);

function javaImportOptional(): string {
  return javaImport("java.util.Optional");
}
registerTemplateFunc("javaImportOptional", javaImportOptional);

function javaImportDeprecated(): string {
  return javaImport("java.lang.Deprecated");
}
registerTemplateFunc("javaImportDeprecated", javaImportDeprecated);

function javaImportOverride(): string {
  return javaImport("java.lang.Override");
}
registerTemplateFunc("javaImportOverride", javaImportOverride);

function javaImportSuppressWarnings(): string {
  return javaImport("java.lang.SuppressWarnings");
}
registerTemplateFunc("javaImportSuppressWarnings", javaImportSuppressWarnings);

function javaImportThrowable(): string {
  return javaImport("java.lang.Throwable");
}
registerTemplateFunc("javaImportThrowable", javaImportThrowable);

function javaImportNullable(): string {
  return javaImport("jakarta.annotation.Nullable");
}
registerTemplateFunc("javaImportNullable", javaImportNullable);

function javaImportUtils(): string {
  return javaImport("utils.Utils", true);
}
registerTemplateFunc("javaImportUtils", javaImportUtils);

function javaImportPaginationEntity(entity: string): string {
  return javaImport("utils.pagination." + entity, true);
}
registerTemplateFunc("javaImportPaginationEntity", javaImportPaginationEntity);

function javaImportMap(): string {
  return javaImport("java.util.Map");
}
registerTemplateFunc("javaImportMap", javaImportMap);

function javaImportList(): string {
  return javaImport("java.util.List");
}
registerTemplateFunc("javaImportList", javaImportList);

function javaImportHashMap(): string {
  return javaImport("java.util.HashMap");
}
registerTemplateFunc("javaImportHashMap", javaImportHashMap);

function javaImportJacksonAnn(simpleClassName: string): string {
  return javaImport("com.fasterxml.jackson.annotation." + simpleClassName);
}
registerTemplateFunc("javaImportJacksonAnn", javaImportJacksonAnn);

function javaImportStar(packageName: string): string {
  const imports = javaImportsStack[javaImportsStack.length - 1];
  imports.addStar(packageName);
  return "";
}
registerTemplateFunc("javaImportStar", javaImportStar);

function javaImportStatic(
  packageNameAndMethod: string,
  isLocal: boolean = false,
): string {
  const imports = javaImportsStack[javaImportsStack.length - 1];
  const prefix = isLocal ? templatePackageName() + "." : "";
  imports.addStatic(prefix + packageNameAndMethod);
  const i = packageNameAndMethod.lastIndexOf(".");
  const methodName = packageNameAndMethod.substring(i + 1);
  return methodName;
}
registerTemplateFunc("javaImportStatic", javaImportStatic);

function javaImportAssertionsMethod(methodName: string): string {
  const imports = javaImportsStack[javaImportsStack.length - 1];
  return imports.addStatic("org.junit.jupiter.api.Assertions." + methodName);
}

function javaImportConstructorParam(
  field: FieldDef,
  isAsync: boolean = false,
): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return getConstructorMetadata(parameterForField(field, isAsync)).paramType;
  }

  return javaImport(
    sanitizeType(
      field.Type,
      field.Optional && !field.IsAdditionalProperties,
      field.Nullable && !field.IsAdditionalProperties,
      true,
      isAsync,
    ),
  );
}

function javaImportField(field: FieldDef, isAsync: boolean = false): string {
  if (context.Global.Config.NullFriendlyParameters) {
    const typeInfo = getFieldTypeInfo(parameterForField(field, isAsync));
    return javaImport(typeInfo.fieldType);
  }

  return javaImport(
    sanitizeType(
      field.Type,
      field.Optional && !field.IsAdditionalProperties,
      field.Nullable && !field.IsAdditionalProperties,
      true,
      isAsync,
    ),
  );
}

registerTemplateFunc("javaImportField", javaImportField);

function javaImportFieldMetadata(
  fieldMetadata: FieldMetadata,
  isAsync: boolean = false,
): string {
  if (!context.Global.Config.NullFriendlyParameters) {
    return javaImportField(fieldMetadata.javaParam.Field, isAsync);
  }

  return javaImport(fieldMetadata.fieldType);
}

registerTemplateFunc("javaImportFieldMetadata", javaImportFieldMetadata);

function javaImportFieldNonPrimitive(
  field: FieldDef,
  isAsync: boolean = false,
  forceNullFriendly: boolean = false,
): string {
  if (context.Global.Config.NullFriendlyParameters || forceNullFriendly) {
    const param = parameterForField(field, isAsync);
    if (forceNullFriendly) {
      // Inner type of Optional<INNER> for error accessors. Must line up with
      // map/flatMap choice in templateDataFieldTransform.
      const { boxedType, nullableType } = getFieldTypeInfo(param);
      if (errorAccessorUsesFlatMap(field) || isRawGetter()) {
        return javaImport(toNonPrimitive(boxedType));
      }
      return javaImport(
        toNonPrimitive(field.Nullable ? nullableType! : boxedType),
      );
    }
    const typ = getBuilderMetadata(param).fieldType;
    return javaImport(typ);
  }

  return javaImport(
    toNonPrimitive(
      sanitizeType(
        field.Type,
        field.Optional && !field.IsAdditionalProperties,
        field.Nullable && !field.IsAdditionalProperties,
        true,
        isAsync,
      ),
    ),
  );
}
registerTemplateFunc(
  "javaImportFieldNonPrimitive",
  javaImportFieldNonPrimitive,
);

function javaImportFieldNoWildcard(
  field: FieldDef,
  isAsync: boolean = false,
  style: GetterStyle = context.Global.Config.GetterStyle,
): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return javaImport(
      getGetterMetadata(parameterForField(field, isAsync), style).returnType,
    );
  }

  const optional = field.Optional && !field.IsAdditionalProperties;
  const nullable = field.Nullable && !field.IsAdditionalProperties;
  if (isRawGetter(style) && (optional || nullable)) {
    return templateNullableAnnotation(
      toNonPrimitive(sanitizeTypeBase(field.Type, false, isAsync)),
    );
  }
  if (
    isAlwaysOptionalGetter(style) &&
    optional === nullable &&
    !field.IsAdditionalProperties
  ) {
    return javaImport(
      `java.util.Optional<${toNonPrimitive(
        sanitizeTypeBase(field.Type, false, isAsync),
      )}>`,
    );
  }

  return javaImport(
    sanitizeType(field.Type, optional, nullable, false, isAsync),
  );
}
registerTemplateFunc("javaImportFieldNoWildcard", javaImportFieldNoWildcard);

function javaImportTypeMandatory(type: TypeDef): string {
  return javaImport(sanitizeTypeMandatory(type));
}
registerTemplateFunc("javaImportTypeMandatory", javaImportTypeMandatory);

function javaImportTypeMandatoryNonPrimitive(type: TypeDef): string {
  return javaImport(toNonPrimitive(sanitizeTypeMandatory(type)));
}
registerTemplateFunc(
  "javaImportTypeMandatoryNonPrimitive",
  javaImportTypeMandatoryNonPrimitive,
);

function javaImportType(type: TypeDef, boxed: boolean = false): string {
  let typeStr = sanitizeTypeMandatory(type);
  if (boxed) {
    typeStr = toNonPrimitive(typeStr);
  }
  return javaImport(typeStr);
}
registerTemplateFunc("javaImportType", javaImportType);

function javaImportTypeAsync(type: TypeDef, boxed: boolean = false): string {
  if (requiresDistinctAsyncType(type)) {
    let modelNamespace = getModelNamespace(
      true,
      "",
      type.OutputLocation,
      sanitizeClassName(type.Name),
      true,
    );
    if (boxed) {
      modelNamespace = toNonPrimitive(modelNamespace);
    }
    return javaImport(modelNamespace);
  }
  return javaImportType(type, boxed);
}
registerTemplateFunc("javaImportTypeAsync", javaImportTypeAsync);

function javaImportReturnType(type: TypeDef, isAsync: boolean = false): string {
  return isAsync
    ? `${javaImport(
        "java.util.concurrent.CompletableFuture",
      )}<${javaImportTypeAsync(type)}>`
    : javaImportType(type);
}

function sanitizeTypeMandatory(
  type: TypeDef,
  isAsync: boolean = false,
): string {
  return sanitizeType(type, false, false, true, isAsync);
}
registerTemplateFunc("sanitizeTypeMandatory", sanitizeTypeMandatory);

function javaImportFieldBase(
  field: FieldDef,
  isAsync: boolean = false,
): string {
  return javaImport(sanitizeTypeBase(field.Type, true, isAsync));
}
registerTemplateFunc("javaImportFieldBase", javaImportFieldBase);

// @ts-ignore
function getScopePath(scope: string): string {
  const imports = context.Global.Config.Imports;
  switch (scope.toString()) {
    case "errors":
      return imports.GetErrorsPath();
    case "shared":
      return imports.GetSharedPath();
    case "operations":
      return imports.GetOperationsPath();
    case "callbacks":
      return imports.GetCallbacksPath();
    case "webhooks":
      return imports.GetWebhooksPath();
  }
  return "models";
}

registerTemplateFunc("getScopePath", getScopePath);

// @ts-ignore
function getModelNamespace(
  full: boolean,
  scope: string,
  scopePath: string = "",
  className: string = "",
  isAsync: boolean = false,
): string {
  const parts = [];
  if (full) {
    parts.push(templatePackageName());
  }
  if (scope) {
    scopePath = getScopePath(scope);
  }
  if (scopePath) {
    scopePath = sanitizeOutputLocation(scopePath);
    parts.push(scopePath.replaceAll("/", "."));
  }
  if (isAsync) {
    parts.push("async");
  }
  if (className) {
    parts.push(className);
  }

  return parts.join(".");
}

registerTemplateFunc("getModelNamespace", getModelNamespace);

/**
 * Centralized helper to get the model package path.
 * Prefers outputLocation if available, otherwise falls back to scope.
 *
 * @param full - Whether to include the full package name (with base package prefix)
 * @param scope - The scope (e.g., "shared", "operations") to use as fallback
 * @param outputLocation - The output location path (takes precedence over scope if provided)
 * @param className - Optional class name to append to the package path
 * @param isAsync - Whether to append ".async" to the package path
 * @returns The resolved package path
 */
function getModelPackage(
  full: boolean,
  scope: string,
  outputLocation: string = "",
  className: string = "",
  isAsync: boolean = false,
): string {
  if (outputLocation) {
    return getModelNamespace(full, "", outputLocation, className, isAsync);
  }
  return getModelNamespace(full, scope, "", className, isAsync);
}

registerTemplateFunc("getModelPackage", getModelPackage);

// taken from Java 17
const javaLangClassNames = new Set<string>([
  "AbstractMethodError",
  "AbstractStringBuilder",
  "Appendable",
  "ApplicationShutdownHooks",
  "ArithmeticException",
  "ArrayIndexOutOfBoundsException",
  "ArrayStoreException",
  "AssertionError",
  "AssertionStatusDirectives",
  "AutoCloseable",
  "Boolean",
  "BootstrapMethodError",
  "Byte",
  "CharacterData00",
  "CharacterData01",
  "CharacterData02",
  "CharacterData03",
  "CharacterData0E",
  "CharacterData",
  "CharacterDataLatin1",
  "CharacterDataPrivateUse",
  "CharacterDataUndefined",
  "Character",
  "CharacterName",
  "CharSequence",
  "ClassCastException",
  "ClassCircularityError",
  "ClassFormatError",
  "Class",
  "ClassLoader",
  "ClassNotFoundException",
  "ClassValue",
  "Cloneable",
  "CloneNotSupportedException",
  "Comparable",
  "Compiler",
  "ConditionalSpecialCasing",
  "Deprecated",
  "Double",
  "EnumConstantNotPresentException",
  "Enum",
  "Error",
  "ExceptionInInitializerError",
  "Exception",
  "FdLibm",
  "Float",
  "FunctionalInterface",
  "IllegalAccessError",
  "IllegalAccessException",
  "IllegalArgumentException",
  "IllegalCallerException",
  "IllegalMonitorStateException",
  "IllegalStateException",
  "IllegalThreadStateException",
  "IncompatibleClassChangeError",
  "IndexOutOfBoundsException",
  "InheritableThreadLocal",
  "InstantiationError",
  "InstantiationException",
  "Integer",
  "InternalError",
  "InterruptedException",
  "Iterable",
  "LayerInstantiationException",
  "LinkageError",
  "LiveStackFrameInfo",
  "LiveStackFrame",
  "Long",
  "Math",
  "Module",
  "ModuleLayer",
  "NamedPackage",
  "NegativeArraySizeException",
  "NoClassDefFoundError",
  "NoSuchFieldError",
  "NoSuchFieldException",
  "NoSuchMethodError",
  "NoSuchMethodException",
  "NullPointerException",
  "NumberFormatException",
  "Number",
  "Object",
  "OutOfMemoryError",
  "Override",
  "package-info",
  "Package",
  "ProcessBuilder",
  "ProcessEnvironment",
  "ProcessHandleImpl",
  "ProcessHandle",
  "ProcessImpl",
  "Process",
  "PublicMethods",
  "Readable",
  "Record",
  "ReflectiveOperationException",
  "Runnable",
  "RuntimeException",
  "Runtime",
  "RuntimePermission",
  "SafeVarargs",
  "SecurityException",
  "SecurityManager",
  "Short",
  "Shutdown",
  "StackFrameInfo",
  "StackOverflowError",
  "StackStreamFactory",
  "StackTraceElement",
  "StackWalker",
  "StrictMath",
  "StringBuffer",
  "StringBuilder",
  "StringCoding",
  "StringConcatHelper",
  "StringIndexOutOfBoundsException",
  "String",
  "StringLatin1",
  "StringUTF16",
  "SuppressWarnings",
  "System",
  "Terminator",
  "ThreadDeath",
  "ThreadGroup",
  "Thread",
  "ThreadLocal",
  "Throwable",
  "TypeNotPresentException",
  "UnknownError",
  "UnsatisfiedLinkError",
  "UnsupportedClassVersionError",
  "UnsupportedOperationException",
  "VerifyError",
  "VersionProps",
  "VirtualMachineError",
  "Void",
  "WeakPairMap",
]);

function inJavaLang(importLine: string): boolean {
  const s = importLine.trim();
  const i = s.lastIndexOf(".");
  if (i != 0) {
    const simpleClassName = s.substring(i + 1, s.length - 1);
    return javaLangClassNames.has(simpleClassName);
  } else {
    return false;
  }
}

function javaImportBoxed(type: string): string {
  return javaImport(toNonPrimitive(type));
}

registerTemplateFunc("javaImportBoxed", javaImportBoxed);

function javaImportNullableIfRequired(
  p: JavaParam,
  force: boolean = false,
  elideTypeParams: boolean = false,
): string {
  const type = p.Type;

  if (elideTypeParams) {
    return stripGenerics(javaImport(type));
  }
  return useNullableWrapperForParam(p, force)
    ? `${javaImportJsonNullable()}<${javaImportBoxed(type)}>`
    : javaImport(type);
}
registerTemplateFunc(
  "javaImportNullableIfRequired",
  javaImportNullableIfRequired,
);

function javaImportRequest(
  param: JavaParam,
  elideTypeParams: boolean = false,
): string {
  if (context.Global.Config.NullFriendlyParameters) {
    return javaImportBoxed(
      javaImportNullableIfRequired(param, true, elideTypeParams),
    );
  }
  return javaImportBoxed(param.EnhancedType);
}
registerTemplateFunc("javaImportRequest", javaImportRequest);

function javaImportNonRetryableException(): string {
  if (asyncEnabled()) {
    return javaImportLocal("utils.NonRetryableException");
  }
  return javaImportStatic("utils.Retries.NonRetryableException", true);
}
registerTemplateFunc(
  "javaImportNonRetryableException",
  javaImportNonRetryableException,
);

function javaImportRetryableException(): string {
  if (asyncEnabled()) {
    return javaImportLocal("utils.RetryableException");
  }
  return javaImportStatic("utils.Retries.RetryableException", true);
}
registerTemplateFunc(
  "javaImportRetryableException",
  javaImportRetryableException,
);

function javaImportAsyncParam(param: JavaParam): string {
  if (param.Field) {
    // Use the field if available and apply async + boxed conversion
    return javaImport(
      toNonPrimitive(sanitizeTypeMandatory(param.Field.Type, true)),
    );
  }
  // Fall back to the param type and ensure it's boxed
  return javaImport(toNonPrimitive(param.Type));
}
registerTemplateFunc("javaImportAsyncParam", javaImportAsyncParam);

function javaImportReactivePublisher(): string {
  return javaImport("org.reactivestreams.Publisher");
}
registerTemplateFunc(
  "javaImportReactivePublisher",
  javaImportReactivePublisher,
);

function javaImportReactiveSubscriber(): string {
  return javaImport("org.reactivestreams.Subscriber");
}
registerTemplateFunc(
  "javaImportReactiveSubscriber",
  javaImportReactiveSubscriber,
);

function javaImportFlowPublisher(): string {
  return `${javaImport("java.util.concurrent.Flow")}.Publisher`;
}
registerTemplateFunc("javaImportFlowPublisher", javaImportFlowPublisher);

function javaImportBaseSDK(isAsync: boolean = false): string {
  let sdkClassName = sanitizeClassName(context.Global.AST.MainSDK.Type.Name);
  if (isAsync) {
    sdkClassName = `Async${sdkClassName}`;
  }
  return javaImport(sdkClassName, true);
}
registerTemplateFunc("javaImportBaseSDK", javaImportBaseSDK);

// Helper function to create SSE return type for async operations (now supports both SSE and JSONL)
function javaImportResponseReturnType(
  operation: Operation,
  isAsync: boolean,
): string {
  if (isAsync && isStreamingOperation(operation)) {
    if (useOkHttp()) {
      // The okhttp arm's async streaming surface is a CompletableFuture of
      // the blocking iterator types; the reactive EventStream backs the JDK
      // arm only.
      return `${javaImport(
        "java.util.concurrent.CompletableFuture",
      )}<${javaImportLocal(
        isSSEOperation(operation) ? "utils.EventStream" : "utils.JsonLStream",
      )}<${getStreamingItemType(operation)}>>`;
    }
    const responseType = javaImportTypeAsync(operation.Response.Type);
    const itemType = getStreamingItemType(operation);
    return `${javaImportLocal(
      "utils.reactive.EventStream",
    )}<${responseType}, ${itemType}>`;
  }

  // Fall back to normal return type
  return javaImportReturnType(operation.Response.Type, isAsync);
}

registerTemplateFunc(
  "javaImportResponseReturnType",
  javaImportResponseReturnType,
);
