// @ts-ignore
function getTestDirectory(): string {
  return "tests";
}

/** Returns the mock server directory based on the tests output directory. */
// @ts-ignore (ignore duplicate definitions)
function getMockServerDirectory(): string {
  const testDirectory = getTestDirectory();
  return `${testDirectory}/mockserver`;
}

registerTemplateFunc("getMockServerDirectory", getMockServerDirectory);
