package exchange

/*
 * Defines what messages look like regardless of the syntax or format
 * of the message itself. At the base we have a string. Higher levels
 * of message have methods for dealing with specifics. Queues and Exchanges
 * don't care about that. Matchers might.
 */
type Messager interface {
	Raw() string
}

/*
 A JsonMessage that validates the message is really JSON
 */
type JsonMessage struct {
	raw string
}

// Create a new message containing a JSON payload
func NewJsonMessage(json string) JsonMessage {
	m := JsonMessage{}
	m.SetMessage(json)
	return m
}

// Get a raw string of the message payload
func (m JsonMessage) Raw() string {

	return m.raw
}

// Set the message payload JSON
func (m *JsonMessage) SetMessage(s string) error {
	m.raw = s

	return nil
}
