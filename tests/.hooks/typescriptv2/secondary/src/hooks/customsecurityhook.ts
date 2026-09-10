import { BeforeRequestContext, BeforeRequestHook } from "./types.js";

import { CustomSchemeAppIdSecurity } from "../sdk/models/operations/customschemeappid.js";

export class CustomSecurityHook implements BeforeRequestHook {
  beforeRequest(hookCtx: BeforeRequestContext, request: Request): Request {
    switch (hookCtx.operationID) {
      case "customSchemeAppId": {
        let sec = hookCtx.securitySource;
        if (typeof sec === "function") {
          sec = sec();
        }
        if (!sec) {
          throw new Error("security source is not defined");
        }

        const customSec = sec as CustomSchemeAppIdSecurity;
        request.headers.set("X-Security-App-Id", customSec.appId);
        request.headers.set("X-Security-Secret", customSec.secret);

        break;
      }
    }

    return request;
  }
}
