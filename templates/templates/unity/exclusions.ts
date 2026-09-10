// @ts-ignore
function setupExclusions(sdkClassName: string) {
  addHeaderPattern(".*\\.cs$");
  addHeaderPattern(".*\\.usage\\.cs$", true);
}
