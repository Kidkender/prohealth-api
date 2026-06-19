package template

import "embed"

//go:embed mail/*.html
var MailFS embed.FS
