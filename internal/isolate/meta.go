package isolate

import "strings"

func ParseMeta(meta string) map[string]string {
	splitted := strings.Split(meta, "\n")

	res := map[string]string{}

	for _, s := range splitted {
		entry := strings.Split(s, ":")
		if len(entry) == 2 {
			res[entry[0]] = entry[1]
		}
	}

	switch res["status"] {
	case "":
		res["status"] = "Accepted"
	case "RE":
		res["status"] = "Runtime Error"
	case "TO":
		res["status"] = "Time Limit Exceeded"
	default:
		res["status"] = "Internal Error"
	}

	return res
}
