// This incomplete file will trick the generator into apply the custom code
// regions to the actual final output version of this sdk file.

// #region imports
import static org.openapis.openapi.utils.Utils.statusCodeMatches;
// #endregion imports

    // #region sdk-class-body
    public boolean customHealthCheck() throws Exception {
        CheckResponse checkResponse = checkDirect();
        return statusCodeMatches(checkResponse.statusCode(), "2XX");
    }
    // #endregion sdk-class-body
