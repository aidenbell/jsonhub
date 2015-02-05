# jsonhub - A Pattern Matching Pub/Sub
A publish/subscribe message queue with JSON pattern matching based subscriptions.

Any JSON value is a valid message and clients can subscribe to an exchange using a description of the JSON messages they wish to match against. The client will then obtain messages that match the pattern. Subscription patterns are themselves JSON and can contain "match values" to specify more complex matching such as greater-than and less-than matches, geospatial matches and case insensitive matches.

## Install
As jsonhub is a Go application, downloading and building it is pretty easy if you have the Go tools installed and setup. Just run

```sh
go get github.com/aidenbell/jsonhub/jsonhubd
```

and then

```sh
]$ cd $GOPATH/src/github.com/aidenbell/jsonhub/jsonhubd
]$ $GOPATH/bin/jsonhubd
2015/02/04 23:21:40 JSON Message Queue
2015/02/04 23:21:40 Running server on port 9977
```
You can then find a Javascript test client at http://localhost:9977/client/ , POSTing JSON and listening for messages happens on http://localhost:9977/ .

## Example Messages and Subscriptions

### Example Published Message
Imagine an RFID office entry system for a shared office building. Every time a person uses their entry card, a message detailing the access is sent to the message queue with the details of the person and the area of the building their card gives them access to.
```json
{
  "name" : "Dave",
  "age" : 22,
  "occupation": "Programmer",
  "hardware" : [ "monitor", "workstation", "laptop", "chair" ],
  "employer" : {
    "name": "Big Co Inc",
    "address1" : "123 BigCo Lane",
    "address2" : "Businesston",
    "postcode" : "ABC XYZ"
  },
  "access_area": {
    "type": "Polygon",
    "coordinates": [
      [ [100.0, 0.0], [101.0, 0.0], [101.0, 1.0], [100.0, 1.0], [100.0, 0.0] ]
    ]
  },
  "door": "stair-22",
  "granted" : true
}
```

### Sub 1: Value Comparison
An client could subscribe using the following patterns and obtain the above message if it was published to the exchange:
```json
{
  "age" : 22,
  "occupation" : "Programmer"
}
```

### Sub 2: Using Matcher Values
This would provide all messages where the occupation is "Programmer" and the individual is 22 years old. We can match a bit more loosely:
```json
{
  "age" : {
      "__match__" : "greater-than",
      "value" : 22
    },
  "occupation" : {
    "__match__" : "case-insensitive",
    "value" : "programmer"
  }
}
```

## Server Sent Events
Currently matching messages are polled via [Server Sent Events](http://dev.w3.org/html5/eventsource/). Websocket support and other protocols are planned.

## Future
Future and needed features include

  * Websocket support
  * Multiple named exchanges
  * Configurable message distribution for equivalent patterns (round-robin, random)
  * Negations in patterns
  * Various matchers


## Disclaimer
This is a little side project I use to learn Go. The source probably isn't very nice or idiomatic, but as I learn I refactor and improve it. It is also missing some important features like negation, tests, documentation etc. Pull requests welcome.
