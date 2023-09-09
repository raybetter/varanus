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

TODOs:

- ...

Design decisions:

- Configuration management
  - decided to use regular YAML parsing for config loading
    - viper, the main config library, doesn't have good support for deeply nested defaults, which
      defeats a lot of its purpose.  The main advantage would have been to parcel out config loading
      of each module to itself, but we can get most of this from the validation logic which we have
      to do anyway
    - the `gopkg.in/yaml.v3` library provides YAML parsing for the config
  - decided against using `github.com/creasty/defaults` setting defaults (e.g. port defaults) -- requiring explicit config for everything is probably better anyway.

References:

- crypto
  - https://www.developer.com/languages/cryptography-in-go/
