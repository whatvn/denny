package denny

type HttpMethod string

const (
	HttpGet    HttpMethod = "GET"
	HttpPost   HttpMethod = "POST"
	HttpPatch  HttpMethod = "PATCH"
	HttpOption HttpMethod = "OPTION"
	HttpDelete HttpMethod = "DELETE"
)
