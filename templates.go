package healthserver

import (
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/css"
	"html/template"
)

const reportTemplateString = `
<html>
    <head>
        <meta charset="utf-8">
        <title>Health Status</title>
        <style>
            table {
                border-collapse: collapse;
            }
            tr {
                height: 2em;
            }
            td {
                padding-left: 0.7em;
                padding-right: 0.7em;
            }
            .status {
                text-align: center;
            }
            .failing {
                background-color: red;
            }
            .passing {
                background-color: lawngreen;
            }
        </style>
    </head>
    <body>
        <table>
            {{range .}}
                <tr class="{{if not .Err}}passing{{else}}failing{{end}}">
                    <td class="status">{{if not .Err}}&#x2714{{else}}&#x2718{{end}}</td>
                    <td>{{.Name}}</td>
                </tr>
            {{end}}
        </table>
    </body>
</html>
`

var reportTemplate *template.Template = nil

func getReportTemplate() *template.Template {
	if reportTemplate == nil {
		// Lazy load template.
		minifier := minify.New()
		minifier.AddFunc("text/html", html.Minify)
		minifier.AddFunc("text/css", css.Minify)

		minifiedTemplateString, err := minifier.String("text/html", reportTemplateString)
		if err != nil {
			panic(err)
		}

		reportTemplate = template.Must(
			template.
				New("healthserver-report-template").
				Parse(minifiedTemplateString),
		)
	}

	return reportTemplate
}
