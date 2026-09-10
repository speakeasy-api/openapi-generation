function createPath(path: string) {
  if (!path) {
    throw new Error("Path is required");
  }
  if (path == "/") return undefined;

  return path
    .split("/")
    .filter((part) => part != null && part != "")
    .map((part) => {
      if (part.startsWith("{")) {
        return `:${part.replace("{", "").replace("}", "")}`;
      }
      return part;
    });
}
