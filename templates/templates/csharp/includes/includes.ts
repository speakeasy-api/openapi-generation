// NOTE: Ensure any changes here are reflected in the target tsconfig.json.

require("common/includes.ts");

require("readme/sections/common.ts");
require("readme/sections/errors.ts");
require("readme/sections/http-client.ts");
require("readme/sections/pagination.ts");
require("readme/sections/server.ts");
require("readme/sections/security.ts");
require("readme/sections/retries.ts");
require("readme/sections/eventstream.ts");

require("./sanitization.ts");
require("./models.ts");
require("./security.ts");
require("./templating.ts");
require("./namespaces.ts");
require("./nuget.ts");
require("../config.ts");
require("./imports.ts");
require("./documentation.ts");
require("./annotations.ts");
require("./errors.ts");
require("./modelusage.ts");
require("./serverusage.ts");
require("./enums.ts");
require("./open-unions.ts");
require("./utils.ts");
require("./compat.ts");
require("./pagination.ts");
require("./responses.ts");
require("./dependencies.ts");
require("./exclusions.ts");
require("./unimplemented.ts");
require("./auxiliary.ts");
require("./tests.ts");

require("hooks/hooks.ts");
