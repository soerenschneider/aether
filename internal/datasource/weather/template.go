package weather

const defaultTemplate = `<h2 class="collapsible">Weather {{.City.Name }} <small>☀️{{ .City.SunriseTime.Format "15:04" }} 🌚{{ .City.SunsetTime.Format "15:04" }}</small></h2>
	<a href="https://openweathermap.org/city/{{ .City.ID }}" target=”_blank”>
		<table>
			<thead>
				<tr>
					<th>Time</th>
					<th>Temp (°C)</th>
					<th>Pop (%)</th>
					<th>Rain 3h(mm)</th>
					<th>Hum (%)</th>
					<th>Wind Speed (m/s)</th>
					<th>Clouds (%)</th>
					<th>Vis. (%)</th>
					<th>Desc</th>
				</tr>
			</thead>
			<tbody>
			{{range $i, $val := .List }}
				{{if lt $i 10}}
					<tr>
						<td>{{.Time.Format "15:00" }} {{ .WeatherEmoji }}</td>
						<td{{ if .Main.TempClass }} class="{{ .Main.TempClass }}"{{ end }}>{{ printf "%.1f" .Main.Temp }} ({{ printf "%.1f" .Main.FeelsLike }})</td>
						<td title="Shiaat"{{ if .PopClass }} class="{{ .PopClass }}"{{ end }}>{{ .Pop }}</td>
						<td{{ if .Rain.CssClass }} class="{{ .Rain.CssClass }}"{{ end }}>{{ .Rain.H3 }}</td>
						<td{{ if .Main.HumidityClass }} class="{{ .Main.HumidityClass }}"{{ end }}>{{.Main.Humidity }}</td>
						<td{{ if .Wind.CssClass }} class="{{ .Wind.CssClass }}"{{ end }}>{{ .Wind.Speed }}</td>
						<td{{ if .Clouds.CssClass }} class="{{ .Clouds.CssClass }}"{{ end }}>{{ .Clouds.All }}</td>
						<td{{ if .VisibilityCssClass }} class="{{ .VisibilityCssClass }}"{{ end }}>{{ .VisibilityPercent }}</td>
						<td>{{ .WeatherDescription }}<img style="width: 25px; height: 25px" src="http://openweathermap.org/img/wn/{{ .WeatherIconName }}.png"/></td>
					</tr>
				{{end}}
			{{end}}
		</tbody>
		</table>
	</a>`
