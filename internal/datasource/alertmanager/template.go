package alertmanager

const defaultTemplate = `
    <div class="collapsible">
		<h2 style="display: inline;">Alerts</h2>
		<div style="text-align: right;"><small>Last update: xxxsdfsadfasdfasd</small></div>
	</div>
    <table>
        <tr>
            <th>Name</th>
            <th>Severity</th>
            <th>Count</th>
        </tr>
        {{range .}}
        <tr>
            <td>{{.Name}}</td>
            <td>{{.Severity}}</td>
            <td>{{.Count}}</td>
        </tr>
        {{end}}
    </table>
    `
