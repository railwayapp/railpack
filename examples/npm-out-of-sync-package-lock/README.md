# Out-of-sync npm lockfile

This example intentionally keeps `package-lock.json` out of sync with
`package.json` to reproduce an `npm ci` failure. The lockfile was generated
with only `dayjs` installed; `is-number` was then added to `package.json`
without regenerating the lockfile.

Do not run `npm install` in this directory unless you are recreating the
out-of-sync fixture.
