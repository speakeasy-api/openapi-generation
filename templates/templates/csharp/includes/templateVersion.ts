// @ts-ignore
function isCsharpV2(): boolean {
  return context.Global.Config.TemplateVersion == "v2";
}

registerTemplateFunc("isCsharpV2", isCsharpV2);
