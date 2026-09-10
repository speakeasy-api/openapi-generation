import { BeforeRequestContext, BeforeRequestHook } from "./types.js";

import { SchemeCustomHTTPSecurity } from "../models/security.js";

export class CustomHttpHook implements BeforeRequestHook {
  beforeRequest(hookCtx: BeforeRequestContext, request: Request): Request {
    switch (hookCtx.operationID) {
      case "customHttpOnly": {
        let sec = hookCtx.securitySource;
        if (typeof sec === "function") {
          sec = sec();
        }

        if (!sec) {
          throw new Error("security source is not defined");
        }

        const customHttp = sec as SchemeCustomHTTPSecurity;
        if (!customHttp) {
          throw new Error("customHttp security scheme is not defined");
        }

        request.headers.set("X-Security-UserID", String(customHttp.userID));
        request.headers.set("X-Security-Role", customHttp.role);
        request.headers.set("X-Security-Passphrase", customHttp.passphrase);
        request.headers.set(
          "X-Security-AccessCode",
          String(customHttp.accessCode),
        );
        if (customHttp.scopes) {
          request.headers.set(
            "X-Security-Scopes",
            JSON.stringify(customHttp.scopes),
          );
        }

        break;
      }
    }

    return request;
  }
}
