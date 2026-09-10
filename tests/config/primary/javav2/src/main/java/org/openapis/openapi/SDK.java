// This incomplete file will trick the generator into apply the custom code
// regions to the actual final output version of this sdk file.

// #region imports
import java.util.List;
import org.openapis.openapi.models.operations.CheckResponse;
import org.openapis.openapi.models.shared.UnwieldyComponent;
import static org.openapis.openapi.utils.Utils.statusCodeMatches;
// #endregion imports

    // #region sdk-class-body
    public boolean customHealthCheck() throws Exception {
        CheckResponse checkResponse = this.health().checkDirect();
        return statusCodeMatches(checkResponse.statusCode(), "2XX");
    }
    // #endregion sdk-class-body

    // #region class-body
    public List<String> unwieldyGetFlattened() throws Exception {
        UnwieldyGetResponse response = unwieldyGetDirect();
        return response.unwieldyComponent()
                .map(UnwieldyComponent::deeplyNestedAttribute)
                .orElse(java.util.Collections.emptyList());
    }
    // #endregion class-body
