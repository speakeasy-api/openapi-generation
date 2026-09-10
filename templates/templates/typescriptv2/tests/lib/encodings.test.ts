import { describe, expect, it } from "vitest";

import {
  encodeMatrix,
  encodeLabel,
  encodeForm,
  encodeSimple,
  encodeSpaceDelimited,
  encodePipeDelimited,
  encodeDeepObject,
  encodeJSON,
} from "../../lib/encodings.js";

describe("matrix encoding", () => {
  const encode = encodeMatrix;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual(";color");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual(";color=blue");
    expect(encode("color", 1)).toEqual(";color=1");
    expect(encode("color", true)).toEqual(";color=true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      ";color=blue,black,brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], { explode: true }),
    ).toEqual(";color=blue;color=black;color=brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      ";color=R,100,G,200,B,150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual(";R=100;G=200;B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual(";dates=2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      ";dates=2006-01-02T15:04:05.099Z,2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      ";dates=start,2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual(";color%5B1%5D=3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual(";color=blue,bl%40ck,brown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual(";color=R,100,G,200,B,%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual(";color=blue,black,brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual(";color=blue;color=black;color=brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual(";color=R,100,G,200,B,150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual(";R=100;G=200;B=150");
  });
});

describe("label encoding", () => {
  const encode = encodeLabel;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual(".");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual(".blue");
    expect(encode("color", 1)).toEqual(".1");
    expect(encode("color", true)).toEqual(".true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      ".blue.black.brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], { explode: true }),
    ).toEqual(".blue.black.brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      ".R.100.G.200.B.150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual(".R=100.G=200.B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual(".2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      ".2006-01-02T15:04:05.099Z.2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      ".start.2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual(".3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual(".blue.bl%40ck.brown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual(".R.100.G.200.B.%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual(".blue.black.brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual(".blue.black.brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual(".R.100.G.200.B.150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual(".R=100.G=200.B=150");
  });
});

describe("form encoding", () => {
  const encode = encodeForm;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual("color=");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual("color=blue");
    expect(encode("color", 1)).toEqual("color=1");
    expect(encode("color", true)).toEqual("color=true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      "color=blue,black,brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], { explode: true }),
    ).toEqual("color=blue&color=black&color=brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      "color=R,100,G,200,B,150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual("R=100&G=200&B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual("dates=2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      "dates=2006-01-02T15:04:05.099Z,2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      "dates=start,2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual("color%5B1%5D=3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual("color=blue%2Cbl%40ck%2Cbrown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual("color=R%2C100%2CG%2C200%2CB%2C%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual("color=blue,black,brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual("color=blue&color=black&color=brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual("color=R,100,G,200,B,150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual("R=100&G=200&B=150");
  });
});

describe("simple encoding", () => {
  const encode = encodeSimple;

  it("percent-encodes with charEncoding percent", () => {
    expect(encode("id", "key:v1/sub", { charEncoding: "percent" })).toEqual(
      "key%3Av1%2Fsub",
    );
  });

  it("passes reserved characters through with charEncoding percentExceptReserved", () => {
    expect(
      encode("id", "key:v1/sub", { charEncoding: "percentExceptReserved" }),
    ).toEqual("key:v1/sub");
    expect(
      encode("id", ":/?#[]@!$&'()*+,;=", {
        charEncoding: "percentExceptReserved",
      }),
    ).toEqual(":/?#[]@!$&'()*+,;=");
  });

  it("still percent-encodes non-reserved characters with charEncoding percentExceptReserved", () => {
    expect(
      encode("id", "a b%c:d", { charEncoding: "percentExceptReserved" }),
    ).toEqual("a%20b%25c:d");
  });

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual("");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual("blue");
    expect(encode("color", 1)).toEqual("1");
    expect(encode("color", true)).toEqual("true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      "blue,black,brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], { explode: true }),
    ).toEqual("blue,black,brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      "R,100,G,200,B,150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual("R=100,G=200,B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual("2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      "2006-01-02T15:04:05.099Z,2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      "start,2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual("3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual("blue,bl%40ck,brown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual("R,100,G,200,B,%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual("blue,black,brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual("blue,black,brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual("R,100,G,200,B,150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual("R=100,G=200,B=150");
  });
});

describe("spaceDelimited encoding", () => {
  const encode = encodeSpaceDelimited;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual("color=");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual("color=blue");
    expect(encode("color", 1)).toEqual("color=1");
    expect(encode("color", true)).toEqual("color=true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      "color=blue black brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], {
        explode: true,
      }),
    ).toEqual("color=blue&color=black&color=brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      "color=R 100 G 200 B 150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual("R=100&G=200&B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual("dates=2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      "dates=2006-01-02T15:04:05.099Z 2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      "dates=start 2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual("color%5B1%5D=3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual("color=blue%20bl%40ck%20brown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual("color=R%20100%20G%20200%20B%20%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual("color=blue black brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual("color=blue&color=black&color=brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual("color=R 100 G 200 B 150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual("R=100&G=200&B=150");
  });
});

describe("pipeDelimited encoding", () => {
  const encode = encodePipeDelimited;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual("color=");
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual("color=blue");
    expect(encode("color", 1)).toEqual("color=1");
    expect(encode("color", true)).toEqual("color=true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      "color=blue|black|brown",
    );
  });

  it("encodes exploded arrays", () => {
    expect(
      encode("color", ["blue", "black", "brown"], {
        explode: true,
      }),
    ).toEqual("color=blue&color=black&color=brown");
  });

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      "color=R|100|G|200|B|150",
    );
  });

  it("encodes exploded objects", () => {
    expect(
      encode("color", { R: 100, G: 200, B: 150 }, { explode: true }),
    ).toEqual("R=100&G=200&B=150");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const d2 = new Date("2007-02-03T16:05:06.099Z");

    expect(encode("dates", d1)).toEqual("dates=2006-01-02T15:04:05.099Z");

    expect(encode("dates", [d1, d2])).toEqual(
      "dates=2006-01-02T15:04:05.099Z|2007-02-03T16:05:06.099Z",
    );

    expect(encode("dates", { start: d1 })).toEqual(
      "dates=start|2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode("color[1]", "3.5%", {
        charEncoding: "percent",
      }),
    ).toEqual("color%5B1%5D=3.5%25");

    expect(
      encode("color", ["blue", "bl@ck", "brown"], {
        charEncoding: "percent",
      }),
    ).toEqual("color=blue%7Cbl%40ck%7Cbrown");

    expect(
      encode("color", { R: 100, G: 200, B: "$" }, { charEncoding: "percent" }),
    ).toEqual("color=R%7C100%7CG%7C200%7CB%7C%24");
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();

    expect(
      encode("color", ["blue", "black", undefined, null, "brown"]),
    ).toEqual("color=blue|black|brown");
    expect(
      encode("color", ["blue", "black", undefined, null, "brown"], {
        explode: true,
      }),
    ).toEqual("color=blue&color=black&color=brown");

    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual("color=R|100|G|200|B|150");
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: 150, A: undefined, X: null },
        { explode: true },
      ),
    ).toEqual("R=100&G=200&B=150");
  });
});

