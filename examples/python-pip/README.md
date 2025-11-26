# Python pip Example

This example tests Python projects using pip with `requirements.txt`.

## `.python-version` Test Coverage

This example includes a `.python-version` file set to **3.11.14**, which is intentionally different from Railpack's default Python version (**3.13**). This ensures that:

1. Idiomatic version files (`.python-version`) are properly detected during analysis
2. The specified version is used instead of the default
3. Integration tests catch regressions in version file detection

If `.python-version` were ignored, the build would use Python 3.13 and the test assertion would fail.
