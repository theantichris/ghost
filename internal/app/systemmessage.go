package app

// systemMessage holds a system message.
type systemMessage string

// String returns the systemMessage as a string.
func (sysMsg systemMessage) String() string {
	return string(sysMsg)
}

// System messages.
const (
	msgClientResponse    systemMessage = "(system) I couldn't reach the LLM. Check your network or make sure the host is running then try again"
	msgNon2xxResponse    systemMessage = "(system) The LLM response with an error. Verify the model is pulled and the server is healthy before retrying."
	msgResponseBody      systemMessage = "(system) I couldn't read the LLM's reply. This might be a transient issue, please try again in a moment."
	msgUnmarshalResponse systemMessage = "(system) The LLM sent back something I couldn't parse. It may be busy, try your request again shortly."
)
