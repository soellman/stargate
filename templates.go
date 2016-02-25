package main

import (
	"html/template"
	"log"
)

func getTemplates() *template.Template {
	t, err := template.New("foo").Parse(`{{define "index.html"}}
<!DOCTYPE html>
<html lang="en" charset="utf-8">
<head>
	<title>Sign In</title>
	<meta name="viewport" content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no">
	<style>
	body {
		font-family: "Helvetica Neue",Helvetica,Arial,sans-serif;
		font-size: 14px;
		line-height: 1.42857143;
		color: #333;
		background: #f0f0f0;
	}
	.signin {
		display:block;
		margin:20px auto;
		max-width:400px;
		background: #fff;
		border:1px solid #ccc;
		border-radius: 10px;
		padding: 20px;
	}
	.center {
		text-align:center;
	}
	.btn {
		color: #fff;
		background-color: #428bca;
		border: 1px solid #357ebd;
		-webkit-border-radius: 4;
		-moz-border-radius: 4;
		border-radius: 4px;
		font-size: 14px;
		padding: 6px 12px;
	  	text-decoration: none;
		cursor: pointer;
	}

	.btn:hover {
		background-color: #3071a9;
		border-color: #285e8e;
		ext-decoration: none;
	}
	label {
		display: inline-block;
		max-width: 100%;
		margin-bottom: 5px;
		font-weight: 700;
	}
	input {
		display: block;
		width: 100%;
		height: 34px;
		padding: 6px 12px;
		font-size: 14px;
		line-height: 1.42857143;
		color: #555;
		background-color: #fff;
		background-image: none;
		border: 1px solid #ccc;
		border-radius: 4px;
		-webkit-box-shadow: inset 0 1px 1px rgba(0,0,0,.075);
		box-shadow: inset 0 1px 1px rgba(0,0,0,.075);
		-webkit-transition: border-color ease-in-out .15s,-webkit-box-shadow ease-in-out .15s;
		-o-transition: border-color ease-in-out .15s,box-shadow ease-in-out .15s;
		transition: border-color ease-in-out .15s,box-shadow ease-in-out .15s;
		margin:0;
		box-sizing: border-box;
	}
	footer {
		display:block;
		font-size:10px;
		color:#aaa;
		text-align:center;
		margin-bottom:10px;
	}
	footer a {
		display:inline-block;
		height:25px;
		line-height:25px;
		color:#aaa;
		text-decoration:underline;
	}
	footer a:hover {
		color:#aaa;
	}
	</style>
</head>
<body>
	{{ if .Message }}
	<div class="signin center">
	<p><label>{{.Message}}<label></p>
	</div>
	{{ end}}

	<div class="signin">
	<form method="POST" action="">
		<label for="password">enter token</label><input type="password" name="key" id="key" size="10" maxlength="30"><br/>
		<button type="submit" class="btn">sign in</button>
	</form>
	</div>
	<footer>
	</footer>
</body>
</html>
{{end}}`)

	if err != nil {
		log.Fatalf("failed parsing template %s", err)
	}
	return t
}
