package submission

type LangaugeHandler func(*Submission, chan SubmissionResult)

var LanguageHandlers = map[string]LangaugeHandler{
	"python3": python3Handler,
}
