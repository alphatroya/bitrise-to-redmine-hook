# Bitrise to Redmine webhook bindings

[![Go Report Card](https://goreportcard.com/badge/github.com/alphatroya/ci-redmine-bindings)](https://goreportcard.com/report/github.com/alphatroya/ci-redmine-bindings)

This hook automatically move all issues marked as "Ready to build" to "Done" section when
internal build was completed. Number of the build will be printed in the specified custom
field.

## Installation

After deploying you need to specify the following ENV items:

- `REDMINE_API_KEY`: Access token for a Redmine bot user
- `REDMINE_HOST`: Redmine installation host address
- `STAMP_BUILD_CUSTOM_FIELD`: ID of the build number issue custom field
- `STAMP_DONE_STATUS`: ID of the done status
- `STAMP_READY_TO_BUILD_STATUS`: ID of the "Ready to the build" status

For optional Mailgun integration you should set also next items:

- `MAILGUN_API`: API key for Mailgun service
- `MAILGUN_DOMAIN`: domain address
- `MAILGUN_RECIPIENT`: a recipient for emails
- `MAILGUN_SENDER`: a sender for emails

## Bitrise configuration

- Add a new Outgoing Webhooks in the Bitrise Code tab.
- Specify <your-host-address>/bitrise as an URL
- Set "REDMINE_PROJECT" header with Redmine project id
