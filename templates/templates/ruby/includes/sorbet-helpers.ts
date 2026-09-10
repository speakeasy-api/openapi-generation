function sorbetCast(
  command: string,
  type: string,
  assign: boolean = false,
): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return assign ? "" : command;
  }
  const cmd = `T.cast(${command}, ${type})`;
  if (assign) {
    return `${command} = ${cmd}`;
  }
  return cmd;
}
registerTemplateFunc("sorbetCast", sorbetCast);

function sorbetMust(command: string, assign: boolean = false): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return assign ? "" : command;
  }
  const cmd = `T.must(${command})`;
  if (assign) {
    return `${command} = ${cmd}`;
  }

  return cmd;
}
registerTemplateFunc("sorbetMust", sorbetMust);

function sorbetLet(
  command: string,
  type: string,
  assign: boolean = false,
): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return assign ? "" : command;
  }
  const cmd = `T.let(${command}, ${type})`;
  if (assign) {
    return `${command} = ${cmd}`;
  }
  return cmd;
}
registerTemplateFunc("sorbetLet", sorbetLet);

function sorbetUnsafe(command: string): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return command;
  }
  return `T.unsafe(${command})`;
}

registerTemplateFunc("sorbetUnsafe", sorbetUnsafe);

function sorbetEnabled(): boolean {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return false;
  }
  return true;
}
registerTemplateFunc("sorbetEnabled", sorbetEnabled);

function sorbetTSig(): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return "";
  }
  return "extend T::Sig";
}
registerTemplateFunc("sorbetTSig", sorbetTSig);

function typedDirective(): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return "# typed: false";
  }
  return "# typed: true";
}
registerTemplateFunc("typedDirective", typedDirective);

function sorbetAbstract(): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return "";
  }
  return "abstract!";
}
registerTemplateFunc("sorbetAbstract", sorbetAbstract);

function sorbetImport(): string {
  if (context.Global.Config.TypingStrategy !== "sorbet") {
    return "";
  }
  return "require 'sorbet-runtime'";
}
registerTemplateFunc("sorbetImport", sorbetImport);
