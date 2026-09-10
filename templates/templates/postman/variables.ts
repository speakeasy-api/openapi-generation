function createVariables(context: Context) {
  // Populate Global Parameters as postman variables
  const variables = [];

  if (context.Global.AST.MainSDK.Globals?.Fields) {
    for (const param of context.Global.AST.MainSDK.Globals.Fields) {
      if (param.Default) {
        variables.push({
          key: param.Name,
          value: param.Default,
        });
      } else if (param.Type?.Examples.length > 0) {
        variables.push({
          key: param.Name,
          value: getExampleValue(
            faker.helpers.arrayElement(param.Type?.Examples),
          ),
        });
      } else {
        variables.push({ key: param.Name, value: parseValue(param) });
      }
    }
  }

  if (context.Global.AST.MainSDK.Servers?.Servers.length > 0) {
    variables.push({
      key: "baseUrl",
      value: context.Global.AST.MainSDK.Servers.Servers[0].URL.replaceAll(
        "{",
        "{{",
      ).replaceAll("}", "}}"),
    });

    for (const variable of context.Global.AST.MainSDK.Servers.Servers[0]
      .Variables) {
      variables.push({ key: variable.Name, value: variable.Default || "" });
    }
  }

  return variables;
}
