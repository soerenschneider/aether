package stocks

const defaultTemplate = `<h2 class="collapsible">Stocks</h2>
		<table>
			<thead>
				<tr>
					<th>Symbol</th>
					{{ range .Timestamps }}
						<th>{{ .Format "02.01.06" }}</th>
					{{ end }}
				</tr>
			</thead>
			<tbody>
				{{ range .Symbols }}
				<tr>
					<td><a href="{{ .Link }}" target=”_blank”>{{ .Name }}</a></td>
					{{range .Values }}
						<td>{{ printf "%.2f" . }}</td>
					{{end}}
				</tr>
				{{ end }}
		</tbody>
		</table>
	`
