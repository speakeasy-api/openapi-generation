<!-- Start SDK Example Usage [usage] -->
```bash
cli post-file --upload ../.speakeasy/testfiles/example.file

```

```bash
cli tag1 post-file-with-encoding --file ../.speakeasy/testfiles/example.file

```

```bash
cli test-group tag2 post-test --deprecated-query-param2 'some example query param' --obj '{"str":"example","bool":true,"int":999999,"int32":1,"num":1.1,"float32":2940.96,"enumProp":"First","date":"2020-01-01","dateTime":"2020-01-01T00:00:00Z","anything":"<value>","boolOpt":true,"intOptNull":999999,"numOptNull":1.1,"intEnum":3,"int32Enum":69,"bigint":702830,"bigintStr":"12345678901234567890","decimal":3.141592653589,"decimalStr":"3858.6","obj":{"str":"example"},"map":{"key":{"str":"example"}},"arr":[{"str":"example"}],"any":"<value>","type":"0","nullableIntEnum":3,"nullableStringEnum":"Second","color":"green","icon":"tick","heroWidth":480}' --type type1

```

### A custom readme heading

A custom usage description

```bash
cli tag1 list-test1 --page 100 --query-param1 'some example query param' --query-param2 1 --header-param1 'some example header param'

```
<!-- End SDK Example Usage [usage] -->