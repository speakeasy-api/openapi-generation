function needsReactQuerySection(): boolean {
  return isReactQueryEnabled();
}

registerReadmeSection("react-query", needsReactQuerySection, (sdk: SDK) =>
  templateString("readme/react-query.stmpl", {}),
);
