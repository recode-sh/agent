FROM recodesh/base-dev-env:latest

LABEL "com.example.vendor"="ACME Incorporated"
LABEL com.example.label-with-value="foo"
LABEL version="1.0"
LABEL sh.recode.vscode.extensions="golang.go,dbaeumer.vscode-eslint"
LABEL description="This text illustrates \
that label-values can span multiple lines."

# Set timezone
ENV TZ=America/Los_Angeles