describe("deepObject encoding", () => {
  const encode = encodeDeepObject;

  it("encodes objects", () => {
    expect(encode("color", { R: 100, G: 200, B: 150 })).toEqual(
      "color[R]=100&color[G]=200&color[B]=150",
    );
  });

  it("encodes array fields", () => {
    expect(encode("filters", { themes: ["light", "dark"] })).toEqual(
      "filters[themes]=light&filters[themes]=dark",
    );
  });

  it("encodes empty fields", () => {
    expect(encode("filters", { themes: [] })).toEqual("");
    expect(encode("color", {})).toEqual("");
  });

  it("encodes date objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");

    expect(encode("dates", { start: d1 })).toEqual(
      "dates[start]=2006-01-02T15:04:05.099Z",
    );
  });

  it("encodes special characters", () => {
    expect(
      encode(
        "color",
        { R: 100, G: 200, B: "3.5%" },
        { charEncoding: "percent" },
      ),
    ).toEqual("color%5BR%5D=100&color%5BG%5D=200&color%5BB%5D=3.5%25");
  });

  it("disallows simple parameters", () => {
    const msg = `Value of parameter 'color' which uses deepObject encoding must be an object`;
    expect(() => encode("color", "")).toThrow(msg);
    expect(() => encode("color", "blue")).toThrow(msg);
    expect(() => encode("color", 1)).toThrow(msg);
    expect(() => encode("color", true)).toThrow(msg);
  });

  it("encodes object fields", () => {
    expect(encode("color", { R: 100, G: { msg: "nested" }, B: 150 })).toEqual(
      "color[R]=100&color[G][msg]=nested&color[B]=150",
    );
  });

  it("discards undefined values", () => {
    expect(encode("color", undefined)).toBeUndefined();
    expect(encode("color", null)).toBeUndefined();
    expect(
      encode("color", { R: 100, G: 200, B: 150, A: undefined, X: null }),
    ).toEqual("color[R]=100&color[G]=200&color[B]=150");
    expect(
      encode("filters", { themes: [undefined, null, "light", "dark"] }),
    ).toEqual("filters[themes]=light&filters[themes]=dark");
  });
});

