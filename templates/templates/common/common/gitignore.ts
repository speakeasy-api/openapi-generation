function getGitFilesJobs(
  fileSuffix: string,
  additionalGitAttributesContent?: string,
): Job[] {
  return [
    {
      ID: "gitignore",
      FileName: ".gitignore",
      Context: null,
    },
    createTemplateFileJob(`gitattributes.stmpl`, ".gitattributes", {
      FileSuffix: fileSuffix,
      AdditionalContent: additionalGitAttributesContent ?? "",
    }),
  ];
}

function templateGitignore() {
  try {
    let existing: string | undefined = readFile(".gitignore");
    let file = templateString(".gitignore.stmpl", {});
    const curLines = existing?.split("\n") ?? [];
    const needLines = file?.split("\n") ?? [];
    const toAdd: string[] = [];
    for (const needed of needLines) {
      const found = curLines.find((line) => line.trim() == needed.trim());
      if (found) {
        continue;
      }
      toAdd.push(needed);
    }
    // Always push ${toAdd} at the beginning of the .gitignore
    const newLines = toAdd.concat(curLines).filter(Boolean);

    if (toAdd.length > 0) {
      writeFile(".gitignore", newLines.join("\n") + "\n", 0, false);
    }
  } catch (e) {
    logger().Warn(
      `failed to modify .gitignore: ${e instanceof Error ? e.message : e}`,
    );
  }
}
