Varanus feature list:
- read configuration from a YAML or JSON file
- schedule monitor tasks at various intervals
- execute scheduled monitor tasks

Architecture

- generic monitoring event structure
  - can run -> returns true (do the event) or false (requeue the event)
  - test event -> execute the setup or test 



Configuration details

  - notifier modules
    - sms - configuration for SMS gateway
    - email - smtp credentials for sending notifications
  - notification groups
    - group name
    - list of notifications
      - notification type (SMS, email)
      - notification config (i.e destination number or email)
  - module configs
    - smtp
      - absolute maximum sending rate for a group of smtp servers
        - maybe the grouping is automatic by resolving the server IPs?
    - sms
      - absolute maximum sending rate for sms messages
    - imap
      - maximum polling frequency
  - report generation
    - report generation frequency
    - notifier group
  - monitor types
    - email e2e -> send a message from an smpt server and verify that it shows up at an imap server
      - config:
        - target email address
        - smpt sender config
          - server address
          - server port
          - server ssl
          - username & password
        - imap receiver config
          - server address
          - server port
          - server ssl
          - username & password
        - sending rate
    - dns monitor - check that DNS entry exists

  