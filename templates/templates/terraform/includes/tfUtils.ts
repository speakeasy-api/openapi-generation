function assertUnreachable(x: never): never {
  throw new Error("unexpected " + JSON.stringify(x));
}

function makeSymbolMananger() {
  const manager: Record<string, boolean> = {};
  manager["break"] = true;
  manager["append"] = true;
  manager["default"] = true;
  manager["func"] = true;
  manager["interface"] = true;
  manager["select"] = true;
  manager["case"] = true;
  manager["defer"] = true;
  manager["go"] = true;
  manager["map"] = true;
  manager["struct"] = true;
  manager["chan"] = true;
  manager["else"] = true;
  manager["goto"] = true;
  manager["package"] = true;
  manager["switch"] = true;
  manager["const"] = true;
  manager["fallthrough"] = true;
  manager["if"] = true;
  manager["range"] = true;
  manager["type"] = true;
  manager["continue"] = true;
  manager["for"] = true;
  manager["import"] = true;
  manager["return"] = true;
  manager["var"] = true;
  manager["float32"] = true;
  manager["float64"] = true;
  manager["int"] = true;
  manager["int32"] = true;
  manager["int64"] = true;
  return manager;
}

function applyPath(curTarget: string, path: any): string {
  for (const pathElement of path.split("/")) {
    if (pathElement == "..") {
      curTarget = curTarget.split(".").slice(0, -2).join(".");
    } else {
      curTarget += "." + sanitizeFieldName(pathElement);
    }
  }
  return curTarget;
}
