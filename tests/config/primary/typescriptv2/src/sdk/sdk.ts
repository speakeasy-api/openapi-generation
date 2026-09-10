// This incomplete file will trick the generator into apply the custom code
// regions to the actual final output version of this sdk file.

// #region imports
import { matchStatusCode } from "../lib/http.js"
// #endregion imports

// #region sdk-class-body
async customHealthCheck(): Promise<boolean> {
  const res = await this.health.check();
  return matchStatusCode(res.httpMeta.response, "2XX");
}
// #endregion sdk-class-body