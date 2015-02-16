package exchange

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"

	// Some built-in matchers
	"github.com/aidenbell/jsonhub/match_modules/ext_ci_match"
	"github.com/aidenbell/jsonhub/match_modules/ext_geojson"
)

// ClientPool A basic queue that accepts a list of clients and matches messages against a
// match specification. The queue has various configuration options.
// TODO make a Queue a type of exchange and have it implement that to allow
// nested layouts
type ClientPool struct {
	*Exchange // ClientPool is an extended exchange

	Parent    Publisher
	matchSpec map[string]interface{}
	pingOnly  bool // You get empty messages, not the src message. Useful for counting
}

// NewClientPool is the constructor for ClientPool instances
func NewClientPool(parent Publisher, spec string) (*ClientPool, error) {
	// Parse the spec JSON
	parsedSpec := map[string]interface{}{}
	err := json.Unmarshal([]byte(spec), &parsedSpec)
	if err != nil {
		return nil, err
	}

	// Create a queue for the client
	q := ClientPool{
		NewExchange(),
		parent,
		parsedSpec,
		false}
	return &q, nil
}

// Receive a message from somewhere. This is the outcome of a subscription
// and allows the ClientPool to meet the Subscriber interface
func (q *ClientPool) Receive(m Messager) {
	log.Printf("Client Pool Receive()")
	go func() {
		if q.messageMatches(m) {
			log.Printf("Message matches!")
			q.Publish(m)
		}
	}()
}

// SetPingOnly allows switching the client pool between full message delivery
// and empty message. Useful if you are implementing a counter a group of
// clients that don't require a full message to do their job.
func (q *ClientPool) SetPingOnly(p bool) {
	q.pingOnly = p
}

// PingOnly is the getter for the current PingOnly state
func (q *ClientPool) PingOnly() bool {
	return q.pingOnly
}

// Test if a message matches the specification of the queue
// returning True if the message can be added and false if not
func (q *ClientPool) messageMatches(m Messager) bool {
	// parse the message JSON
	parsed := map[string]interface{}{}
	err := json.Unmarshal([]byte(m.Raw()), &parsed)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", m.Raw(), err)
		return false
	}

	return matchObject(parsed, q.matchSpec)
}

// A general function for matching an input parsed JSON message
// and a parsed subscription specification
func matchObject(msg map[string]interface{}, spec map[string]interface{}) bool {
	// TODO: Change the terminology to message and subscription so it isn't
	// so confusing to read.

	matches := true // Does this message attribute match the spec?

	// Loop through each attribute on the subscription. This is important
	// and allows empty subscriptions to match all.
	for k, sv := range spec {
		v := msg[k]

		// If there is an attribute in the spec that isn't in
		// the message, that counts as a fail. The spec definition implies "attribute exists"
		if v == nil {
			return false
		}

		svtype := reflect.TypeOf(sv)
		vtype := reflect.TypeOf(v)

		// Switch on spec value type, defaulting to basic Go comparison
		switch svcast := sv.(type) {
		// Spec value is an object, either extension
		// or plain object comparison
		case map[string]interface{}:
			if ext, ok := svcast["__match__"]; ok {
				// Spec is an extension, so delegate to that
				// extension for match result
				extstr := ext.(string)
				switch extstr {
				case "case-insensitive":
					matches = ext_ci_match.ExtCaseInsensitiveMatch(v, svcast)
				case "geojson-within":
					matches = ext_geojson.ExtGeoJSONWithin(v, svcast)
				default:
					return false
				}
			} else {
				// Spec is a plain object, ensure type match
				// and compare objects
				if svtype != vtype {
					return false
				}
				vcast := sv.(map[string]interface{})
				matches = matchObject(vcast, svcast)
			}
			break

		// Spec val isn't an object, just do type comparison
		// and value comparison for matching types.
		default:
			// Check that the types match firstly
			if svtype != vtype {
				return false
			}
			// Types match, compare values
			matches = (v == sv)
		}

		if matches == false {
			break // Don't keep processing message attrs if we get a false
		}
	}

	return matches
}
