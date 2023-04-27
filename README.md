# tackle2-addon-analyzer

[![Analyser Addon Repository on Quay](https://quay.io/repository/konveyor/tackle2-addon-analyzer/status "Analyser Addon Repository on Quay")](https://quay.io/repository/konveyor/tackle2-addon-analyzer) [![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html) [![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/konveyor/tackle2-addon-analyzer/pulls) [![Test Analyser Addon](https://github.com/konveyor/tackle2-addon-analyzer/actions/workflows/test-analyzer.yml/badge.svg?branch=main)](https://github.com/konveyor/tackle2-addon-analyzer/actions/workflows/test-analyzer.yml)

Tackle (2nd generation) addon for Analyser.


Task data.

*=_optional_

```
{
  output: string,
  mode: {
    binary: bool,
    withDeps: bool,
    artifact: string,
  },
  sources: [str,],
  targets: [str,],
  scope: {
    withKnown: bool,
    packages: {
      included: [str,],
      excluded: [str,]
    }
  },
  rules*: {
    path: str, 
    tags: {
      included: [str,],
      excluded: [str,]
    }
  }
}
```


## Code of Conduct
Refer to Konveyor's Code of Conduct [here](https://github.com/konveyor/community/blob/main/CODE_OF_CONDUCT.md).
