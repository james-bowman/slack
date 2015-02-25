#Slack RTM (Real Time Messaging) API Integration

This package allows integration with the [Slack Real Time Messaging API](https://api.slack.com/rtm) for consuming and publishing messages in real time to/from [Slack](https://slack.com).

To make use of the API, you will need to setup a user account or a ['bot' user](https://api.slack.com/bot-users) (a special kind of user account specifically designed for programatic access to APIs) and obtain an [authentication token](https://api.slack.com/web#basics) from Slack.

Once you have a user account and an authentication token you can use the API to connect to slack.  Here is an example:

``` go
conn, err := slack.Connect(slackToken)
	
if err != nil {
	log.Fatal(err)
}

slack.EventProcessor(conn, func(msg *slack.Message) {
	msg.Respond(message.Text)
})
```

The above snippet of code connects to the Slack RTM API over a web socket and starts listening for messages directed at the user account used to connect to slack (either a direct message or a message in a channel preceded by @<username>: ).  Tt will then repeat the same message back to Slack.

