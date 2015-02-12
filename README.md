#Slack RTM (Real Time Messaging) API Integration

This package allows integration with the [Slack Real Time Messaging API](https://api.slack.com/rtm) for consuming and publishing messages in real time to/from [Slack](https://slack.com).

To make use of the API, you will need to setup a user account or a ['bot' user](https://api.slack.com/bot-users) (a special kind of user account specifically designed for programatic access to APIs) and obtain an [authentication token](https://api.slack.com/web#basics) from Slack.

Once you have a user account and an authentication token you can use the API to connect to slack.  Here is an example:

'''
conn, err := slack.Connect(slackToken)
	
if err != nil {
	log.Fatal(err)
}
		
conn.Start(func(msg slack.Event) *slack.Event {
	if msg.Type == "message" {
		return &slack.Event{Type: "message", Channel: msg.Channel, Text: msg.Text}
	} else {
		return nil
	} 
})
'''

The above snippet of code connects to the Slack RTM API over a web socket and starts listening for all events.  If the event is a message, it will repeat the same message back to Slack.

