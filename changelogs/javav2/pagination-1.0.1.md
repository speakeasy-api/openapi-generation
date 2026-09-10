## pagination: 1.0.1 - 2026-03-09
### :wrench: Refactors
- unify Paginator/AsyncPaginator API to use `BiFunction<ReqT, ProgressParamT, ResponseT>` page fetcher, eliminating separate request modifier and data fetcher arguments
