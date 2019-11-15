{{range .configList}}
[INPUT]
    Name              tail
    Buffer_Chunk_Size 32K
    Buffer_Max_Size   32K
    Mem_Buf_Limit     10m
    Path              {{ .HostDir }}/{{ .File }}
    Path_Key          log_path
    Exclude_Path      *.gz,*.zip,*.db
    {{if .Stdout}}
    Parser            json
    {{end}}
    {{if eq .Format "json"}}
    Parser            json
    {{end}}
    {{if eq .Format "regexp"}}
    Multiline         On
    Parser_Firstline  polit
    {{end}}
    Tag               {{ $.containerId }}
    DB                /fluent-bit/db/fluent-bit.db
    DB.Sync           Off

[FILTER]
    Name record_modifier
    Match {{ $.containerId }}
    {{range $key, $value := .Tags}}
    Record {{ $key }} {{ $value }}
    {{end}}

[OUTPUT]
{{if eq $.output "elasticsearch"}}
    Name        es
{{end}}
{{if eq $.output "kafka"}}
    Name        kafka
    Brokers     {{ $.endpoints }}
    Topics      {{ index .Tags "topic" }}
    rdkafka.log.connection.close false
    rdkafka.request.required.acks 1
{{end}}
    Match       {{ $.containerId }}
    Timestamp_Key timestamp

{{end}}