describe("json encoding", () => {
  const encode = encodeJSON;

  it("encodes empty parameters", () => {
    expect(encode("color", "")).toEqual('color=""');
  });

  it("encodes simple parameters", () => {
    expect(encode("color", "blue")).toEqual('color="blue"');
    expect(encode("color", 1)).toEqual("color=1");
    expect(encode("color", true)).toEqual("color=true");
  });

  it("encodes arrays", () => {
    expect(encode("color", ["blue", "black", "brown"])).toEqual(
      'color=["blue","black","brown"]',
    );
  });

  it("encodes objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const val = {
      colors: ["light", "dark"],
      inStock: true,
      pub: { gt: d1 },
    };
    const expected = JSON.stringify(val);

    expect(
      encode("filters", {
        colors: ["light", "dark"],
        inStock: true,
        pub: { gt: d1 },
      }),
    ).toEqual(`filters=${expected}`);
  });

  it("encodes exploded objects", () => {
    const d1 = new Date("2006-01-02T15:04:05.099Z");
    const val = {
      colors: ["light", "dark"],
      inStock: true,
      pub: { gt: d1 },
    };
    const expected = JSON.stringify(val);

    expect(
      encode(
        "filters",
        {
          colors: ["light", "dark"],
          inStock: true,
          pub: { gt: d1 },
        },
        { explode: true },
      ),
    ).toEqual(expected);
  });

  it("discards undefined values", () => {
    expect(encode("filters", undefined)).toBeUndefined();
    expect(encode("filters", null)).toEqual("filters=null");
    expect(
      encode(
        "filters",
        {
          price: undefined,
          colors: [undefined, "light", "dark"],
          theme: { light: undefined, dark: "#000000" },
        },
        { explode: true },
      ),
    ).toEqual(`{"colors":[null,"light","dark"],"theme":{"dark":"#000000"}}`);
  });
});

