package carddav

const defaultTemplate = `
    {{ if .Cards }}
    <h2 class="collapsible">Anniversaries {{ .From.Format "02.01.06" }} – {{ .To.Format "02.01.06" }}</h2>
    <table>
        <tr>
            <th>Name</th>
            <th>Date</th>
            <th>Type</th>
        </tr>
        {{ range .Cards }}
        <tr>
            <td>{{ .Name }}</td>
            <td>{{ .DateFormatted }} ({{ .Years }})</td>
            <td>{{ .Type }}</td>
        </tr>
        {{ end }}
    </table>
	{{ end }}
    `
