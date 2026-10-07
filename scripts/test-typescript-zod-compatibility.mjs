import { execFileSync } from "node:child_process";
import {
  cpSync,
  mkdtempSync,
  mkdirSync,
  readFileSync,
  readdirSync,
  rmSync,
  symlinkSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

const [sdkPath, ...requestedVersions] = process.argv.slice(2);
if (!sdkPath) {
  throw new Error(
    "Usage: node scripts/test-typescript-zod-compatibility.mjs <generated-sdk> [zod-versions...]",
  );
}

const sdk = resolve(sdkPath);
const typesPath = "src/sdk/types";
const testPath = "src/__tests__/smart_union_additional.test.ts";
const mini = readFileSync(
  join(sdk, typesPath, "smartUnion.ts"),
  "utf8",
).includes('"zod/v4-mini"');
const versions = requestedVersions.length
  ? requestedVersions
  : [mini ? "3.25.65" : "3.25.1", "3", "4.0.0", "4"];
const failures = [];

for (const version of versions) {
  const directory = mkdtempSync(join(tmpdir(), "zod-compatibility-"));
  try {
    const packed = JSON.parse(
      execFileSync(
        "npm",
        ["pack", `zod@${version}`, "--ignore-scripts", "--json"],
        {
          cwd: directory,
          encoding: "utf8",
          stdio: ["ignore", "pipe", "inherit"],
        },
      ),
    )[0];
    console.log(`Checking Zod ${packed.version}`);
    const modules = join(directory, "node_modules");
    mkdirSync(join(modules, "zod"), { recursive: true });
    execFileSync("tar", [
      "-xzf",
      join(directory, packed.filename),
      "--strip-components=1",
      "-C",
      join(modules, "zod"),
    ]);
    for (const dependency of readdirSync(join(sdk, "node_modules"))) {
      if (dependency !== "zod") {
        symlinkSync(
          join(sdk, "node_modules", dependency),
          join(modules, dependency),
        );
      }
    }
    writeFileSync(
      join(directory, "package.json"),
      JSON.stringify({ private: true, type: "module" }),
    );
    mkdirSync(join(directory, typesPath), { recursive: true });
    for (const name of [
      "smartUnion",
      "decimal",
      "defaultToZeroValue",
      "discriminatedUnion",
      "rfcdate",
      "unrecognized",
    ]) {
      cpSync(
        join(sdk, typesPath, `${name}.ts`),
        join(directory, typesPath, `${name}.ts`),
      );
    }
    mkdirSync(join(directory, "src/__tests__"), { recursive: true });
    cpSync(join(sdk, testPath), join(directory, testPath));
    writeFileSync(
      join(directory, "tsconfig.json"),
      JSON.stringify({
        compilerOptions: {
          target: "es2022",
          module: "nodenext",
          strict: true,
          skipLibCheck: true,
          noEmit: true,
        },
        include: [testPath],
      }),
    );
    for (const [command, args] of [
      ["tsc", ["--project", "tsconfig.json"]],
      ["vitest", ["run", testPath]],
    ]) {
      execFileSync(join(sdk, "node_modules/.bin", command), args, {
        cwd: directory,
        stdio: "inherit",
      });
    }
  } catch (error) {
    console.error(`Zod ${version} compatibility failed: ${error.message}`);
    failures.push(version);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
}

if (failures.length) {
  throw new Error(`Zod compatibility failed for: ${failures.join(", ")}`);
}
