package pilot

import "fmt"

const (
	regexp_nginx_error     = "^(?<time>[\\d{4}\\/\\d{2}\\/\\d{2} \\d{2}:\\d{2}:\\d{2}]*) \\[(?<log_level>(.*?))\\] (?<pid>(\\d*?))#(?<tid>(\\d*?)): \\*(?<connection_number>(\\d*?)) (?<msg>(.*?))$"
	regexp_nginx_access    = "^(?<client_ip>[^ ]*) ([ -]*) \\[(?<time>[^\\]]*)\\] \"(?<method>[^ ]*) (?<uri>[^ ]*) (?<protocol>[^ ]*)\" (?<status>[^ ]*) (?<bytes_send>[^ ]*) [^ ]* \"(?<agent>.*)\"$"
	regexp_jvm_gc          = "^(?<time>\\d{4}-\\d{2}-\\d{2}T\\d{2}:\\d{2}:\\d{2}.\\d{3}[+|-]\\d{4}*): (?<jvm_time_offset>[^ ]+): \\[(?<gc_type>.*) \\((?<gc_reason>.*)\\) (?<data>.*), (?<cost>.*) secs\\] \\[Times: user=(?<cpu_user_cost_time>.*) sys=(?<cpu_sys_cost_time>.*), real=(?<cpu_real_cost_time>.*) (?<cost_time_unit>.*)\\]"
	regexp_tomcat_access   = "^(?<client_ip>[^ ]*) \\[(?<time>[^\\]]*)\\] (?<protocol>[^ ]*) (?<method>[^ ]*) (?<uri>[^ ]*) (?<status>[^ ]*) (?<bytes_send>[^ ]*) (?<cost>[^ ]*) (?<uid>[^ ]*)$"
	regexp_tomcat_catalina = "^(?<time>\\d{2}-[a-zA-Z]+-\\d{4} \\d{2}:\\d{2}:\\d{2}.\\d{3}*) (?<level>[^ ]*) \\[(?<thread>[^ ]*)\\] (?<method>[^ ]*)(?<message>.+)$"
)

// FormatConverter converts node info to map
type FormatConverter func(info *LogInfoNode) (map[string]string, error)

var converters = make(map[string]FormatConverter)

// Register format converter instance
func Register(format string, converter FormatConverter) {
	converters[format] = converter
}

// Convert convert node info to map
func Convert(info *LogInfoNode) (map[string]string, error) {
	converter := converters[info.value]
	if converter == nil {
		return nil, fmt.Errorf("unsupported log format: %s", info.value)
	}
	return converter(info)
}

// SimpleConverter simple format converter
type SimpleConverter struct {
	properties map[string]bool
}

func init() {

	simpleConverter := func(properties []string) FormatConverter {
		return func(info *LogInfoNode) (map[string]string, error) {
			validProperties := make(map[string]bool)
			for _, property := range properties {
				validProperties[property] = true
			}
			ret := make(map[string]string)
			for k, v := range info.children {
				if _, ok := validProperties[k]; !ok {
					return nil, fmt.Errorf("%s is not a valid properties for format %s", k, info.value)
				}
				ret[k] = v.value
			}
			return ret, nil
		}
	}

	Register("nonex", simpleConverter([]string{}))
	Register("csv", simpleConverter([]string{"time_key", "time_format", "keys"}))
	Register("json", simpleConverter([]string{"time_key", "time_format"}))
	Register("regexp", simpleConverter([]string{"time_key", "time_format"}))
	Register("apache2", simpleConverter([]string{}))
	Register("apache_error", simpleConverter([]string{}))
	Register("nginx", simpleConverter([]string{}))
	Register("regexp", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{"pattern", "time_format"})(info)
		if err != nil {
			return ret, err
		}
		if ret["pattern"] == "" {
			return nil, fmt.Errorf("regex pattern can not be empty")
		}
		return ret, nil
	})
	Register("nginx_error", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{})(info)
		if err != nil {
			return ret, err
		}
		info.value = "regexp"
		ret["pattern"] = regexp_nginx_error
		return ret, nil
	})
	Register("nginx_access", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{})(info)
		if err != nil {
			return ret, err
		}
		info.value = "regexp"
		ret["pattern"] = regexp_nginx_access
		ret["time_format"] = "%d/%b/%Y:%H:%M:%S %z"
		return ret, nil
	})
	Register("jvm_gc", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{})(info)
		if err != nil {
			return ret, err
		}
		info.value = "regexp"
		ret["pattern"] = regexp_jvm_gc
		return ret, nil
	})
	Register("tomcat_access", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{})(info)
		if err != nil {
			return ret, err
		}
		info.value = "regexp"
		ret["pattern"] = regexp_tomcat_access
		ret["time_format"] = "%d/%b/%Y:%H:%M:%S %z"
		return ret, nil
	})
	Register("tomcat_catalina", func(info *LogInfoNode) (map[string]string, error) {
		ret, err := simpleConverter([]string{})(info)
		if err != nil {
			return ret, err
		}
		info.value = "multiline"
		ret["format_firstline"] = fmt.Sprintf("/%s/", "\\d{2}-[a-zA-Z]+-\\d{4}")
		ret["format1"] = fmt.Sprintf("/%s/", regexp_tomcat_catalina)
		return ret, nil
	})
}
