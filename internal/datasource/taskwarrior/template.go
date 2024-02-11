package taskwarrior

const defaultTemplate = `
    <h2 class="collapsible">Taskwarrior</h2>
    <table>
        <tr>
			<th>Id</th>
            <th>Name</th>
            <th>Due</th>
            <th>Project</th>
            <th>Urgency</th>
        </tr>
        {{range .}}
        <tr>
			<td>{{ .Id }}</td>
            <td>{{ .Description }}</td>
            <td{{ if .DueCssClass }} class="{{ .DueCssClass }}"{{ end }}>{{ .Due }}</td>
            <td>{{ .Project }}</td>
            <td>{{ printf "%.1f" .Urgency }}</td>
        </tr>
        {{end}}
    </table>
    `
