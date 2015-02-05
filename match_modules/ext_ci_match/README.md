# Case Insensitive Matching
You can match against strings using case insensitive matching:

```json
"foo" : {
  "__match__" : "case-insensitive",
  "value" : "FooBaR"
}
```
The above will be a positive match for any message containing the attribute "foo" with a string value of "foobar", "fOObAr" and so on.
