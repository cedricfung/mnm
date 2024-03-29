package main

import (
	_ "embed"
	"encoding/base64"
	"net/http"
	"text/template"
)

//go:embed mnm.sh
var MNMSH []byte

//go:embed html/index.html
var TemplateIndex []byte

//go:embed html/token.html
var TemplateToken []byte

//go:embed html/fonts/SpaceMono-Bold.woff2
var SpaceMonoBold []byte

//go:embed html/fonts/SpaceMono-Regular.woff2
var SpaceMonoRegular []byte

//go:embed html/style.css
var Stylesheet []byte

//go:embed html/icons/android-chrome-192x192.png
var Icon []byte

func (hdr *Handler) renderHTML(w http.ResponseWriter, r *http.Request, tb []byte, data map[string]interface{}) {
	tpl, err := template.New("index").Parse(string(tb))
	if err != nil {
		panic(err)
	}

	data["Fonts"] = map[string]interface{}{
		"SpaceMonoRegular": base64.StdEncoding.EncodeToString(SpaceMonoRegular),
		"SpaceMonoBold":    base64.StdEncoding.EncodeToString(SpaceMonoBold),
	}
	data["Stylesheet"] = string(Stylesheet)
	data["Icon"] = base64.StdEncoding.EncodeToString(Icon)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tpl.Execute(w, data)
}
