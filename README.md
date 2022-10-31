# Bitrise to Redmine webhook

[![Go Report Card](https://goreportcard.com/badge/github.com/alphatroya/ci-redmine-bindings)](https://goreportcard.com/report/github.com/alphatroya/ci-redmine-bindings)

This app automatically move all issues marked as "Ready to build" to "Done" section after internal build completed. A build number printed in the specified custom field.

## Installation
Following ENV items are required to run :

- `REDMINE_API_KEY`: access token for a Redmine bot user
- `REDMINE_HOST`: Redmine installation host address
- `STAMP_BUILD_CUSTOM_FIELD`: Redmine ID of a build number issue custom field
- `STAMP_DONE_STATUS`: Redmine ID of a done status
- `STAMP_READY_TO_BUILD_STATUS`: Redmine ID of a "Ready to the build" status

For Mailgun integration you should add following items:

- `MAILGUN_API`: API key for Mailgun service
- `MAILGUN_DOMAIN`: domain address
- `MAILGUN_RECIPIENT`: a recipient for emails
- `MAILGUN_SENDER`: a sender for emails

## Bitrise configuration

- Add a new Outgoing Webhooks in the Bitrise Code tab.
- Specify <your-host-address>/bitrise/v2 as an URL
- Set "REDMINE_PROJECT" header with Redmine project id
