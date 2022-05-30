FROM ubuntu

LABEL sh.recode.vscode.extensions="golang.go"

# Add Go to path
ENV PATH=$PATH:/usr/local/go/bin:$HOME/go/bin

FROM base

LABEL "sh.recode.vscode.extensions"="dbaeumer.vscode-eslint"

# Set timezone
ENV TZ=America/Los_Angeles