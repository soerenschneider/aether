package astral

const defaultTemplate = `
    <h2 class="collapsible">Photo</h2>
    <table>
        <tr>
            <th>Name</th>
            <th>Date</th>
        </tr>
        <tr>
            <td>Blue Hour</td>
            <td>{{ .BlueHourRising.Start.Format "15:04:05" }} - {{ .BlueHourRising.End.Format "15:04:05" }}</td>
        </tr>
        <tr>
            <td>Blue Hour Setting</td>
            <td>{{ .BlueHourSetting.Start.Format "15:04:05" }} - {{ .BlueHourSetting.End.Format "15:04:05" }}</td>
        </tr>
        <tr>
            <td>Golden Hour</td>
            <td>{{ .GoldenHourRising.Start.Format "15:04:05" }} - {{ .GoldenHourRising.End.Format "15:04:05" }}</td>
        </tr>
        <tr>
            <td>Golden Hour Setting</td>
            <td>{{ .GoldenHourSetting.Start.Format "15:04:05" }} - {{ .GoldenHourSetting.End.Format "15:04:05" }}</td>
        </tr>
    </table>
    `
