// @ts-ignore
function templateRetractions(): string {
  const retractions = context.Global.Config.Retractions;
  if (!retractions || !Array.isArray(retractions) || retractions.length === 0) {
    return "";
  }

  let retractionsBlock = "\nretract (\n";
  for (const retraction of retractions) {
    // Handle both object and map-like structures
    const version = retraction.version || retraction.Version || "";
    const comment = retraction.comment || retraction.Comment || "";

    if (version) {
      retractionsBlock += `\t${version}`;
      if (comment) {
        retractionsBlock += ` // ${comment}`;
      }
      retractionsBlock += "\n";
    }
  }
  retractionsBlock += ")";

  return retractionsBlock;
}
registerTemplateFunc("templateRetractions", templateRetractions);
