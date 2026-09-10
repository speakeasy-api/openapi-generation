require("common/examples.ts");

// @ts-ignore
function getPrecalculatedExamples(
  mainSDK: SDK | null,
  preCalculatedExamples: Examples | null,
): Examples | null {
  return preCalculateExamples(mainSDK, preCalculatedExamples);
}
