package main

var prefix = `
<html>
<head>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8"> 
	<meta http-equiv="refresh" content="60">
	<link rel="icon" href="data:image/svg+xml,<svg xmlns=%22http://www.w3.org/2000/svg%22 viewBox=%220 0 100 100%22><text y=%22.9em%22 font-size=%2290%22>📌</text></svg>">
	<style>
		body {
			font-family: Arial, sans-serif;
		}

		@media (prefers-color-scheme: light) {
			body {
				background-color: white;
				color: black;
			}
		}

		@media (prefers-color-scheme: dark) {
			body {
				background-color: black;
				color: white;
			}
		}
		
		.collapsible {
			background-color: #777;
			color: white;
			cursor: pointer;
			padding: 10px;
			width: 100%;
			border: none;
			text-align: left;
			outline: none;
		}
		
		.red {
			background-color: #FFA0A0; /* Pastel red */
			color: white;
		}
		
		.green {
			background-color: #A0D0C0; /* Pastel green */
			color: white;
		}
		
		.yellow {
			background-color: #FFE0A0; /* Pastel yellow */
			color: black;
		}
		
		.orange {
			background-color: #FFC0A0; /* Pastel orange */
			color: white;
		}
		
		.blue {
			background-color: #A0C0E0; /* Pastel blue */
			color: white;
		}
		
		.lightblue {
			background-color: #B3E0F2; /* Lighter pastel blue for cold */
			color: white;
		}
		
		.active, .collapsible:hover {
			background-color: #555;
		}
		
		table {
			border-collapse: collapse;
			width: 90%;
			margin: 20px auto;
			box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
		}
		
		th, td {
			border: 1px solid #dddddd;
			text-align: left;
			padding: 8px;
		}
		
		thead {
			background-color: #f2f2f2;
		}
	</style>
</head>
<body>
`

var postfix = `
<script>
var coll = document.getElementsByClassName("collapsible");
var i;

for (i = 0; i < coll.length; i++) {
  coll[i].addEventListener("click", function() {
    this.classList.toggle("active");
    var content = this.nextElementSibling;
    if (content.style.display === "none") {
      content.removeAttribute("style")
    } else {
      content.style.display = "none";
    }
  });
}
</script>
</body>
`
