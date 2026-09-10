// This incomplete file will trick the generator into apply the custom code
// regions to the actual final output version of this sdk file.

// #region imports
import java.util.Collection;
import java.util.List;
import java.util.stream.Collectors;
// #endregion imports

    // #region class-body
    /**
     * Convenience method to get all deeply nested attributes from the optional
     * list.
     *
     * @return A list of all deeply nested attributes from the optional list.
     */
    public List<String> deeplyNestedAttribute() {
        return optionalList()
                .stream()
                .flatMap(Collection::stream)
                .flatMap(ol -> ol.optionalItem().stream())
                .flatMap(oi -> oi.deeplyNestedAttribute().stream())
                .collect(Collectors.toList());
    }
    // #endregion class-body
