# Slack RTM (Real Time Messaging) API Integration

[![GoDoc](https://godoc.org/github.com/james-bowman/slack?status.svg)](https://godoc.org/github.com/james-bowman/slack) 
[![Build Status](https://travis-ci.org/james-bowman/slack.svg?branch=master)](https://travis-ci.org/james-bowman/slack)

This package allows integration with the [Slack Real Time Messaging API](https://api.slack.com/rtm) for consuming and publishing messages in real time to/from [Slack](https://slack.com).

To make use of the API, you will need to setup a user account or a ['bot' (robot) user](https://api.slack.com/bot-users) (a special kind of user account specifically designed for programatic access to APIs) and obtain an [authentication token](https://api.slack.com/web#basics) from Slack.

Once you have a user account and an authentication token you can use the API to connect to slack.  Here is an example:

``` go
conn, err := slack.Connect(slackToken)
	
if err != nil {
	log.Fatal(err)
}

replier := func(msg *slack.Message) {
	msg.Respond(msg.Text)
}

slack.EventProcessor(conn, replier, nil)
```

The above snippet of code connects to the Slack RTM API over a web socket and starts listening for all messages directed specifically at the user account used to connect to slack (either a direct message or a message in a channel preceded by '@_username_:' ).  It will then echo the same message back to Slack.

To also process messages not directed specifically at the connected user, a similar function can be passed as the third parameter to the EventProcessor method (either in addition to or instead of the second parameter).

This package is used by [Talbot](http://github.com/james-bowman/talbot), a bot that is available to be used directly, extended or simply as an example.

## Features

Features implemented

- Processing Slack message events
- Option to respond 
    - just to directed messages (those sent as private messages or preceeded by '@_username_:' in open channels)
    - to all messages 
    - or to both directed and all messages independently.
- Sending messages to Slack
- Automatic reconnection following a lost connection
- Support for explicit web proxies (running on corporate LANs)
- Chunking of large messages into multiple smaller messages for sending to Slack
- Updating configuration based upon new member events, etc.

## To Do

Still outstanding...

- Reliable message sending i.e. checking for Ack's for sent messages (especially upon reconnection)
- Processing of Slack message changed events (currently ignored)
- Processing other Slack event types


