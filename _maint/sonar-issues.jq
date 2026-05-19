#!/usr/bin/env jq -f

# Usage:
#  curl -K .sonar.curlrc.txt 'https://sonarcloud.io/api/issues/search?componentKeys=dolmen-go_sqlfunc&resolved=false&branch='"$(git symbolic-ref --short HEAD)" | jq -f sonar-issues.jq

.issues[]
| {severity,type,rule,component,line,message}
| {severity, type, rule, file: .component | .[1+index(":"):], line, message}
# | ( .severity + " " + .type " " + .rule + " " + .component + " " + .message )