describe("allowReserved query encoding", () => {
  const reserved = { charEncoding: "percentExceptReserved" } as const;
  const percent = { charEncoding: "percent" } as const;

  describe("scalars", () => {
    it("passes reserved characters through", () => {
      expect(encodeForm("q", "a:b/c", reserved)).toEqual("q=a:b/c");
    });

    it("still encodes characters outside the reserved set", () => {
      expect(encodeForm("q", "a b", reserved)).toEqual("q=a%20b");
    });

    it("encodes a literal percent sign", () => {
      expect(encodeForm("q", "100%", reserved)).toEqual("q=100%25");
    });

    it("does not decode a pre-existing percent escape", () => {
      expect(encodeForm("q", "a%2Fb", reserved)).toEqual("q=a%252Fb");
    });

    it("percent-encodes multi-byte UTF-8", () => {
      expect(encodeForm("q", "é☃", reserved)).toEqual("q=%C3%A9%E2%98%83");
    });
  });

  describe("arrays", () => {
    it("passes reserved characters through each item when not exploded", () => {
      expect(encodeForm("q", ["a:b", "c/d"], reserved)).toEqual("q=a:b,c/d");
    });

    it("passes reserved characters through each item when exploded", () => {
      expect(
        encodeForm("q", ["a:b", "c/d"], { ...reserved, explode: true }),
      ).toEqual("q=a:b&q=c/d");
    });
  });

  describe("objects", () => {
    it("applies allowReserved to member names serialized into the value", () => {
      expect(encodeForm("q", { "k:1": "v/1" }, reserved)).toEqual("q=k:1,v/1");
    });

    it("always encodes member names that become query parameter names", () => {
      expect(
        encodeForm("q", { "k:1": "v/1" }, { ...reserved, explode: true }),
      ).toEqual("k%3A1=v/1");
    });

    it("encodes deepObject parameter names but not their values", () => {
      expect(encodeDeepObject("filter", { "k:1": "v/1" }, reserved)).toEqual(
        "filter%5Bk%3A1%5D=v/1",
      );
    });
  });

  describe("delimited styles", () => {
    // Space and "|" are not RFC 3986 reserved characters, so the separators
    // are still percent-encoded even though the values are not.
    it("encodes the space separator but not reserved characters in values", () => {
      expect(encodeSpaceDelimited("q", ["a:b", "c/d"], reserved)).toEqual(
        "q=a:b%20c/d",
      );
    });

    it("encodes the pipe separator but not reserved characters in values", () => {
      expect(encodePipeDelimited("q", ["a:b", "c/d"], reserved)).toEqual(
        "q=a:b%7Cc/d",
      );
    });
  });

  describe("mixed with ordinary parameters", () => {
    it("encodes an ordinary parameter that shares a value with a reserved one", () => {
      expect(encodeForm("q", "a:b", percent)).toEqual("q=a%3Ab");
      expect(encodeForm("q", "a:b", reserved)).toEqual("q=a:b");
    });
  });
});

describe("allowReserved path encoding", () => {
  const reserved = { charEncoding: "percentExceptReserved" } as const;
  const percent = { charEncoding: "percent" } as const;

  describe("simple style", () => {
    it("passes reserved characters through a scalar", () => {
      expect(encodeSimple("id", "a:b/c", reserved)).toEqual("a:b/c");
    });

    it("passes reserved characters through array items", () => {
      expect(encodeSimple("id", ["a:b", "c/d"], reserved)).toEqual("a:b,c/d");
    });

    // Every character encodeSimple emits is part of the path segment, so
    // member names follow the same rule as the values beside them.
    it("applies allowReserved to object member names", () => {
      expect(encodeSimple("id", { "k:1": "v/1" }, reserved)).toEqual("k:1,v/1");
    });

    it("still encodes object member names under plain percent encoding", () => {
      expect(encodeSimple("id", { "k:1": "v/1" }, percent)).toEqual(
        "k%3A1,v%2F1",
      );
    });

    it("encodes characters outside the reserved set in member names", () => {
      expect(encodeSimple("id", { "a b": "c d" }, reserved)).toEqual(
        "a%20b,c%20d",
      );
    });

    it("percent-encodes multi-byte UTF-8 in member names", () => {
      expect(encodeSimple("id", { é: "☃" }, reserved)).toEqual(
        "%C3%A9,%E2%98%83",
      );
    });
  });

  describe("matrix style", () => {
    it("applies allowReserved to object member names but not the parameter name", () => {
      expect(encodeMatrix("id", { "k:1": "v/1" }, reserved)).toEqual(
        ";id=k:1,v/1",
      );
    });
  });

  describe("label style", () => {
    it("applies allowReserved to object member names", () => {
      expect(encodeLabel("id", { "k:1": "v/1" }, reserved)).toEqual(".k:1.v/1");
    });
  });
});
