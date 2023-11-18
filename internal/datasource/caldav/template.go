package caldav

const defaultTemplate = `
    {{if .Entries }}
		<h2 class="collapsible">Agenda {{ .From.Format "02.01.06" }} – {{ .To.Format "02.01.06" }}</h2>
		<table>
			<tr>
				<th>Summary</th>
				<th>Date</th>
				<th>Location</th>
			</tr>
			{{ range .Entries }}
			<tr>
				<td>{{ .Summary }}</td>
				<td>{{ .Formatted }}</td>
				<td>{{ if .LocationUrl }}<a href="{{ .LocationUrl }}" target=”_blank”>{{ .Location }}</a>{{ else }}{{ .Location }}{{ end }}</td>
			</tr>
			{{end}}
		</table>
	{{end }}
    `
