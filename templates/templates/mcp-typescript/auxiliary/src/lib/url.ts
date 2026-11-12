const hasOwn = Object.prototype.hasOwnProperty;

export type Params = Partial<Record<string, string | number>>;

// WHATWG URL parsing strips tab and newline characters, treats "\" as "/" in
// special schemes, and resolves "%2e" as "." when removing dot segments.
const ignoredURLCharsRE = /[\t\n\r]/g;
const segmentSeparatorRE = /[/\\?#]/;
const dotSegmentRE = /^(?:\.|%2e){1,2}$/i;

function hasDotSegment(value: string): boolean {
  return value
    .replace(ignoredURLCharsRE, "")
    .split(segmentSeparatorRE)
    .some((part) => dotSegmentRE.test(part));
}

export function pathToFunc(
  pathPattern: string,
  options?: { charEncoding?: "percent" | "none" },
): (params?: Params) => string {
  const paramRE = /\{([a-zA-Z0-9_]+?)\}/g;

  return function buildURLPath(params: Record<string, unknown> = {}): string {
    return pathPattern
      .split("/")
      .map((segment) => {
        const placeholders: string[] = [];
        const rendered = segment.replace(paramRE, function (_, placeholder) {
          if (!hasOwn.call(params, placeholder)) {
            throw new Error(`Parameter '${placeholder}' is required`);
          }

          const value = params[placeholder];
          if (typeof value !== "string" && typeof value !== "number") {
            throw new Error(
              `Parameter '${placeholder}' must be a string or number`,
            );
          }

          placeholders.push(placeholder);
          return options?.charEncoding === "percent"
            ? encodeURIComponent(`${value}`)
            : `${value}`;
        });

        if (placeholders.length > 0 && hasDotSegment(rendered)) {
          const names = placeholders.join("', '");
          throw new Error(
            `Parameter '${names}' must not form a '.' or '..' path segment`,
          );
        }

        return rendered;
      })
      .join("/");
  };
}
