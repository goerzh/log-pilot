{{range .configList}}
{{if eq .Format "regexp"}}
[PARSER]
    Name        polit
    Format      regex
    Regex       {{ index .FormatConfig "pattern" }}
    Time_Key    time
    Time_Format %Y-%m-%dT%H:%M:%S.%L%z
{{end}}
{{end}}
